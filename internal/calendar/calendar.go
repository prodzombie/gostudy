// Package calendar converts instants into calendar days.
package calendar

import "time"

// Date is a calendar day with no time of day or location.
// It is held as UTC midnight, so day arithmetic ignores daylight-saving
// changes and equal days compare equal with ==.
type Date struct{ t time.Time }

// DateOf returns the calendar day of t in t's own location.
func DateOf(t time.Time) Date {
	year, month, day := t.Date()
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// AddDays returns the date n calendar days later.
func (d Date) AddDays(n int) Date {
	return Date{d.t.AddDate(0, 0, n)}
}

// After reports whether d is a later day than other.
func (d Date) After(other Date) bool {
	return d.t.After(other.t)
}

// DaysSince returns the number of calendar days from earlier to d. It is
// negative when earlier is the later date.
func (d Date) DaysSince(earlier Date) int {
	// Both dates are UTC midnights, so the difference is a whole number of days.
	return int(d.t.Sub(earlier.t) / (24 * time.Hour))
}

// String formats the date as YYYY-MM-DD.
func (d Date) String() string {
	return d.t.Format(time.DateOnly)
}
