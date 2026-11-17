package schedule

import (
	"time"

	"github.com/prodzombie/gostudy/internal/calendar"
	"github.com/prodzombie/gostudy/internal/huid"
)

// Schedule tracks a card's memory state and when it is next due.
// A nil LastReviewed means the card has not been reviewed yet.
type Schedule struct {
	Card         huid.HUID
	HalfLife     float64 // days until recall probability halves
	Difficulty   int     // 1 to MaxDifficulty
	Interval     int     // days from LastReviewed until the card is due
	LastReviewed *time.Time
}

// Due reports whether the card should be reviewed on now's calendar day.
// The review time is read as a day in now's location, so a card becomes due
// at local midnight. A card that has never been reviewed is always due.
func (s Schedule) Due(now time.Time) bool {
	if s.LastReviewed == nil {
		return true
	}
	reviewed := calendar.DateOf(s.LastReviewed.In(now.Location()))
	return !reviewed.AddDays(s.Interval).After(calendar.DateOf(now))
}

// Review applies one review, recalled or forgotten, at now and schedules the
// next one with the solved SSP-MMC policy.
//
// A card's first review is the exposure that starts its memory: whatever the
// answer, the card enters the model at initialDifficulty and its starting
// half-life. gostudy fixes the starting difficulty rather than letting the
// first answer choose it.
//
// Later reviews read the recall probability from the whole calendar days
// actually elapsed since the last review, in now's location (at least one),
// not from the scheduled interval, so a late review that is still recalled
// strengthens memory more. A recall lengthens the half-life; a lapse resets it
// and makes the card harder.
func Review(s Schedule, recalled bool, now time.Time) Schedule {
	if s.LastReviewed == nil {
		s.Difficulty = initialDifficulty
		s.HalfLife = initialHalfLife(initialDifficulty)
	} else {
		reviewed := calendar.DateOf(s.LastReviewed.In(now.Location()))
		// At zero elapsed days recall is certain and the forgetting formula is infinite.
		elapsed := max(calendar.DateOf(now).DaysSince(reviewed), 1)
		p := recallProbability(s.HalfLife, float64(elapsed))
		if recalled {
			s.HalfLife = halfLifeAfterRecall(s.HalfLife, s.Difficulty, p)
		} else {
			s.HalfLife = halfLifeAfterForgetting(s.HalfLife, s.Difficulty, p)
			s.Difficulty = difficultyAfterForgetting(s.Difficulty)
		}
	}
	s.Interval = solvedPolicy().interval(s.HalfLife, s.Difficulty)
	s.LastReviewed = &now
	return s
}
