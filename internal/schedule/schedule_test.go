package schedule

import (
	"math"
	"testing"
	"time"
	_ "time/tzdata" // Load Europe/Rome regardless of the machine's zone database.
)

func TestNeverReviewedCardIsDue(t *testing.T) {
	if !(Schedule{}).Due(time.Now()) {
		t.Fatal("a card that was never reviewed should be due")
	}
}

func TestCardIsDueOnceItsIntervalHasPassed(t *testing.T) {
	reviewed := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	s := Schedule{Interval: 1, LastReviewed: &reviewed}

	if !s.Due(reviewed.AddDate(0, 0, 1)) {
		t.Fatal("a card should be due once its interval has passed")
	}
}

func TestCardIsNotDueBeforeItsIntervalHasPassed(t *testing.T) {
	reviewed := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	s := Schedule{Interval: 1, LastReviewed: &reviewed}

	if s.Due(time.Date(2026, time.November, 17, 23, 59, 59, 0, time.UTC)) {
		t.Fatal("a card should not be due before its interval has passed")
	}
}

func TestCardIsDueAtAnyTimeOnItsDueDay(t *testing.T) {
	rome := loadRome(t)
	reviewed := time.Date(2026, time.November, 17, 22, 30, 0, 0, rome)
	s := Schedule{Interval: 1, LastReviewed: &reviewed}

	if !s.Due(time.Date(2026, time.November, 18, 8, 0, 0, 0, rome)) {
		t.Fatal("a card reviewed yesterday evening should be due this morning")
	}
}

func TestReviewDayIsTheLocalCalendarDay(t *testing.T) {
	rome := loadRome(t)
	// Stored as 23:30 UTC on 17 November, but reviewed on 18 November in Rome.
	reviewed := time.Date(2026, time.November, 18, 0, 30, 0, 0, rome).UTC()
	s := Schedule{Interval: 1, LastReviewed: &reviewed}

	if s.Due(time.Date(2026, time.November, 18, 20, 0, 0, 0, rome)) {
		t.Fatal("a card reviewed after local midnight should not be due again the same local day")
	}
}

func loadRome(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	return location
}

// reviewedSchedule is a card last reviewed at the given time with a 4-day
// half-life at the starting difficulty.
func reviewedSchedule(at time.Time) Schedule {
	return Schedule{HalfLife: 4, Difficulty: initialDifficulty, Interval: 2, LastReviewed: &at}
}

func TestFirstReviewStartsTheCardAtTheStartingHalfLifeWhateverTheAnswer(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	recalled := Review(Schedule{}, true, now)
	forgotten := Review(Schedule{}, false, now)

	if recalled.HalfLife != initialHalfLife(initialDifficulty) || recalled.Difficulty != initialDifficulty {
		t.Fatalf("first recalled review = %+v, want the starting half-life and difficulty", recalled)
	}
	if forgotten.HalfLife != recalled.HalfLife || forgotten.Difficulty != recalled.Difficulty {
		t.Fatalf("first forgotten review = %+v, want the same start as a recalled one (%+v)", forgotten, recalled)
	}
}

func TestReviewSchedulesTheSolvedIntervalForTheNewMemoryState(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	s := Review(reviewedSchedule(now.AddDate(0, 0, -2)), true, now)

	if want := solvedPolicy().interval(s.HalfLife, s.Difficulty); s.Interval != want {
		t.Fatalf("interval = %d, want the solved interval %d", s.Interval, want)
	}
}

func TestRecalledReviewLengthensTheHalfLife(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	before := reviewedSchedule(now.AddDate(0, 0, -2))
	after := Review(before, true, now)

	if after.HalfLife <= before.HalfLife || after.Difficulty != before.Difficulty {
		t.Fatalf("after a recalled review: %+v, want a longer half-life and the same difficulty as %+v", after, before)
	}
}

func TestForgottenReviewRaisesDifficulty(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	before := reviewedSchedule(now.AddDate(0, 0, -2))
	after := Review(before, false, now)

	if after.Difficulty != before.Difficulty+2 {
		t.Fatalf("difficulty after a forgotten review = %d, want %d", after.Difficulty, before.Difficulty+2)
	}
	if want := halfLifeAfterForgetting(4, initialDifficulty, recallProbability(4, 2)); after.HalfLife != want {
		t.Fatalf("half-life after a forgotten review = %v, want %v", after.HalfLife, want)
	}
}

func TestLateRecalledReviewEarnsALongerHalfLifeThanAnOnTimeOne(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	onTime := Review(reviewedSchedule(now.AddDate(0, 0, -2)), true, now)
	late := Review(reviewedSchedule(now.AddDate(0, 0, -6)), true, now)

	if late.HalfLife <= onTime.HalfLife {
		t.Fatalf("half-life after a late recall = %v, want more than after an on-time recall (%v)", late.HalfLife, onTime.HalfLife)
	}
}

func TestSameDaySecondReviewCountsAsOneElapsedDay(t *testing.T) {
	morning := time.Date(2026, time.November, 17, 9, 0, 0, 0, time.UTC)
	evening := time.Date(2026, time.November, 17, 20, 0, 0, 0, time.UTC)
	sameDay := Review(reviewedSchedule(morning), false, evening)
	nextDay := Review(reviewedSchedule(morning), false, morning.AddDate(0, 0, 1))

	if math.IsInf(sameDay.HalfLife, 0) || sameDay.HalfLife != nextDay.HalfLife {
		t.Fatalf("half-life after a same-day review = %v, want the one-day result %v", sameDay.HalfLife, nextDay.HalfLife)
	}
}

func TestElapsedDaysAreLocalCalendarDays(t *testing.T) {
	rome := loadRome(t)
	// 23:30 on 17 November and 00:30 on 19 November in Rome are two local days
	// apart, but only one day apart in UTC (22:30 on the 17th, 23:30 on the 18th).
	reviewed := time.Date(2026, time.November, 17, 23, 30, 0, 0, rome).UTC()
	now := time.Date(2026, time.November, 19, 0, 30, 0, 0, rome)
	got := Review(reviewedSchedule(reviewed), true, now)

	if want := halfLifeAfterRecall(4, initialDifficulty, recallProbability(4, 2)); got.HalfLife != want {
		t.Fatalf("half-life = %v, want %v from two elapsed local days", got.HalfLife, want)
	}
}

func TestReviewRecordsTheReviewTime(t *testing.T) {
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)
	s := Review(reviewedSchedule(now.AddDate(0, 0, -2)), true, now)

	if s.LastReviewed == nil || !s.LastReviewed.Equal(now) {
		t.Fatalf("last reviewed = %v, want %v", s.LastReviewed, now)
	}
}
