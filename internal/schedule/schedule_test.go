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
