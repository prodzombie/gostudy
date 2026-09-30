package cli

import (
	"io"
	"time"

	"github.com/emiliopalmerini/gostudy/internal/card"
	"github.com/emiliopalmerini/gostudy/internal/cardstore"
	"github.com/emiliopalmerini/gostudy/internal/review"
)

func reviewCommand(root string, args []string, in io.Reader, out, errOut io.Writer) int {
	fs := commandFlags("review", "[-t TAG]", errOut)
	tag := fs.String("t", "", "review only this tag")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 0 || *tag != "" && !card.ValidTag(*tag) {
		fs.Usage()
		return 2
	}
	if err := review.Run(cardstore.New(root), *tag, in, out, time.Now); err != nil {
		return commandError(errOut, err)
	}
	return 0
}
