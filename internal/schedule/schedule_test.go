package schedule

import (
	"math"
	"testing"
	"time"
)

func TestReviewProgressionAndForgottenCard(t *testing.T) {
	now := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	s := Schedule{}

	for _, wantInterval := range []int{1, 6, 16} {
		var err error
		s, err = Review(s, 5, now)
		if err != nil {
			t.Fatal(err)
		}
		if s.Interval != wantInterval {
			t.Fatalf("interval = %d, want %d", s.Interval, wantInterval)
		}
		now = now.Add(24 * time.Hour)
	}
	if s.Repetitions != 3 || math.Abs(s.EaseFactor-2.8) > 1e-9 {
		t.Fatalf("after three successful reviews: repetitions = %d, ease = %v", s.Repetitions, s.EaseFactor)
	}

	s, err := Review(s, 2, now)
	if err != nil {
		t.Fatal(err)
	}
	if s.Repetitions != 0 || s.Interval != 1 || math.Abs(s.EaseFactor-2.8) > 1e-9 {
		t.Fatalf("after a forgotten review: %+v", s)
	}
	if s.LastReviewed == nil || !s.LastReviewed.Equal(now) {
		t.Fatalf("last reviewed = %v, want %v", s.LastReviewed, now)
	}
}

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

	if s.Due(reviewed.AddDate(0, 0, 1).Add(-time.Nanosecond)) {
		t.Fatal("a card should not be due before its interval has passed")
	}
}

func TestReviewRejectsInvalidQuality(t *testing.T) {
	for _, quality := range []int{-1, 6} {
		original := Schedule{EaseFactor: 2.5, Interval: 6, Repetitions: 2}
		got, err := Review(original, quality, time.Now())
		if err == nil {
			t.Fatalf("quality %d: expected an error", quality)
		}
		if got != original {
			t.Fatalf("quality %d: schedule changed: %+v", quality, got)
		}
	}
}
