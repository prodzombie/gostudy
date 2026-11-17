package cardstore

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prodzombie/gostudy/internal/schedule"
)

// cardWithFrontmatter adds a card and inserts extra frontmatter lines after
// the opening "---", returning the store, the card's id, and its path.
func cardWithFrontmatter(t *testing.T, extra string) (Store, string, string) {
	t.Helper()
	store := New(t.TempDir())
	id, err := store.Add("go", "Q", "A")
	if err != nil {
		t.Fatal(err)
	}
	path, err := store.Path(id.String())
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(content), "---\n", "---\n"+extra, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	return store, id.String(), path
}

const sm2Progress = "ease_factor: 2.6\ninterval: 6\nrepetitions: 2\nlast_reviewed: \"2026-09-30T12:00:00Z\"\n"

func TestCardWithSM2ProgressLoadsAsNeverReviewed(t *testing.T) {
	store, id, _ := cardWithFrontmatter(t, sm2Progress)

	_, progress, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	if progress.LastReviewed != nil || progress.HalfLife != 0 || progress.Interval != 0 {
		t.Fatalf("schedule of an SM-2 card = %+v, want never reviewed", progress)
	}
}

func TestAnySM2PropertyMarksTheCardAsNew(t *testing.T) {
	// No ease_factor: repetitions alone is enough.
	store, id, _ := cardWithFrontmatter(t, "interval: 6\nrepetitions: 2\nlast_reviewed: \"2026-09-30T12:00:00Z\"\n")

	if _, progress, err := store.Load(id); err != nil || progress.LastReviewed != nil {
		t.Fatalf("card with repetitions but no ease_factor: schedule %+v, error %v; want never reviewed", progress, err)
	}
}

func TestSavingAReviewedSM2CardDropsTheOldProperties(t *testing.T) {
	store, id, path := cardWithFrontmatter(t, "custom: keep\n"+sm2Progress)
	_, progress, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SaveSchedule(id, schedule.Review(progress, true, time.Now())); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(saved)
	if strings.Contains(text, "ease_factor:") || strings.Contains(text, "repetitions:") {
		t.Fatalf("SM-2 properties survived saving: %s", text)
	}
	if !strings.Contains(text, "half_life:") || !strings.Contains(text, "difficulty:") || !strings.Contains(text, "custom: keep\n") {
		t.Fatalf("saved card is missing new or custom properties: %s", text)
	}
	if strings.Count(text, "interval:") != 1 || strings.Count(text, "last_reviewed:") != 1 {
		t.Fatalf("schedule properties duplicated: %s", text)
	}
}

func TestIncompleteSchedulePropertiesAreStillAnError(t *testing.T) {
	store, id, _ := cardWithFrontmatter(t, "interval: 6\nlast_reviewed: \"2026-09-30T12:00:00Z\"\n")

	if _, _, err := store.Load(id); err == nil {
		t.Fatal("loaded a card with interval and last_reviewed but no half_life or difficulty")
	}
}
