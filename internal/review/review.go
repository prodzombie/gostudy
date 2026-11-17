// Package review runs interactive study sessions using the SM-2 scheduler.
package review

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/prodzombie/gostudy/internal/card"
	"github.com/prodzombie/gostudy/internal/cardstore"
	"github.com/prodzombie/gostudy/internal/schedule"
)

// Options controls whether a session records scheduled reviews.
type Options struct{ Practice bool }

type dueCard struct {
	card     card.Card
	progress schedule.Schedule
}

// Run reviews due cards, saving each grade before continuing to the next card.
// Quitting or reaching EOF leaves the current ungraded card unchanged.
func Run(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time) error {
	return RunWithOptions(store, tag, in, out, now, Options{})
}

// RunWithOptions runs either scheduled review or practice of all matching cards.
func RunWithOptions(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time, options Options) error {
	terminal := false
	if file, ok := out.(*os.File); ok && os.Getenv("TERM") != "dumb" {
		terminal = term.IsTerminal(int(file.Fd()))
	}
	if file, ok := in.(*os.File); !ok || !term.IsTerminal(int(file.Fd())) {
		terminal = false
	}
	return runSession(store, tag, in, out, now, terminal, options)
}

func run(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time, terminal bool) error {
	return runSession(store, tag, in, out, now, terminal, Options{})
}

func runSession(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time, terminal bool, options Options) error {
	entries, err := store.List(tag)
	if err != nil {
		return err
	}
	current := now()
	var due []dueCard
	for _, entry := range entries {
		content, progress, err := store.Load(entry.ID)
		if err != nil {
			return err
		}
		if options.Practice || progress.Due(current) {
			due = append(due, dueCard{content, progress})
		}
	}
	if len(due) == 0 {
		message := "No cards are due."
		if options.Practice {
			message = "No matching cards found."
		}
		_, err := fmt.Fprintln(out, message)
		return err
	}
	if terminal {
		return runScreen(store, due, in, out, now, options)
	}

	if options.Practice {
		if _, err := fmt.Fprintln(out, "Practice · no progress will be saved."); err != nil {
			return err
		}
	}
	scanner := bufio.NewScanner(in)
	read := func(prompt string) (string, bool, error) {
		if _, err := fmt.Fprint(out, prompt); err != nil {
			return "", false, err
		}
		if !scanner.Scan() {
			return "", false, scanner.Err()
		}
		return strings.TrimSpace(scanner.Text()), true, nil
	}
	reviewed := 0
	finish := func(completed bool) error {
		noun := "cards"
		if reviewed == 1 {
			noun = "card"
		}
		status := "Session stopped"
		if completed {
			status = "Session complete"
		}
		verb := "Reviewed"
		if options.Practice {
			verb = "Practiced"
		}
		_, err := fmt.Fprintf(out, "\n%s. %s %d %s. %d remaining in this session.\n", status, verb, reviewed, noun, len(due)-reviewed)
		return err
	}
	confirmation := ""
	for index, item := range due {
		if err := showCard(out, item.card, false, index+1, len(due)); err != nil {
			return err
		}
		for {
			input, ok, err := read("Press Enter to reveal the answer (q to quit): ")
			if err != nil {
				return err
			}
			if !ok || input == "q" {
				return finish(false)
			}
			if input == "" {
				break
			}
		}
		if err := showCard(out, item.card, true, index+1, len(due)); err != nil {
			return err
		}
		for {
			prompt := "Grade 0–5 (0–2 forgotten, 3 hard, 4 good, 5 easy; q to quit): "
			if options.Practice {
				prompt = "Press Enter for the next card (q to quit): "
			}
			input, ok, err := read(prompt)
			if err != nil {
				return err
			}
			if !ok || input == "q" {
				return finish(false)
			}
			if options.Practice {
				if input != "" {
					continue
				}
				reviewed++
				break
			}
			grade, err := strconv.Atoi(input)
			if err != nil || grade < 0 || grade > 5 {
				if _, err := fmt.Fprintln(out, "Enter a grade from 0 to 5."); err != nil {
					return err
				}
				continue
			}
			progress := schedule.Review(item.progress, grade >= 3, now())
			if err := store.SaveSchedule(item.card.ID.String(), progress); err != nil {
				return err
			}
			reviewed++
			unit := "days"
			if progress.Interval == 1 {
				unit = "day"
			}
			confirmation = fmt.Sprintf("Previous card saved · due in %d %s", progress.Interval, unit)
			if _, err := fmt.Fprintln(out, confirmation); err != nil {
				return err
			}
			break
		}
	}
	return finish(true)
}
