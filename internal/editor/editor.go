// Package editor opens files in the user's configured terminal editor.
package editor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Open runs EDITOR with a file argument and waits until the editor exits.
// EDITOR is interpreted as a shell command; the path is passed separately.
func Open(path string, in io.Reader, out, errOut io.Writer) error {
	command := os.Getenv("EDITOR")
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("EDITOR is not set")
	}
	cmd := exec.Command("sh", "-c", "exec "+command+` "$@"`, "gostudy-editor", path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = in, out, errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	return nil
}
