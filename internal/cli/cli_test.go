package cli

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
		status := Run(append([]string{"-d", root}, args...), strings.NewReader(""), &out, &errOut)
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
		if status := Run(append([]string{"-d", root}, args...), strings.NewReader(""), &out, &errOut); status != 2 || out.Len() != 0 {
			t.Errorf("%v: status %d, output %q", args, status, out.String())
		}
	}
	if _, err := os.Stat(filepath.Join(root, "cards")); !os.IsNotExist(err) {
		t.Fatalf("cards directory created: %v", err)
	}
}

func TestEditorCommands(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '\\nEdited\\n' >> \"$1\"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", "sh '"+script+"'")
	var out, errOut bytes.Buffer
	status := Run([]string{"-d", root, "add", "-e", "-t", "go"}, strings.NewReader(""), &out, &errOut)
	if status != 0 {
		t.Fatalf("add -e: %d, %s", status, errOut.String())
	}
	id := strings.TrimSpace(out.String())
	out.Reset()
	status = Run([]string{"-d", root, "edit", id}, strings.NewReader(""), &out, &errOut)
	if status != 0 {
		t.Fatalf("edit: %d, %s", status, errOut.String())
	}
	content, err := os.ReadFile(filepath.Join(root, "cards", "go", id, "CARD.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "Edited") != 2 {
		t.Fatalf("editor did not modify card twice: %s", content)
	}
}

func TestReviewCommand(t *testing.T) {
	root := t.TempDir()
	var out, errOut bytes.Buffer
	if status := Run([]string{"-d", root, "add", "-t", "go", "-q", "Q", "-a", "A"}, strings.NewReader(""), &out, &errOut); status != 0 {
		t.Fatal(errOut.String())
	}
	out.Reset()
	if status := Run([]string{"-d", root, "review", "-t", "go"}, strings.NewReader("\n5\n"), &out, &errOut); status != 0 || !strings.Contains(out.String(), "Reviewed 1 card.") {
		t.Fatalf("review output: %s; error: %s; status: %d", &out, &errOut, status)
	}
	for _, args := range [][]string{{"review", "-t", "../bad"}, {"review", "extra"}} {
		if status := Run(args, strings.NewReader(""), &out, &errOut); status != 2 {
			t.Errorf("%v: status %d", args, status)
		}
	}
}
