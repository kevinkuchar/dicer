package main

import "dicer/pkg/stack"

/*************************************
* Turn
*************************************/
type Turn struct {
	round          int
	dice           []Dice
	result         int
	expression     string
	removedAilment bool
	lostLife       bool
	stack          *stack.ArrayStack[GameState]
}

func CreateTurn(round int) *Turn {
	turn := &Turn{
		round: round,
		stack: CreateTurnStack(),
	}

	return turn
}

func CreateTurnStack() *stack.ArrayStack[GameState] {
	stack := stack.CreateArrayStack[GameState]()

	stack.Push(GS_ResultsPhase)
	stack.Push(GS_ExpressionPhase)
	stack.Push(GS_RollPhase)
	stack.Push(GS_TurnStart)

	return stack
}

func (t *Turn) RollDice() {
	dice := make([]Dice, NUM_DICE)
	for i := range dice {
		dice[i] = CreateAndRollDie()
	}

	t.dice = dice
}

func (t *Turn) RollSelectedDice(selected map[int]struct{}) {
	for i := range selected {
		t.dice[i].Roll()
	}
}

func (t *Turn) ApplyResult(player *Player) {
	if player.Ailments.HasAilment(t.result) {
		player.Ailments.RemoveAilment(t.result)
		t.removedAilment = true
	} else {
		player.RemoveLife()
		t.lostLife = true
	}
}
