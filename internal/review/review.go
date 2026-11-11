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

type dueCard struct {
	card     card.Card
	progress schedule.Schedule
}

// Run reviews due cards, saving each grade before continuing to the next card.
// Quitting or reaching EOF leaves the current ungraded card unchanged.
func Run(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time) error {
	terminal := false
	if file, ok := out.(*os.File); ok && os.Getenv("TERM") != "dumb" {
		terminal = term.IsTerminal(int(file.Fd()))
	}
	if file, ok := in.(*os.File); !ok || !term.IsTerminal(int(file.Fd())) {
		terminal = false
	}
	return run(store, tag, in, out, now, terminal)
}

func run(store cardstore.Store, tag string, in io.Reader, out io.Writer, now func() time.Time, terminal bool) error {
	entries, err := store.List(tag)
	if err != nil {
		return err
	}
	cutoff := now().UTC()
	var due []dueCard
	for _, entry := range entries {
		content, progress, err := store.Load(entry.ID)
		if err != nil {
			return err
		}
		if progress.LastReviewed == nil || !progress.LastReviewed.UTC().AddDate(0, 0, progress.Interval).After(cutoff) {
			due = append(due, dueCard{content, progress})
		}
	}
	if len(due) == 0 {
		_, err := fmt.Fprintln(out, "No cards are due.")
		return err
	}
	if terminal {
		return runScreen(store, due, in, out, now)
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
		_, err := fmt.Fprintf(out, "\n%s. Reviewed %d %s. %d remaining in this session.\n", status, reviewed, noun, len(due)-reviewed)
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
			input, ok, err := read("Grade 0–5 (0–2 forgotten, 3 hard, 4 good, 5 easy; q to quit): ")
			if err != nil {
				return err
			}
			if !ok || input == "q" {
				return finish(false)
			}
			grade, err := strconv.Atoi(input)
			if err != nil || grade < 0 || grade > 5 {
				if _, err := fmt.Fprintln(out, "Enter a grade from 0 to 5."); err != nil {
					return err
				}
				continue
			}
			progress, err := schedule.Review(item.progress, grade, now().UTC())
			if err != nil {
				return err
			}
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
