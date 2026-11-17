package schedule

import (
	"math"
	"sync"
	"testing"
)

// solvedPolicy shares one solve across tests because solving takes a while.
var solvedPolicy = sync.OnceValue(solvePolicy)

func TestReachingTheGoalCostsNothing(t *testing.T) {
	if c := solvedPolicy().expectedCost(366, 5); c != 0 {
		t.Fatalf("expected cost at the goal half-life = %v, want 0", c)
	}
}

func TestStrongerMemoryCostsLessToFinish(t *testing.T) {
	weak := solvedPolicy().expectedCost(2, 5)
	strong := solvedPolicy().expectedCost(20, 5)
	if strong >= weak {
		t.Fatalf("expected cost with a 20-day half-life = %v, want less than with a 2-day half-life (%v)", strong, weak)
	}
}

func TestHarderCardCostsMoreToFinish(t *testing.T) {
	easy := solvedPolicy().expectedCost(5, 5)
	hard := solvedPolicy().expectedCost(5, 10)
	if hard <= easy {
		t.Fatalf("expected cost at difficulty 10 = %v, want more than at difficulty 5 (%v)", hard, easy)
	}
}

func TestEveryIntervalIsAtLeastOneDay(t *testing.T) {
	for d := 1; d <= MaxDifficulty; d++ {
		for i := range gridSize {
			if n := solvedPolicy().interval(gridHalfLife(i), d); n < 1 {
				t.Fatalf("interval at half-life %v and difficulty %d = %d, want at least 1", gridHalfLife(i), d, n)
			}
		}
	}
}

func TestCardPastTheGoalIsReviewedAtNinetyPercentRecall(t *testing.T) {
	// round(500 * log2(1/0.9)) = 76 days.
	if n := solvedPolicy().interval(500, 5); n != 76 {
		t.Fatalf("interval at a 500-day half-life = %d, want 76", n)
	}
}

func TestIntervalsMatchAnIndependentSolve(t *testing.T) {
	p := solvedPolicy()
	if p.interval(9.9, 5) != 2 || p.interval(30.4, 5) != 5 || p.interval(98.1, 5) != 12 {
		t.Fatalf("difficulty 5 intervals at 9.9, 30.4, 98.1 days = %d, %d, %d, want 2, 5, 12",
			p.interval(9.9, 5), p.interval(30.4, 5), p.interval(98.1, 5))
	}
	if p.interval(9.9, 1) != 2 || p.interval(9.9, 10) != 2 || p.interval(9.9, 18) != 3 {
		t.Fatalf("intervals at 9.9 days for difficulties 1, 10, 18 = %d, %d, %d, want 2, 2, 3",
			p.interval(9.9, 1), p.interval(9.9, 10), p.interval(9.9, 18))
	}
}

func TestStartingCostsMatchAnIndependentSolve(t *testing.T) {
	p := solvedPolicy()
	for d, want := range map[int]float64{1: 15.6, 5: 33.0, 10: 51.5, 18: 72.1} {
		if c := p.expectedCost(initialHalfLife(d), d); math.Abs(c-want) > 0.05 {
			t.Fatalf("expected cost of a new card at difficulty %d = %v, want about %v", d, c, want)
		}
	}
}
