package editor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenPassesArgumentsAndStreams(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "editor with spaces.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\\n' \"$1\" \"$2\"\ncat\nprintf 'diagnostic' >&2\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", "sh '"+script+"' -w")
	path := filepath.Join(dir, "card ' $HOME.md")
	var out, errOut bytes.Buffer
	if err := Open(path, strings.NewReader("input"), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if want := "-w\n" + path + "\ninput"; out.String() != want {
		t.Fatalf("output = %q; want %q", out.String(), want)
	}
	if errOut.String() != "diagnostic" {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestOpenReportsConfigurationAndProcessErrors(t *testing.T) {
	for _, command := range []string{"", "sh -c 'exit 7'"} {
		t.Setenv("EDITOR", command)
		var out bytes.Buffer
		if err := Open("CARD.md", strings.NewReader(""), &out, &out); err == nil {
			t.Errorf("EDITOR=%q: expected an error", command)
		}
	}
}
