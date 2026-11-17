package review

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/prodzombie/gostudy/internal/cardstore"
	"github.com/prodzombie/gostudy/internal/schedule"
)

// cardReviewedBefore saves a card first reviewed three days before now, so it
// is due and its next review depends on the answer.
func cardReviewedBefore(t *testing.T, now time.Time) (cardstore.Store, string, schedule.Schedule) {
	t.Helper()
	store := cardstore.New(t.TempDir())
	id, err := store.Add("go", "Question?", "Answer")
	if err != nil {
		t.Fatal(err)
	}
	first := schedule.Review(schedule.Schedule{}, true, now.AddDate(0, 0, -3))
	if err := store.SaveSchedule(id.String(), first); err != nil {
		t.Fatal(err)
	}
	_, stored, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	return store, id.String(), stored
}

func loadSchedule(t *testing.T, store cardstore.Store, id string) schedule.Schedule {
	t.Helper()
	_, progress, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	return progress
}

func revealedScreen(t *testing.T, store cardstore.Store, id string, now time.Time) *screen {
	t.Helper()
	c, progress, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: func() time.Time { return now }, viewport: newViewport()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	keyPress(m, tea.KeyEnter)
	return m
}

func TestLineReviewSavesRecalledForR(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	if err := Run(store, "go", strings.NewReader("\nr\n"), &bytes.Buffer{}, func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	if got, want := loadSchedule(t, store, id), schedule.Review(before, true, now); got.HalfLife != want.HalfLife || got.Difficulty != want.Difficulty {
		t.Fatalf("schedule after r = %+v, want a recalled review %+v", got, want)
	}
}

func TestLineReviewSavesForgotForF(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	if err := Run(store, "go", strings.NewReader("\nf\n"), &bytes.Buffer{}, func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	if got, want := loadSchedule(t, store, id), schedule.Review(before, false, now); got.HalfLife != want.HalfLife || got.Difficulty != want.Difficulty {
		t.Fatalf("schedule after f = %+v, want a forgotten review %+v", got, want)
	}
}

func TestLineReviewRepromptsUntilROrF(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	var out bytes.Buffer
	if err := Run(store, "go", strings.NewReader("\n5\nrecalled\nf\n"), &out, func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "Enter r or f.") != 2 {
		t.Fatalf("session output: %s", &out)
	}
	if got := loadSchedule(t, store, id); got.Difficulty != schedule.Review(before, false, now).Difficulty {
		t.Fatalf("schedule after re-prompting = %+v, want the forgotten review", got)
	}
}

func TestScreenSavesRecalledAfterSelectingR(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	m := revealedScreen(t, store, id, now)
	keyPress(m, 'r')
	keyPress(m, tea.KeyEnter)
	if got, want := loadSchedule(t, store, id), schedule.Review(before, true, now); got.HalfLife != want.HalfLife || got.Difficulty != want.Difficulty {
		t.Fatalf("schedule after r and Enter = %+v, want a recalled review %+v", got, want)
	}
}

func TestScreenSavesForgotAfterSelectingF(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	m := revealedScreen(t, store, id, now)
	keyPress(m, 'f')
	keyPress(m, tea.KeyEnter)
	if got, want := loadSchedule(t, store, id), schedule.Review(before, false, now); got.HalfLife != want.HalfLife || got.Difficulty != want.Difficulty {
		t.Fatalf("schedule after f and Enter = %+v, want a forgotten review %+v", got, want)
	}
}

func TestScreenShowsTheSelectedAnswer(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, _ := cardReviewedBefore(t, now)
	m := revealedScreen(t, store, id, now)
	keyPress(m, 'f')
	if !strings.Contains(m.footer(), "Answer: forgot") {
		t.Fatalf("footer after f = %q, want the selected answer", m.footer())
	}
}

func TestScreenBackspaceClearsTheSelectedAnswer(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	store, id, before := cardReviewedBefore(t, now)
	m := revealedScreen(t, store, id, now)
	keyPress(m, 'f')
	keyPress(m, tea.KeyBackspace)
	keyPress(m, tea.KeyEnter)
	if got := loadSchedule(t, store, id); !got.LastReviewed.Equal(*before.LastReviewed) || m.index != 0 {
		t.Fatalf("Enter after clearing the answer saved %+v", got)
	}
}

func TestPressingFBeforeTheAnswerIsRevealedNeitherScrollsNorSelects(t *testing.T) {
	store := cardstore.New(t.TempDir())
	id, err := store.Add("go", strings.Repeat("A long question paragraph.\n\n", 40), "Answer")
	if err != nil {
		t.Fatal(err)
	}
	c, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: time.Now, viewport: newViewport()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	keyPress(m, 'f')
	if m.viewport.YOffset() != 0 || m.grade != "" {
		t.Fatalf("after f before reveal: offset %d, selected %q; want no scroll and no selection", m.viewport.YOffset(), m.grade)
	}
}
