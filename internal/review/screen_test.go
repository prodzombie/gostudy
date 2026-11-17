package review

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/prodzombie/gostudy/internal/cardstore"
)

func keyPress(m *screen, code rune) {
	m.Update(tea.KeyPressMsg{Code: code})
}

func TestScreenRevealScrollResizeAndSave(t *testing.T) {
	store := cardstore.New(t.TempDir())
	id, err := store.Add("go", "Question?", strings.Repeat("Answer paragraph with several words.\n\n", 40))
	if err != nil {
		t.Fatal(err)
	}
	c, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: time.Now, viewport: viewport.New()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if strings.Contains(m.viewport.View(), "Answer paragraph") {
		t.Fatal("answer leaked")
	}
	keyPress(m, 'r')
	_, early, _ := store.Load(id.String())
	if early.LastReviewed != nil {
		t.Fatal("grade saved before reveal")
	}
	keyPress(m, tea.KeyEnter)
	if !strings.Contains(m.viewport.View(), "Answer paragraph") || m.viewport.YOffset() == 0 {
		t.Fatal("reveal did not navigate to answer")
	}
	before := m.viewport.YOffset()
	keyPress(m, tea.KeyDown)
	if m.viewport.YOffset() <= before {
		t.Fatal("down did not scroll")
	}
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	if m.viewport.Width() != 39 || m.viewport.Height() >= 20 {
		t.Fatal("viewport did not resize")
	}
	keyPress(m, 'r')
	_, saved, err := store.Load(id.String())
	if err != nil || saved.LastReviewed == nil || m.index != 1 {
		t.Fatalf("save failed: %v", err)
	}
}

func TestScreenQuitDoesNotSave(t *testing.T) {
	store := cardstore.New(t.TempDir())
	id, err := store.Add("go", "Q", "A")
	if err != nil {
		t.Fatal(err)
	}
	c, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	m := &screen{store: store, due: []dueCard{{c, progress}}, now: time.Now, viewport: viewport.New()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	keyPress(m, tea.KeyEnter)
	keyPress(m, 'q')
	_, saved, err := store.Load(id.String())
	if err != nil || saved.LastReviewed != nil {
		t.Fatalf("quit saved grade: %v", err)
	}
}
