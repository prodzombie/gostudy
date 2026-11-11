package review

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestMarkdownPresentation(t *testing.T) {
	for _, width := range []int{24, 80} {
		renderer, err := newMarkdownRenderer(width)
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := renderer.Render("A paragraph with **important words** and some additional words that should wrap at word boundaries.\n\n```go\n  fmt.Println(1)\n```")
		if err != nil {
			t.Fatal(err)
		}
		plain := ansi.Strip(rendered)
		if strings.Contains(plain, "**") || strings.Contains(plain, "```") || !strings.Contains(plain, "  fmt.Println(1)") {
			t.Fatalf("Markdown or code indentation not rendered: %q", plain)
		}
		for _, line := range strings.Split(plain, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("line exceeds width %d: %q", width, line)
			}
		}
	}
}
