package review

import (
	"fmt"
	"io"

	"charm.land/glamour/v2"

	"github.com/prodzombie/gostudy/internal/card"
)

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

// Plain output appends the answer, preserving a readable session transcript.
func showCard(out io.Writer, content card.Card, revealed bool, position, total int) error {
	if !revealed {
		if _, err := fmt.Fprintf(out, "\nCard %d of %d · %s\n\nQuestion\n\n%s\n\n", position, total, content.Tags.Value, content.Question); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(out, "Answer\n\n%s\n\n", content.Answer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(out, "[%s] · %s\n\n", content.Tags.Value, content.ID)
	return err
}
