package models

import (
	"dicer/pkg/config"
	"dicer/pkg/stack"
)

/*************************************
* Turn
*************************************/
/*************************************
* Game State Flow
*************************************/
type TurnPhase int

const (
	GS_TurnStart       TurnPhase = iota
	GS_RollPhase                 // Re-roll selection phase
	GS_ExpressionPhase           // Gather user input for expression
	GS_ResultsPhase              // Show results of expression
	GS_GameOver                  // Game over state
)

type Turn struct {
	Round          int
	Dice           []Dice
	Result         int
	Expression     string
	RemovedAilment bool
	LostLife       bool
	Stack          *stack.ArrayStack[TurnPhase]
}

func CreateTurn(round int) *Turn {
	turn := &Turn{
		Round: round,
		Stack: CreateTurnStack(),
	}

	return turn
}

func CreateTurnStack() *stack.ArrayStack[TurnPhase] {
	stack := stack.CreateArrayStack[TurnPhase]()

	stack.Push(GS_ResultsPhase)
	stack.Push(GS_ExpressionPhase)
	stack.Push(GS_RollPhase)
	stack.Push(GS_TurnStart)

	return stack
}

func (t *Turn) RollDice() {
	dice := make([]Dice, config.NumDice)
	for i := range dice {
		dice[i] = CreateAndRollDie()
	}

	t.Dice = dice
}

func (t *Turn) RollSelectedDice(selected map[int]struct{}) {
	for i := range selected {
		t.Dice[i].Roll()
	}
}

func (t *Turn) ApplyResult(player *Player) {
	if player.Ailments.HasAilment(t.Result) {
		player.Ailments.RemoveAilment(t.Result)
		t.RemovedAilment = true
	} else {
		player.RemoveLife()
		t.LostLife = true
	}
}
