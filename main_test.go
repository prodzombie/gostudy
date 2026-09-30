package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCardCommands(t *testing.T) {
	root := t.TempDir()
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(append([]string{"-d", root}, args...), &out, &errOut)
		return status, out.String(), errOut.String()
	}

	status, idOutput, diagnostic := call("add", "-t", "go", "-q", "What is a slice?", "-a", "A view over an array.")
	if status != 0 || diagnostic != "" {
		t.Fatalf("add: status %d, diagnostic %q", status, diagnostic)
	}
	id := strings.TrimSpace(idOutput)
	path := filepath.Join(root, "cards", "go", id, "CARD.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "---\nid: \"" + id + "\"\ntags:\n  - \"go\"\n---\n\n## Question\n\nWhat is a slice?\n\n## Answer\n\nA view over an array.\n"
	if string(content) != want {
		t.Errorf("CARD.md = %q; want %q", content, want)
	}
	status, listed, diagnostic := call("list", "-t", "go")
	if status != 0 || listed != id+"\tgo\n" || diagnostic != "" {
		t.Errorf("list: status %d, output %q, diagnostic %q", status, listed, diagnostic)
	}
	status, shown, diagnostic := call("show", id)
	if status != 0 || shown != want || diagnostic != "" {
		t.Errorf("show: status %d, output %q, diagnostic %q", status, shown, diagnostic)
	}
}

func TestInvalidInputsDoNotCreateCards(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"add", "-t", "../outside", "-q", "Q", "-a", "A"},
		{"add", "-t", "123", "-q", "Q", "-a", "A"},
		{"add", "-t", "go", "-q", "Q"},
		{"add", "-t", "go", "-q", "Q", "-a", "A", "extra"},
	} {
		var out, errOut bytes.Buffer
		if status := run(append([]string{"-d", root}, args...), &out, &errOut); status != 2 || out.Len() != 0 {
			t.Errorf("%v: status %d, output %q", args, status, out.String())
		}
	}
	if _, err := os.Stat(filepath.Join(root, "cards")); !os.IsNotExist(err) {
		t.Fatalf("cards directory created: %v", err)
	}
}
