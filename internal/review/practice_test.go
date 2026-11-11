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

func TestPracticeIncludesFutureCardsWithoutWriting(t *testing.T) {
	store := cardstore.New(t.TempDir())
	now := time.Now()
	id, err := store.Add("go", "Future card", "Answer")
	if err != nil {
		t.Fatal(err)
	}
	_, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	progress, err = schedule.Review(progress, 4, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSchedule(id.String(), progress); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add("other", "Filtered card", "Answer"); err != nil {
		t.Fatal(err)
	}
	before, err := store.Read(id.String())
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RunWithOptions(store, "go", strings.NewReader("\n4\n\n"), &out, func() time.Time { return now }, Options{Practice: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Future card") || !strings.Contains(out.String(), "Practiced 1 card") || strings.Contains(out.String(), "Filtered card") || strings.Contains(out.String(), "Grade 0") {
		t.Fatalf("practice output: %s", &out)
	}
	after, err := store.Read(id.String())
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("practice changed the card")
	}

	c, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: time.Now, viewport: viewport.New(), practice: true}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	keyPress(m, tea.KeyEnter)
	keyPress(m, '4')
	if m.grade != "" || !strings.Contains(m.footer(), "Enter: next card") {
		t.Fatal("practice offered grading")
	}
	keyPress(m, tea.KeyEnter)
	after, err = store.Read(id.String())
	if err != nil || !bytes.Equal(before, after) || m.index != 1 {
		t.Fatal("screen practice changed card or failed to advance")
	}
}
