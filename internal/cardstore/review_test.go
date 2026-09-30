package cardstore

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/emiliopalmerini/gostudy/internal/schedule"
)

func TestLoadAndSaveSchedulePreserveMarkdown(t *testing.T) {
	store := New(t.TempDir())
	question := "What does this print?\n\n```go\n## Answer\n```"
	id, err := store.Add("go", question, "A heading.")
	if err != nil {
		t.Fatal(err)
	}
	path, err := store.Path(id.String())
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate properties added in a Markdown editor, with Windows line endings.
	content := strings.Replace(string(original), "tags:\n", "aliases:\n  - sample\ncustom: 'keep this'\ntags:\n", 1)
	content = strings.ReplaceAll(content, "\n", "\r\n")
	if err := os.WriteFile(path, []byte(content), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0640); err != nil {
		t.Fatal(err)
	}
	card, progress, err := store.Load(id.String())
	if err != nil {
		t.Fatal(err)
	}
	if card.ID != id || card.Tags.Value != "go" || card.Question != strings.ReplaceAll(question, "\n", "\r\n") || progress.LastReviewed != nil {
		t.Fatalf("loaded card: %+v; schedule: %+v", card, progress)
	}
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	for _, interval := range []int{1, 6} {
		progress, err = schedule.Review(progress, 5, now)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.SaveSchedule(id.String(), progress); err != nil {
			t.Fatal(err)
		}
		_, progress, err = store.Load(id.String())
		if err != nil {
			t.Fatal(err)
		}
		if progress.Interval != interval || !progress.LastReviewed.Equal(now) {
			t.Fatalf("persisted schedule: %+v", progress)
		}
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	oldBody := strings.SplitN(content, "\r\n---\r\n", 2)[1]
	newBody := strings.SplitN(string(saved), "\r\n---\r\n", 2)[1]
	if oldBody != newBody || !strings.Contains(string(saved), "aliases:\r\n  - sample\r\ncustom: 'keep this'\r\n") {
		t.Fatalf("card content changed: %s", saved)
	}
	if strings.Count(string(saved), "ease_factor:") != 1 {
		t.Fatalf("duplicated schedule: %s", saved)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0640 {
		t.Fatalf("permissions changed: %v", info.Mode())
	}
}

func TestLoadRejectsMalformedCards(t *testing.T) {
	for _, content := range []string{
		"## Question\nQ\n## Answer\nA\n",
		"---\ninterval: bad\n---\n## Question\nQ\n## Answer\nA\n",
		"---\ninterval: 1\ninterval: 2\n---\n## Question\nQ\n## Answer\nA\n",
		"---\nid: sample\n---\n## Question\n\n## Answer\nA\n",
	} {
		store := New(t.TempDir())
		id, err := store.Add("go", "Q", "A")
		if err != nil {
			t.Fatal(err)
		}
		path, _ := store.Path(id.String())
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := store.Load(id.String()); err == nil {
			t.Errorf("accepted malformed card: %q", content)
		}
	}
}
