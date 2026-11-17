package calendar

import (
	"testing"
	"time"
	_ "time/tzdata" // Load Europe/Rome regardless of the machine's zone database.
)

func rome(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	return location
}

func TestDateIgnoresTimeOfDay(t *testing.T) {
	morning := time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC)
	night := time.Date(2026, time.November, 17, 23, 59, 59, 0, time.UTC)

	if DateOf(morning) != DateOf(night) {
		t.Fatalf("morning %v and night %v should be the same date", DateOf(morning), DateOf(night))
	}
}

func TestDateIsTheCalendarDayInTheTimesOwnLocation(t *testing.T) {
	// 00:30 in Rome is still the previous day in UTC.
	afterMidnightInRome := time.Date(2026, time.November, 18, 0, 30, 0, 0, rome(t))

	if got := DateOf(afterMidnightInRome).String(); got != "2026-11-18" {
		t.Fatalf("date in Rome = %s, want 2026-11-18", got)
	}
	if got := DateOf(afterMidnightInRome.UTC()).String(); got != "2026-11-17" {
		t.Fatalf("date in UTC = %s, want 2026-11-17", got)
	}
}

func TestAddingDaysCrossesMonthEnd(t *testing.T) {
	lastOfNovember := DateOf(time.Date(2026, time.November, 30, 12, 0, 0, 0, time.UTC))

	if got := lastOfNovember.AddDays(1).String(); got != "2026-12-01" {
		t.Fatalf("day after %v = %s, want 2026-12-01", lastOfNovember, got)
	}
}

func TestAddingDaysIgnoresDaylightSavingChange(t *testing.T) {
	// Rome's clocks go back on 25 October 2026, so that local day lasts 25 hours.
	dayBeforeChange := DateOf(time.Date(2026, time.October, 24, 23, 30, 0, 0, rome(t)))

	if got := dayBeforeChange.AddDays(2).String(); got != "2026-10-26" {
		t.Fatalf("two days after %v = %s, want 2026-10-26", dayBeforeChange, got)
	}
}

func TestLaterDateIsAfterEarlierDate(t *testing.T) {
	today := DateOf(time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC))

	if !today.AddDays(1).After(today) {
		t.Fatal("tomorrow should be after today")
	}
	if today.After(today) {
		t.Fatal("a date should not be after itself")
	}
}

func TestDaysSinceTheSameDayIsZero(t *testing.T) {
	morning := DateOf(time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC))
	night := DateOf(time.Date(2026, time.November, 17, 23, 59, 59, 0, time.UTC))

	if got := night.DaysSince(morning); got != 0 {
		t.Fatalf("days between two times on the same day = %d, want 0", got)
	}
}

func TestDaysSinceCountsCalendarDaysAcrossDaylightSavingChange(t *testing.T) {
	// Rome's clocks go back on 25 October 2026; 26 hours pass but only two days.
	before := DateOf(time.Date(2026, time.October, 24, 23, 30, 0, 0, rome(t)))
	after := DateOf(time.Date(2026, time.October, 26, 0, 30, 0, 0, rome(t)))

	if got := after.DaysSince(before); got != 2 {
		t.Fatalf("days from %v to %v = %d, want 2", before, after, got)
	}
}
