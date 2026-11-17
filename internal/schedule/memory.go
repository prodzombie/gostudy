package schedule

import "math"

// The memory model and its parameter values come from Ye, Su and Cao, "A
// Stochastic Shortest Path Algorithm for Optimizing Spaced Repetition
// Scheduling" (KDD 2022). The parameter values are taken from the
// maimemo/SSP-MMC repository, MIT licence, Copyright (c) 2022 MaiMemo, Inc.
//
// h is the half-life in days, d the difficulty, p the recall probability.
const (
	// h' = h * (1 + e^recallScale * d^recallDifficultyExp * h^recallHalfLifeExp * (1-p)^recallForgettingExp)
	recallScale         = 3.81
	recallDifficultyExp = -0.534
	recallHalfLifeExp   = -0.127
	recallForgettingExp = 0.97

	// h' = e^forgetScale * d^forgetDifficultyExp * h^forgetHalfLifeExp * (1-p)^forgetForgettingExp
	forgetScale         = -0.041
	forgetDifficultyExp = -0.041
	forgetHalfLifeExp   = 0.377
	forgetForgettingExp = -0.227

	// Forgetting raises d by forgetDifficultyStep, up to MaxDifficulty.
	forgetDifficultyStep = 2

	// h0 = -1 / log2(max(initialRecallBase - initialRecallStep*d, initialRecallFloor))
	initialRecallBase  = 0.925
	initialRecallStep  = 0.05
	initialRecallFloor = 0.025
)

// initialDifficulty is the difficulty every new card starts with, giving a
// starting half-life of about 1.76 days. This fixed value is gostudy's own
// choice, not part of the paper or the reference code.
const initialDifficulty = 5

// MaxDifficulty is the hardest a card can be. Difficulty runs from 1 to
// MaxDifficulty.
const MaxDifficulty = 18

// recallProbability is the chance of recalling a card elapsedDays after its
// last review, given its memory half-life in days: p = 2^(-t/h).
func recallProbability(halfLife, elapsedDays float64) float64 {
	return math.Exp2(-elapsedDays / halfLife)
}

// halfLifeAfterRecall is the half-life after a successful review that had
// recall probability p. A recall that was less likely strengthens memory more.
func halfLifeAfterRecall(halfLife float64, difficulty int, p float64) float64 {
	growth := math.Exp(recallScale) *
		math.Pow(float64(difficulty), recallDifficultyExp) *
		math.Pow(halfLife, recallHalfLifeExp) *
		math.Pow(1-p, recallForgettingExp)
	return halfLife * (1 + growth)
}

// halfLifeAfterForgetting is the half-life after a failed review that had
// recall probability p.
func halfLifeAfterForgetting(halfLife float64, difficulty int, p float64) float64 {
	return math.Exp(forgetScale) *
		math.Pow(float64(difficulty), forgetDifficultyExp) *
		math.Pow(halfLife, forgetHalfLifeExp) *
		math.Pow(1-p, forgetForgettingExp)
}

// difficultyAfterForgetting is the difficulty after a failed review.
// A recall leaves difficulty unchanged.
func difficultyAfterForgetting(difficulty int) int {
	return min(difficulty+forgetDifficultyStep, MaxDifficulty)
}

// initialHalfLife is the half-life of a card of the given difficulty before
// its first review. Harder cards start with a shorter half-life.
func initialHalfLife(difficulty int) float64 {
	p := max(initialRecallBase-initialRecallStep*float64(difficulty), initialRecallFloor)
	return -1 / math.Log2(p)
}
