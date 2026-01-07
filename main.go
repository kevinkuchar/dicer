package main

import (
	"dicer/pkg/math"
	"dicer/pkg/stack"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
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

/*************************************
* Turn Functions
*************************************/
func DoTurnStart(player Player, turn Turn) {

}

func DoRollDice(turn *Turn) {
	dice := make([]Dice, NUM_DICE)
	dice[0] = CreateAndRollDie()
	dice[1] = CreateAndRollDie()
	dice[2] = CreateAndRollDie()
	turn.dice = dice
}

func DoViewRoll(turn Turn, area *pterm.AreaPrinter) {
	RenderDice(turn, area, "Initial Roll")
}

func DoDiceChoice(turn *Turn) {
	// Initialize an empty slice to hold the options.
	var options []string

	// Populate the options slice with 100 options.
	for i := range turn.dice {
		options = append(options, fmt.Sprintf("Reroll Die %d?", i+1))
	}

	selectedOptions, _ := pterm.DefaultInteractiveMultiselect.WithOptions(options).WithShowSelectedOptions(false).Show()

	for idx := range selectedOptions {
		turn.dice[idx].Roll()
	}
}

func DoViewFinalRoll(turn Turn, area *pterm.AreaPrinter) {
	RenderDice(turn, area, "Final Roll")
}

func DoTypeExpression(turn *Turn) {
	result, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Type expression").Show()
	pterm.Println()
	turn.expression = result
	time.Sleep(time.Second)
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
		pterm.Println(pterm.Green("Boom! You removed an ailment! Your expression was equal to ", turn.result, "!"))
	} else {
		pterm.Println(pterm.Red("Oh no! You didnt have an ailment equal to ", turn.result, ". You lose a life!"))
	}
	time.Sleep(time.Second * 2)
}

func DoTurnEnd(player Player) {
	if !player.Ailments.HasAilments() {
		pterm.Println(pterm.Green("Wow, You Win!"))
	}

	if !player.HasLives() {
		pterm.Println(pterm.Red("Darn, you lose!"))
	}
}

/*************************************
* UI
*************************************/
func Render(player Player, turn Turn, area *pterm.AreaPrinter) {
	paddedBox := pterm.DefaultBox.WithRightPadding(6).WithLeftPadding(6).WithTopPadding(1).WithBottomPadding(1)

	// Logo
	logo := GetLogo()

	// Ailments
	var asb strings.Builder
	for i := range player.Ailments.remaining {
		if player.Ailments.remaining[i] != REMOVED_AILMENT_VALUE {
			asb.WriteString(pterm.LightWhite(fmt.Sprintf("[ %d ]", i+1)))
		} else {
			asb.WriteString(pterm.Red(fmt.Sprintf("[ %d ]", i+1)))
		}
	}
	ailmentsTitle := pterm.Yellow("Ailments")
	ailmentsParent := paddedBox.WithTitle(ailmentsTitle).WithTopPadding(1).Sprint(asb.String())

	// Turn
	roundTitle := pterm.Green("Turn")
	round := paddedBox.WithTitle(roundTitle).WithTitleTopCenter().Sprintf("%d", turn.round)

	// Lives
	livesTitle := pterm.Blue("Lives")
	lives := paddedBox.WithTitle(livesTitle).WithTitleTopCenter().Sprintf("%d", player.Lives)

	statusPanels := pterm.Panels{
		{
			{Data: round},
			{Data: lives},
			{Data: ailmentsParent},
		},
	}

	// Render the panels with a padding of 5
	statusLayout, _ := pterm.DefaultPanel.WithPanels(statusPanels).WithPadding(1).Srender()

	area.Update(
		logo,
		statusLayout,
	)
}

func RenderDice(turn Turn, area *pterm.AreaPrinter, header string) {
	paddedBox := pterm.DefaultBox.WithVerticalPadding(0).WithHorizontalPadding(4)

	// Create boxes with the title positioned differently and containing different content
	box1 := paddedBox.WithTitle(pterm.LightRed("Die 1")).WithTitleTopCenter().Sprint(turn.dice[0].value)
	box2 := paddedBox.WithTitle(pterm.LightRed("Die 2")).WithTitleTopCenter().Sprint(turn.dice[1].value)
	box3 := paddedBox.WithTitle(pterm.LightRed("Die 3")).WithTitleTopCenter().Sprint(turn.dice[2].value)

	pterm.Println(pterm.Cyan(header))

	pterm.DefaultPanel.WithPanels([][]pterm.Panel{
		{{Data: box1}, {Data: box2}, {Data: box3}},
	}).Render()

}

func GetLogo() string {
	text, _ := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("Dice", pterm.FgCyan.ToStyle()),
		putils.LettersFromStringWithStyle("rooski", pterm.FgGreen.ToStyle())).
		Srender()

	return "\n" + text + "\n"
}

/*************************************
* Main Loop
*************************************/
func main() {
	player := CreatePlayer()
	roundNumber := 0

	area, _ := pterm.DefaultArea.Start()

	for player.Ailments.HasAilments() && player.HasLives() {
		roundNumber++
		turn := Turn{round: roundNumber}
		turnStack := CreateTurnStack()

		Render(player, turn, area)

		for !turnStack.IsEmpty() {
			turnState, _ := turnStack.Top()
			turnStack.Pop()

			switch turnState {
			case GS_TurnStart:
				DoTurnStart(player, turn)
			case GS_RollDice:
				DoRollDice(&turn)
			case GS_ViewRoll:
				DoViewRoll(turn, area)
			case GS_DiceChoice:
				DoDiceChoice(&turn)
			case GS_ViewFinalRoll:
				DoViewFinalRoll(turn, area)
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

	area.Stop()
}
