// Package review runs interactive study sessions using the SM-2 scheduler.
package review

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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
	finish := func() error {
		noun := "cards"
		if reviewed == 1 {
			noun = "card"
		}
		_, err := fmt.Fprintf(out, "\nReviewed %d %s.\n", reviewed, noun)
		return err
	}
	for _, item := range due {
		if _, err := fmt.Fprintf(out, "\nCard %s [%s]\n\n%s\n\n", item.card.ID, item.card.Tags.Value, item.card.Question); err != nil {
			return err
		}
		for {
			input, ok, err := read("Press Enter to reveal the answer (q to quit): ")
			if err != nil {
				return err
			}
			if !ok || input == "q" {
				return finish()
			}
			if input == "" {
				break
			}
		}
		if _, err := fmt.Fprintf(out, "\n%s\n\n", item.card.Answer); err != nil {
			return err
		}
		for {
			input, ok, err := read("Grade 0–5 (0–2 forgotten, 3 hard, 4 good, 5 easy; q to quit): ")
			if err != nil {
				return err
			}
			if !ok || input == "q" {
				return finish()
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
			if _, err := fmt.Fprintf(out, "Next review: %s.\n", progress.LastReviewed.AddDate(0, 0, progress.Interval).Format(time.RFC3339)); err != nil {
				return err
			}
			break
		}
	}
	return finish()
}
