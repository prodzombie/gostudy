// Package cardstore persists cards as Markdown files with YAML frontmatter.
package cardstore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prodzombie/gostudy/internal/card"
	"github.com/prodzombie/gostudy/internal/huid"
)

// Store locates cards under a root directory.
type Store struct{ root string }

// New selects the directory containing the cards folder.
func New(root string) Store { return Store{root: root} }

// Add creates a card, optionally with blank question and answer sections.
func (s Store) Add(tag, question, answer string) (huid.HUID, error) {
	if !card.ValidTag(tag) {
		return huid.HUID{}, fmt.Errorf("add requires a safe tag, a question, and an answer")
	}
	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return huid.HUID{}, err
	}
	id, err := huid.New(time.Now(), hex.EncodeToString(suffix[:]))
	if err != nil {
		return huid.HUID{}, err
	}
	path := filepath.Join(s.root, "cards", tag, id.String())
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return huid.HUID{}, err
	}
	if err := os.Mkdir(path, 0755); err != nil {
		return huid.HUID{}, err
	}
	file, err := os.OpenFile(filepath.Join(path, "CARD.md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return huid.HUID{}, err
	}
	content := fmt.Sprintf("---\nid: %s\ntags:\n  - %s\n---\n\n## Question\n\n%s\n\n## Answer\n\n%s\n", yamlString(id.String()), yamlString(tag), strings.TrimSpace(question), strings.TrimSpace(answer))
	if _, err := io.WriteString(file, content); err != nil {
		file.Close()
		return huid.HUID{}, err
	}
	if err := file.Close(); err != nil {
		return huid.HUID{}, err
	}
	return id, nil
}

// JSON strings use YAML-compatible double quoting and escape special characters.
func yamlString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// Entry identifies a persisted card.
type Entry struct{ ID, Tag string }

// List returns cards in tag and ID order, optionally filtered by tag.
func (s Store) List(filter string) ([]Entry, error) {
	base := filepath.Join(s.root, "cards")
	tags, err := os.ReadDir(base)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, tag := range tags {
		if !tag.IsDir() || !card.ValidTag(tag.Name()) || filter != "" && tag.Name() != filter {
			continue
		}
		ids, err := os.ReadDir(filepath.Join(base, tag.Name()))
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if !id.IsDir() || !huid.Valid(id.Name()) {
				continue
			}
			info, err := os.Stat(filepath.Join(base, tag.Name(), id.Name(), "CARD.md"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if info.Mode().IsRegular() {
				entries = append(entries, Entry{id.Name(), tag.Name()})
			}
		}
	}
	return entries, nil
}

// Read returns a card's complete Markdown file.
func (s Store) Read(id string) ([]byte, error) {
	path, err := s.Path(id)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// Path locates the Markdown file for an existing card.
func (s Store) Path(id string) (string, error) {
	entries, err := s.List("")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.ID == id {
			return filepath.Abs(filepath.Join(s.root, "cards", entry.Tag, entry.ID, "CARD.md"))
		}
	}
	return "", fmt.Errorf("card %s not found", id)
}
