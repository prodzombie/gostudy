package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/emiliopalmerini/gostudy/internal/card"
	"github.com/emiliopalmerini/gostudy/internal/cardstore"
	"github.com/emiliopalmerini/gostudy/internal/editor"
	"github.com/emiliopalmerini/gostudy/internal/huid"
)

// Run executes the CLI and returns its process exit status.
func Run(args []string, in io.Reader, out, errOut io.Writer) int {
	usage := func(w io.Writer) {
		fmt.Fprint(w, "Usage: gostudy [-d DIR] COMMAND [OPTIONS] [ARGUMENTS]\n\nCommands:\n  add  [-e] -t TAG [-q QUESTION] [-a ANSWER]\n  list [-t TAG]\n  show HUID\n  edit HUID\n  help\n\nCards live in DIR/cards/TAG/HUID/CARD.md (DIR defaults to .).\n")
	}
	if len(args) == 0 {
		usage(errOut)
		return 2
	}
	fs := flag.NewFlagSet("gostudy", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() { usage(errOut) }
	root := fs.String("d", ".", "directory containing cards")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	args = fs.Args()
	if len(args) == 0 {
		usage(errOut)
		return 2
	}
	switch args[0] {
	case "help":
		if len(args) != 1 {
			usage(errOut)
			return 2
		}
		usage(out)
		return 0
	case "add":
		return addCommand(*root, args[1:], in, out, errOut)
	case "list":
		return listCommand(*root, args[1:], out, errOut)
	case "edit":
		return editCommand(*root, args[1:], in, out, errOut)
	case "show":
		return showCommand(*root, args[1:], out, errOut)
	default:
		fmt.Fprintf(errOut, "gostudy: unknown command %q\n", args[0])
		usage(errOut)
		return 2
	}
}

func flagStatus(err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}

func commandFlags(name, synopsis string, errOut io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() { fmt.Fprintf(errOut, "Usage: gostudy [-d DIR] %s %s\n", name, synopsis) }
	return fs
}

func addCommand(root string, args []string, in io.Reader, out, errOut io.Writer) int {
	fs := commandFlags("add", "[-e] -t TAG [-q QUESTION] [-a ANSWER]", errOut)
	open := fs.Bool("e", false, "open the card in EDITOR")
	tag := fs.String("t", "", "card tag")
	question := fs.String("q", "", "question text")
	answer := fs.String("a", "", "answer text")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 0 || !card.ValidTag(*tag) || !*open && (strings.TrimSpace(*question) == "" || strings.TrimSpace(*answer) == "") {
		fmt.Fprintln(errOut, "gostudy: add requires a safe tag, a question, and an answer")
		fs.Usage()
		return 2
	}
	id, err := cardstore.New(root).Add(*tag, *question, *answer)
	if err != nil {
		return commandError(errOut, err)
	}
	fmt.Fprintln(out, id)
	if *open {
		path, err := cardstore.New(root).Path(id.String())
		if err != nil {
			return commandError(errOut, err)
		}
		if err := editor.Open(path, in, out, errOut); err != nil {
			return commandError(errOut, err)
		}
	}
	return 0
}

func listCommand(root string, args []string, out, errOut io.Writer) int {
	fs := commandFlags("list", "[-t TAG]", errOut)
	tag := fs.String("t", "", "show only this tag")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 0 || *tag != "" && !card.ValidTag(*tag) {
		fs.Usage()
		return 2
	}
	entries, err := cardstore.New(root).List(*tag)
	if err != nil {
		return commandError(errOut, err)
	}
	for _, entry := range entries {
		fmt.Fprintf(out, "%s\t%s\n", entry.ID, entry.Tag)
	}
	return 0
}

func showCommand(root string, args []string, out, errOut io.Writer) int {
	fs := commandFlags("show", "HUID", errOut)
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 1 || !huid.Valid(fs.Arg(0)) {
		fs.Usage()
		return 2
	}
	content, err := cardstore.New(root).Read(fs.Arg(0))
	if err != nil {
		return commandError(errOut, err)
	}
	if _, err := out.Write(content); err != nil {
		return commandError(errOut, err)
	}
	return 0
}

func commandError(errOut io.Writer, err error) int {
	fmt.Fprintf(errOut, "gostudy: %v\n", err)
	return 1
}

func editCommand(root string, args []string, in io.Reader, out, errOut io.Writer) int {
	fs := commandFlags("edit", "HUID", errOut)
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 1 || !huid.Valid(fs.Arg(0)) {
		fs.Usage()
		return 2
	}
	path, err := cardstore.New(root).Path(fs.Arg(0))
	if err != nil {
		return commandError(errOut, err)
	}
	if err := editor.Open(path, in, out, errOut); err != nil {
		return commandError(errOut, err)
	}
	return 0
}
