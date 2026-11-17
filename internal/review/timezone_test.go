package review

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/prodzombie/gostudy/internal/cardstore"
	"github.com/prodzombie/gostudy/internal/schedule"
)

// cardReviewedTwoLocalDaysAgo saves a card first reviewed at 23:30 on
// 17 November in Rome and returns the store, the card's id, a session clock at
// 00:30 on 19 November in Rome, and the schedule a recalled review at that
// time should produce. The two times are two local days apart but only one
// day apart in UTC.
func cardReviewedTwoLocalDaysAgo(t *testing.T) (cardstore.Store, string, time.Time, schedule.Schedule) {
	t.Helper()
	rome, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	store := cardstore.New(t.TempDir())
	id, err := store.Add("go", "Question?", "Answer")
	if err != nil {
		t.Fatal(err)
	}
	first := schedule.Review(schedule.Schedule{}, true, time.Date(2026, time.November, 17, 23, 30, 0, 0, rome))
	if err := store.SaveSchedule(id.String(), first); err != nil {
		t.Fatal(err)
	}
	_, stored, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	session := time.Date(2026, time.November, 19, 0, 30, 0, 0, rome)
	return store, id.String(), session, schedule.Review(stored, true, session)
}

func TestLineReviewCountsElapsedDaysInTheSessionsTimeZone(t *testing.T) {
	store, id, session, want := cardReviewedTwoLocalDaysAgo(t)
	var out bytes.Buffer
	if err := Run(store, "go", strings.NewReader("\n5\n"), &out, func() time.Time { return session }); err != nil {
		t.Fatal(err)
	}
	_, got, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.HalfLife != want.HalfLife {
		t.Fatalf("half-life = %v, want %v from two elapsed local days", got.HalfLife, want.HalfLife)
	}
}

func TestScreenReviewCountsElapsedDaysInTheSessionsTimeZone(t *testing.T) {
	store, id, session, want := cardReviewedTwoLocalDaysAgo(t)
	c, progress, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: func() time.Time { return session }, viewport: viewport.New()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	keyPress(m, tea.KeyEnter)
	keyPress(m, '5')
	keyPress(m, tea.KeyEnter)
	_, got, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.HalfLife != want.HalfLife {
		t.Fatalf("half-life = %v, want %v from two elapsed local days", got.HalfLife, want.HalfLife)
	}
}
