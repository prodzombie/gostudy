package review

import (
	"encoding/json"
	"fmt"
	"io"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"

	"github.com/prodzombie/gostudy/internal/card"
)

// The plain renderer retains typographic emphasis without explicit colors.
func newMarkdownRenderer(width int) (*glamour.TermRenderer, error) {
	return themedMarkdownRenderer(width, palette{noColor: true})
}

func themedMarkdownRenderer(width int, p palette) (*glamour.TermRenderer, error) {
	var style ansi.StyleConfig
	if err := json.Unmarshal([]byte(`{
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
        }`), &style); err != nil {
		return nil, err
	}
	if !p.noColor {
		accent, _, _ := p.colors()
		style.Heading.Color = &accent
		style.Code.Color = &accent
		style.Link.Color = &accent
		syntax := styles.DarkStyleConfig.CodeBlock.Chroma
		if !p.dark {
			syntax = styles.LightStyleConfig.CodeBlock.Chroma
		}
		style.CodeBlock.Chroma = syntax
	}
	return glamour.NewTermRenderer(glamour.WithWordWrap(width), glamour.WithStyles(style))
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
