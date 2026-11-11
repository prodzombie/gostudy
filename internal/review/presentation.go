package review

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/glamour/v2"
	"golang.org/x/term"

	"github.com/prodzombie/gostudy/internal/card"
)

// Leave one column free to avoid triggering terminal autowrap at the edge.
// If querying the terminal fails, use the same readable cap as a wide terminal.
func terminalWidth(out io.Writer) int {
	if file, ok := out.(*os.File); ok {
		if width, _, err := term.GetSize(int(file.Fd())); err == nil && width > 1 {
			return min(80, width-1)
		}
	}
	return 80
}

// Theme-neutral styles inherit the user's terminal foreground and background.
// Code keeps its indentation; prose is wrapped by Glamour.
func newMarkdownRenderer(width int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStylesFromJSONBytes([]byte(`{
            "document": {"margin": 0},
            "heading": {"bold": true, "block_suffix": "\n"},
            "strong": {"bold": true},
            "emph": {"italic": true},
            "strikethrough": {"crossed_out": true},
            "code": {"bold": true},
            "code_block": {"indent": 2},
            "list": {"level_indent": 2},
            "item": {"block_prefix": "• "},
            "enumeration": {"block_prefix": ". "},
            "block_quote": {"indent": 1, "indent_token": "│ "},
            "link": {"underline": true},
            "task": {"ticked": "[x] ", "unticked": "[ ] "}
        }`)),
	)
}

// Terminal frames redraw both sections so metadata stays below the study content.
// Plain output appends the answer, preserving a readable session transcript.
func showCard(out io.Writer, content card.Card, revealed, terminal bool, position, total int, confirmation string) error {
	question, answer := content.Question, content.Answer
	if terminal {
		renderer, err := newMarkdownRenderer(terminalWidth(out))
		if err != nil {
			return err
		}
		question, err = renderer.Render(question)
		if err != nil {
			return err
		}
		question = strings.Trim(question, "\n")
		if revealed {
			answer, err = renderer.Render(answer)
			if err != nil {
				return err
			}
			answer = strings.Trim(answer, "\n")
		}
	}
	bold, dim, reset := "", "", ""
	if terminal {
		bold, dim, reset = "\x1b[1m", "\x1b[2m", "\x1b[0m"
		// Clear the visible screen and home the cursor, preserving scrollback.
		if _, err := fmt.Fprint(out, "\x1b[2J\x1b[H"); err != nil {
			return err
		}
	}
	if !revealed || terminal {
		if _, err := fmt.Fprintf(out, "\nCard %d of %d · %s\n", position, total, content.Tags.Value); err != nil {
			return err
		}
		if terminal && confirmation != "" {
			if _, err := fmt.Fprintln(out, confirmation); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(out, "\n%sQuestion%s\n\n%s\n\n", bold, reset, question); err != nil {
			return err
		}
	}
	if revealed {
		if _, err := fmt.Fprintf(out, "%sAnswer%s\n\n%s\n\n", bold, reset, answer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(out, "%s[%s] · %s%s\n\n", dim, content.Tags.Value, content.ID, reset)
	return err
}
