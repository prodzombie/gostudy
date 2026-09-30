package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliopalmerini/gostudy/internal/huid"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	usage := func(w io.Writer) {
		fmt.Fprint(w, "Usage: gostudy [-d DIR] COMMAND [OPTIONS] [ARGUMENTS]\n\nCommands:\n  add  -t TAG -q QUESTION -a ANSWER\n  list [-t TAG]\n  show HUID\n  help\n\nCards live in DIR/cards/TAG/HUID/CARD.md (DIR defaults to .).\n")
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
		return addCommand(*root, args[1:], out, errOut)
	case "list":
		return listCommand(*root, args[1:], out, errOut)
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

func addCommand(root string, args []string, out, errOut io.Writer) int {
	fs := commandFlags("add", "-t TAG -q QUESTION -a ANSWER", errOut)
	tag := fs.String("t", "", "card tag")
	question := fs.String("q", "", "question text")
	answer := fs.String("a", "", "answer text")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 0 || !validTag(*tag) || strings.TrimSpace(*question) == "" || strings.TrimSpace(*answer) == "" {
		fmt.Fprintln(errOut, "gostudy: add requires a safe tag, a question, and an answer")
		fs.Usage()
		return 2
	}
	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return commandError(errOut, err)
	}
	id, err := huid.New(time.Now(), hex.EncodeToString(suffix[:]))
	if err != nil {
		return commandError(errOut, err)
	}
	path := filepath.Join(root, "cards", *tag, id.String())
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return commandError(errOut, err)
	}
	if err := os.Mkdir(path, 0755); err != nil {
		return commandError(errOut, err)
	}
	file, err := os.OpenFile(filepath.Join(path, "CARD.md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return commandError(errOut, err)
	}
	content := fmt.Sprintf("---\nid: %s\ntags:\n  - %s\n---\n\n## Question\n\n%s\n\n## Answer\n\n%s\n", yamlString(id.String()), yamlString(*tag), strings.TrimSpace(*question), strings.TrimSpace(*answer))
	if _, err := io.WriteString(file, content); err != nil {
		file.Close()
		return commandError(errOut, err)
	}
	if err := file.Close(); err != nil {
		return commandError(errOut, err)
	}
	fmt.Fprintln(out, id)
	return 0
}

// JSON strings use YAML-compatible double quoting and escape special characters.
func yamlString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func validTag(tag string) bool {
	if tag == "" || tag == "." || tag == ".." {
		return false
	}
	hasNonDigit := false
	for _, c := range tag {
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '_' || c == '-') {
			return false
		}
		if c < '0' || c > '9' {
			hasNonDigit = true
		}
	}
	return hasNonDigit
}

type cardEntry struct{ id, tag string }

func cardEntries(root, filter string) ([]cardEntry, error) {
	base := filepath.Join(root, "cards")
	tags, err := os.ReadDir(base)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []cardEntry
	for _, tag := range tags {
		if !tag.IsDir() || !validTag(tag.Name()) || filter != "" && tag.Name() != filter {
			continue
		}
		ids, err := os.ReadDir(filepath.Join(base, tag.Name()))
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if !id.IsDir() || !huid.Valid(id.Name()) {
				continue
			}
			info, err := os.Stat(filepath.Join(base, tag.Name(), id.Name(), "CARD.md"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if info.Mode().IsRegular() {
				entries = append(entries, cardEntry{id.Name(), tag.Name()})
			}
		}
	}
	return entries, nil
}

func listCommand(root string, args []string, out, errOut io.Writer) int {
	fs := commandFlags("list", "[-t TAG]", errOut)
	tag := fs.String("t", "", "show only this tag")
	if err := fs.Parse(args); err != nil {
		return flagStatus(err)
	}
	if fs.NArg() != 0 || *tag != "" && !validTag(*tag) {
		fs.Usage()
		return 2
	}
	entries, err := cardEntries(root, *tag)
	if err != nil {
		return commandError(errOut, err)
	}
	for _, entry := range entries {
		fmt.Fprintf(out, "%s\t%s\n", entry.id, entry.tag)
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
	entries, err := cardEntries(root, "")
	if err != nil {
		return commandError(errOut, err)
	}
	for _, entry := range entries {
		if entry.id == fs.Arg(0) {
			content, err := os.ReadFile(filepath.Join(root, "cards", entry.tag, entry.id, "CARD.md"))
			if err != nil {
				return commandError(errOut, err)
			}
			if _, err := out.Write(content); err != nil {
				return commandError(errOut, err)
			}
			return 0
		}
	}
	fmt.Fprintf(errOut, "gostudy: card %s not found\n", fs.Arg(0))
	return 1
}

func commandError(errOut io.Writer, err error) int {
	fmt.Fprintf(errOut, "gostudy: %v\n", err)
	return 1
}
