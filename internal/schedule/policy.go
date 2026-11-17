package schedule

import "math"

// Half-lives are solved on a grid of gridBase^k days for k from gridLowestPower
// to gridHighestPower, about 0.23 to 366 days. The last grid point is the
// goal: a card whose half-life reaches it needs no further cost counted.
// See the credit notice in memory.go for where these values come from.
const (
	gridBase         = 1.05
	gridLowestPower  = -30
	gridHighestPower = 121
	gridSize         = gridHighestPower - gridLowestPower + 1
	goalIndex        = gridSize - 1
)

// Solver parameters. A recall costs recallCost seconds and a lapse
// forgetCost seconds. Intervals are never long enough to let recall fall
// below lowestRecallProbability.
const (
	recallCost              = 3
	forgetCost              = 9
	lowestRecallProbability = 0.3
)

// pastGoalRecallProbability sets the interval for a card at or past the goal,
// where the solved problem has ended but the card is still reviewed. This is
// gostudy's own rule, not part of the paper or the reference code.
const pastGoalRecallProbability = 0.9

// The solver starts every state at unsolvedCost, which is higher than any
// real cost, and stops sweeping once no cost drops by more than
// convergedCostChange seconds.
const (
	unsolvedCost        = 2000
	convergedCostChange = 1e-6
)

// A policy holds, for each difficulty and half-life grid point, the expected
// review cost in seconds until the goal and the interval in days that
// achieves it. Index 0 of the difficulty dimension is unused.
type policy struct {
	cost      [MaxDifficulty + 1][gridSize]float64
	intervals [MaxDifficulty + 1][gridSize]int
}

// gridIndex is the grid point nearest to halfLife in log space.
func gridIndex(halfLife float64) int {
	k := int(math.Round(math.Log(halfLife) / math.Log(gridBase)))
	return min(max(k-gridLowestPower, 0), goalIndex)
}

// gridHalfLife is the half-life in days at a grid point.
func gridHalfLife(index int) float64 {
	return math.Pow(gridBase, float64(index+gridLowestPower))
}

// solvePolicy finds, by value iteration, the interval that minimises the
// expected review cost from every memory state until the goal half-life.
// A lapse moves a card to a harder difficulty whose costs it reads, so
// difficulties are solved from the hardest down.
func solvePolicy() *policy {
	p := &policy{}
	for d := MaxDifficulty; d >= 1; d-- {
		for i := range goalIndex {
			p.cost[d][i] = unsolvedCost
		}
		for p.sweep(d) > convergedCostChange {
		}
	}
	return p
}

// sweep lowers each cost at difficulty d to the cheapest interval found from
// the current costs, and returns the largest drop.
func (p *policy) sweep(d int) float64 {
	largestDrop := 0.0
	for i := range goalIndex {
		h := gridHalfLife(i)
		longest := max(1, int(math.Round(h*math.Log2(1/lowestRecallProbability))))
		for t := 1; t <= longest; t++ {
			c := p.costOfReviewingAfter(h, d, float64(t))
			if c < p.cost[d][i] {
				largestDrop = max(largestDrop, p.cost[d][i]-c)
				p.cost[d][i] = c
				p.intervals[d][i] = t
			}
		}
	}
	return largestDrop
}

// costOfReviewingAfter is the expected cost from state (h, d) of reviewing
// after t days and then following the current costs.
func (p *policy) costOfReviewingAfter(h float64, d int, t float64) float64 {
	recall := recallProbability(h, t)
	recalled := recallCost + p.expectedCost(halfLifeAfterRecall(h, d, recall), d)
	forgotten := forgetCost + p.expectedCost(halfLifeAfterForgetting(h, d, recall), difficultyAfterForgetting(d))
	return recall*recalled + (1-recall)*forgotten
}

// expectedCost is the expected review cost in seconds from a memory state
// until its half-life reaches the goal.
func (p *policy) expectedCost(halfLife float64, difficulty int) float64 {
	return p.cost[difficulty][gridIndex(halfLife)]
}

// interval is the number of days to wait before reviewing a card in the given
// memory state. A card at or past the goal waits until its recall probability
// falls to pastGoalRecallProbability.
func (p *policy) interval(halfLife float64, difficulty int) int {
	i := gridIndex(halfLife)
	if i == goalIndex {
		return max(1, int(math.Round(halfLife*math.Log2(1/pastGoalRecallProbability))))
	}
	return p.intervals[difficulty][i]
}
