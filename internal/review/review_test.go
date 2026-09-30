package review

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/prodzombie/gostudy/internal/cardstore"
	"github.com/prodzombie/gostudy/internal/schedule"
)

func TestReviewSelectsDueCardsAndPersistsGrades(t *testing.T) {
	store := cardstore.New(t.TempDir())
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	add := func(tag, question string) string {
		id, err := store.Add(tag, question, "secret answer")
		if err != nil {
			t.Fatal(err)
		}
		return id.String()
	}
	never := add("go", "Never reviewed")
	due := add("go", "Due now")
	future := add("go", "Not due yet")
	other := add("other", "Filtered out")
	for id, reviewedAt := range map[string]time.Time{due: now.AddDate(0, 0, -1), future: now} {
		_, progress, err := store.Load(id)
		if err != nil {
			t.Fatal(err)
		}
		progress, err = schedule.Review(progress, 5, reviewedAt)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.SaveSchedule(id, progress); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := Run(store, "go", strings.NewReader("\n6\ninvalid\n5\n\n5\n"), &out, func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Reviewed 2 cards.") || !strings.Contains(out.String(), "Enter a grade from 0 to 5.") || strings.Contains(out.String(), "Not due yet") || strings.Contains(out.String(), "Filtered out") {
		t.Fatalf("session output: %s", &out)
	}
	for id, wantInterval := range map[string]int{never: 1, due: 6, future: 1} {
		_, progress, err := store.Load(id)
		if err != nil {
			t.Fatal(err)
		}
		if progress.Interval != wantInterval || !progress.LastReviewed.Equal(now) {
			t.Fatalf("schedule for %s: %+v", id, progress)
		}
	}
	_, progress, err := store.Load(other)
	if err != nil {
		t.Fatal(err)
	}
	if progress.LastReviewed != nil {
		t.Fatal("reviewed a filtered card")
	}
	out.Reset()
	if err := Run(store, "go", strings.NewReader(""), &out, func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	if out.String() != "No cards are due.\n" {
		t.Fatalf("second session: %s", &out)
	}
}

type revealReader struct {
	t    *testing.T
	out  *bytes.Buffer
	step int
}

func (r *revealReader) Read(p []byte) (int, error) {
	switch r.step {
	case 0:
		if strings.Contains(r.out.String(), "secret answer") {
			r.t.Fatal("answer revealed before confirmation")
		}
		r.step++
		return copy(p, "\n"), nil
	case 1:
		if !strings.Contains(r.out.String(), "secret answer") {
			r.t.Fatal("answer missing before grade prompt")
		}
		r.step++
		return copy(p, "4\n"), nil
	default:
		return 0, io.EOF
	}
}

func TestAnswerIsHiddenUntilReveal(t *testing.T) {
	store := cardstore.New(t.TempDir())
	if _, err := store.Add("go", "Question", "secret answer"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	in := &revealReader{t: t, out: &out}
	if err := Run(store, "", in, &out, time.Now); err != nil {
		t.Fatal(err)
	}
}

func TestQuitAndEOFDoNotGradeCurrentCard(t *testing.T) {
	for _, input := range []string{"q\n", "\nq\n", "", "\n", "\ninvalid\n"} {
		store := cardstore.New(t.TempDir())
		id, err := store.Add("go", "Q", "A")
		if err != nil {
			t.Fatal(err)
		}
		before, _ := store.Read(id.String())
		var out bytes.Buffer
		if err := Run(store, "", strings.NewReader(input), &out, time.Now); err != nil {
			t.Fatal(err)
		}
		after, _ := store.Read(id.String())
		if !bytes.Equal(before, after) {
			t.Errorf("input %q changed ungraded card", input)
		}
	}
}

func TestCompletedGradeSurvivesQuittingNextCard(t *testing.T) {
	store := cardstore.New(t.TempDir())
	for i := 0; i < 2; i++ {
		if _, err := store.Add("go", "Q", "A"); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := store.List("")
	var out bytes.Buffer
	if err := Run(store, "", strings.NewReader("\n2\nq\n"), &out, time.Now); err != nil {
		t.Fatal(err)
	}
	_, first, err := store.Load(entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := store.Load(entries[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.LastReviewed == nil || first.Interval != 1 || first.Repetitions != 0 || second.LastReviewed != nil {
		t.Fatalf("first: %+v; second: %+v", first, second)
	}
}
