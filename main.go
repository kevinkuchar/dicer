package main

import (
	"bufio"
	"dicer/pkg/logger"
	"dicer/pkg/math"
	"dicer/pkg/stack"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

/*************************************
* Game State Flow
*************************************/
type GameState int

const (
	GS_TurnStart          GameState = iota
	GS_RollDice                     // Shows prompt to let begin roll
	GS_ViewRoll                     // Outputs roll for viewing
	GS_DiceChoice                   // User selects what to do with each dice
	GS_ViewFinalRoll                // Outputs final roll for viewing
	GS_TypeExpression               // Type expression into terminal
	GS_EvaluateExpression           // Evaluate expression
	GS_ApplyResult                  // Applies result against ailments and lives
	GS_ShowTurnResult               // Shows player status as result of turn
	GS_TurnEnd                      // Yee
)

func CreateTurnStack() *stack.ArrayStack[GameState] {
	stack := stack.CreateArrayStack[GameState]()

	stack.Push(GS_TurnEnd)
	stack.Push(GS_ShowTurnResult)
	stack.Push(GS_ApplyResult)
	stack.Push(GS_EvaluateExpression)
	stack.Push(GS_TypeExpression)
	stack.Push(GS_ViewFinalRoll)
	stack.Push(GS_DiceChoice)
	stack.Push(GS_ViewRoll)
	stack.Push(GS_RollDice)
	stack.Push(GS_TurnStart)

	return stack
}

/*************************************
* Game Configuration
*************************************/
const MAX_LIVES = 3
const NUM_AILMENTS = 9
const NUM_DICE = 3
const REMOVED_AILMENT_VALUE = -1

/*************************************
* Player
*************************************/
type Player struct {
	Lives    int
	Ailments *Ailments
}

func CreatePlayer() Player {
	var player *Player = &Player{Lives: MAX_LIVES}
	player.Ailments = CreateAilments()

	return *player
}

func (player *Player) HasLives() bool {
	return player.Lives > 0
}

/*************************************
* Ailments
*************************************/
type Ailments struct {
	remaining []int
}

func CreateAilments() *Ailments {
	slice := make([]int, NUM_AILMENTS)

	for i := range slice {
		slice[i] = i + 1
	}

	ailments := &Ailments{remaining: slice}
	return ailments
}

func (ailments *Ailments) HasAilments() bool {
	for a := range ailments.remaining {
		if ailments.remaining[a] != REMOVED_AILMENT_VALUE {
			return true
		}
	}
	return false
}

func (ailments *Ailments) HasAilment(num int) bool {
	if len(ailments.remaining) < num || num < 1 {
		return false
	}
	return ailments.remaining[num-1] != REMOVED_AILMENT_VALUE
}

func (ailments *Ailments) RemoveAilment(result int) {
	ailments.remaining[result-1] = REMOVED_AILMENT_VALUE
}

/*************************************
* Turn Based Structs
*************************************/
type Turn struct {
	round          int
	dice           []Dice
	result         int
	expression     string
	removedAilment bool
	lostLife       bool
}

type Dice struct {
	value int
}

func (d *Dice) Roll() {
	d.value = rand.IntN(6) + 1
}

func CreateAndRollDie() Dice {
	die := Dice{}
	die.Roll()
	return die
}

func (t Turn) PrintRoll() {
	var sb strings.Builder
	for i := range t.dice {
		sb.WriteString("[ ")
		sb.WriteString(fmt.Sprintf("%d", t.dice[i].value))
		sb.WriteString(" ]")
		sb.WriteString(" ")
	}
	logger.LogSuccess(sb.String())
}

/*************************************
* Turn Functions
*************************************/
func DoTurnStart(player Player, turn Turn) {
	var sb strings.Builder
	for i := range player.Ailments.remaining {
		if player.Ailments.remaining[i] != REMOVED_AILMENT_VALUE {
			sb.WriteString(fmt.Sprintf("[ %d ]", player.Ailments.remaining[i]))
		}
	}

	logger.LogInfo("Starting Turn:", turn.round)
	logger.LogInfo("Ailments:", sb.String())
	logger.LogInfo("Lives:", player.Lives)
	time.Sleep(2 * time.Second)
}

func DoRollDice(turn *Turn) {
	dice := make([]Dice, NUM_DICE)
	dice[0] = CreateAndRollDie()
	dice[1] = CreateAndRollDie()
	dice[2] = CreateAndRollDie()
	turn.dice = dice
}

func DoViewRoll(turn *Turn) {
	turn.PrintRoll()
}

func DoDiceChoice(turn *Turn) {
	reader := bufio.NewReader(os.Stdin)
	for i := range turn.dice {
		logger.LogWarning("Reroll die", i+1, "with value", turn.dice[i].value, "? (y/n)")
		fmt.Print("> ")
		opt, _ := reader.ReadString('\n')
		opt = strings.TrimSpace(opt)

		if opt == "y" {
			turn.dice[i].Roll()
		}
	}
}

func DoTypeExpression(turn *Turn) {
	reader := bufio.NewReader(os.Stdin)

	logger.LogWarning("Type the math expression using +  - * / operators")
	fmt.Print("> ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	turn.expression = text
}

func DoEvaluateExpression(turn *Turn) {
	postfix := math.InfixToPostfix(turn.expression)
	val, _ := math.EvaluatePostfixExpression(postfix)
	turn.result = val
}

func DoApplyResult(player *Player, turn *Turn) {
	if player.Ailments.HasAilment(turn.result) {
		player.Ailments.RemoveAilment(turn.result)
		turn.removedAilment = true
	} else {
		player.Lives--
		turn.lostLife = true
	}
}

func DoShowTurnResult(turn Turn) {
	if turn.removedAilment {
		logger.LogSuccess("Boom! You removed an ailment", turn.result)
	} else {
		logger.LogError("Oh no! You didnt have an ailment equal to", turn.result, "You lose a life!")
	}
}

func DoTurnEnd(player Player) {
	if !player.Ailments.HasAilments() {
		logger.LogSuccess("Wow, you win!")
	}

	if !player.HasLives() {
		logger.LogError("Darn, you lose!")
	}
}

/*************************************
* Main Loop
*************************************/
func main() {
	player := CreatePlayer()
	roundNumber := 0

	for player.Ailments.HasAilments() && player.HasLives() {
		roundNumber++
		turn := Turn{round: roundNumber}
		turnStack := CreateTurnStack()

		for !turnStack.IsEmpty() {
			turnState, _ := turnStack.Top()
			turnStack.Pop()

			switch turnState {
			case GS_TurnStart:
				DoTurnStart(player, turn)
			case GS_RollDice:
				DoRollDice(&turn)
			case GS_ViewRoll:
				DoViewRoll(&turn)
			case GS_DiceChoice:
				DoDiceChoice(&turn)
			case GS_ViewFinalRoll:
				DoViewRoll(&turn)
			case GS_TypeExpression:
				DoTypeExpression(&turn)
			case GS_EvaluateExpression:
				DoEvaluateExpression(&turn)
			case GS_ApplyResult:
				DoApplyResult(&player, &turn)
			case GS_ShowTurnResult:
				DoShowTurnResult(turn)
			case GS_TurnEnd:
				DoTurnEnd(player)
			}
		}
	}
}
