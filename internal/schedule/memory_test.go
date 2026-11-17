package schedule

import (
	"math"
	"testing"
)

func TestRecallProbabilityIsHalfAtTheHalfLife(t *testing.T) {
	if p := recallProbability(3, 3); math.Abs(p-0.5) > 1e-12 {
		t.Fatalf("recall probability after one half-life = %v, want 0.5", p)
	}
}

func TestLongerWaitLowersRecallProbability(t *testing.T) {
	if recallProbability(3, 2) <= recallProbability(3, 4) {
		t.Fatal("waiting longer should lower the recall probability")
	}
}

func TestRecallAfterALongerWaitGrowsHalfLifeMore(t *testing.T) {
	early := halfLifeAfterRecall(2, 5, recallProbability(2, 1))
	late := halfLifeAfterRecall(2, 5, recallProbability(2, 4))
	if late <= early {
		t.Fatalf("half-life after a late recall = %v, want more than after an early recall (%v)", late, early)
	}
}

func TestRecallGrowsHalfLifeByThePublishedParameters(t *testing.T) {
	// 2 * (1 + e^3.81 * 5^-0.534 * 2^-0.127 * 0.5^0.97), computed independently.
	if h := halfLifeAfterRecall(2, 5, 0.5); math.Abs(h-19.873602044944235) > 1e-9 {
		t.Fatalf("half-life after recall = %v, want 19.8736...", h)
	}
}

func TestForgettingAfterAShortWaitKeepsALongerHalfLife(t *testing.T) {
	early := halfLifeAfterForgetting(2, 5, recallProbability(2, 1))
	late := halfLifeAfterForgetting(2, 5, recallProbability(2, 4))
	if early <= late {
		t.Fatalf("half-life after forgetting early = %v, want more than after forgetting late (%v)", early, late)
	}
}

func TestForgettingResetsHalfLifeByThePublishedParameters(t *testing.T) {
	// e^-0.041 * 5^-0.041 * 2^0.377 * 0.5^-0.227, computed independently.
	if h := halfLifeAfterForgetting(2, 5, 0.5); math.Abs(h-1.3657093208607873) > 1e-9 {
		t.Fatalf("half-life after forgetting = %v, want 1.3657...", h)
	}
}

func TestForgettingRaisesDifficultyByTwo(t *testing.T) {
	if d := difficultyAfterForgetting(5); d != 7 {
		t.Fatalf("difficulty after forgetting = %d, want 7", d)
	}
}

func TestDifficultyStopsAtItsMaximum(t *testing.T) {
	if d := difficultyAfterForgetting(17); d != 18 {
		t.Fatalf("difficulty after forgetting = %d, want the maximum 18", d)
	}
}

func TestHarderCardsStartWithAShorterHalfLife(t *testing.T) {
	if initialHalfLife(6) >= initialHalfLife(5) {
		t.Fatal("a harder card should start with a shorter half-life")
	}
}

func TestNewCardsStartWithAHalfLifeOfAboutOneAndThreeQuarterDays(t *testing.T) {
	// -1 / log2(0.925 - 0.05*5), computed independently.
	if h := initialHalfLife(initialDifficulty); math.Abs(h-1.7635421746374427) > 1e-9 {
		t.Fatalf("starting half-life of a new card = %v, want 1.7635...", h)
	}
}
