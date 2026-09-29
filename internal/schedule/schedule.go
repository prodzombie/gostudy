package schedule

import (
	"fmt"
	"math"
	"time"

	"github.com/emiliopalmerini/gostudy/internal/huid"
)

// Schedule tracks a card's progress through spaced repetition.
// A zero EaseFactor means the card has not been reviewed yet.
type Schedule struct {
	Card         huid.HUID
	EaseFactor   float64
	Interval     int
	LastReviewed *time.Time
	Repetitions  int
}

// Review applies one SM-2 grade to a schedule and records when it happened.
// Grades must be from 0 to 5. A forgotten card restarts at a one-day interval
// without changing its ease factor; successful reviews update the ease factor.
func Review(s Schedule, quality int, now time.Time) (Schedule, error) {
	if quality < 0 || quality > 5 {
		return s, fmt.Errorf("review quality must be between 0 and 5: %d", quality)
	}

	if s.EaseFactor == 0 {
		s.EaseFactor = 2.5
	}

	s.LastReviewed = &now

	if quality < 3 {
		s.Repetitions = 0
		s.Interval = 1

		return s, nil
	}

	s.Repetitions++

	switch s.Repetitions {
	case 1:
		s.Interval = 1
	case 2:
		s.Interval = 6
	default:
		s.Interval = int(math.Round(float64(s.Interval) * s.EaseFactor))
	}

	q := float64(quality)

	s.EaseFactor += 0.1 - (5-q)*(0.08+(5-q)*0.02)

	if s.EaseFactor < 1.3 {
		s.EaseFactor = 1.3
	}

	s.LastReviewed = &now

	return s, nil
}
