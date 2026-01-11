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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	stack          *stack.ArrayStack[GameState]
}

func CreateTurn(round int) *Turn {
	turn := &Turn{
		round: round,
		stack: CreateTurnStack(),
	}

	return turn
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
* Bubble Tea
*************************************/
type model struct {
	roundNumber int
	player      Player
	turn        *Turn
}

func initialModel() model {
	return model{
		roundNumber: 1,
		player:      CreatePlayer(),
		turn:        CreateTurn(1),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.player.Ailments.HasAilments() && m.player.HasLives() {

	}

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {
		case "r":
			m.player.Lives--
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {

	logo := getLogo()
	statusGrid := getStatusGrid(m.player.Lives, m.roundNumber, m.player.Ailments)

	// Combine logo and grid with some spacing
	return lipgloss.JoinVertical(lipgloss.Left, logo, statusGrid)
}

/**********************************
* Vibe Coded UI Components
**********************************/
func getLogo() string {
	// Pastel color scheme
	pastelGreen := lipgloss.Color("#9FE2BF")
	pastelBlue := lipgloss.Color("#87CEEB")

	// Style for "Dice" part (green)
	diceStyle := lipgloss.NewStyle().
		Foreground(pastelGreen).
		Bold(true)

	// Style for "r" part (blue)
	rStyle := lipgloss.NewStyle().
		Foreground(pastelBlue).
		Bold(true)

	// Combine the styled parts
	logoText := diceStyle.Render("///////////////////// Dice") + rStyle.Render("r ///////////////////////")

	// Add padding around the whole logo
	paddingStyle := lipgloss.NewStyle().Padding(1)

	return paddingStyle.Render(logoText)
}

func getStatusGrid(lives int, turnNumber int, ailments *Ailments) string {
	textColor := lipgloss.Color("#EAEAEA")
	// Pastel color scheme
	red := lipgloss.Color("#963c31")
	blue := lipgloss.Color("#45657A")
	gray := lipgloss.Color("#333333")

	createStyle := func(background lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(background).
			Foreground(textColor).
			Padding(1, 2).
			Align(lipgloss.Center)
	}
	// Style for Turn Number section
	turnStyle := createStyle(blue)

	// Style for Lives section
	livesStyle := createStyle(red)

	// Style for Ailments section
	ailmentsStyle := createStyle(gray)

	// Build Lives content
	livesText := fmt.Sprintf("Lives: %d", lives)
	livesView := livesStyle.Render(livesText)

	// Build Turn Number content
	turnText := fmt.Sprintf("Turn: %d", turnNumber)
	turnView := turnStyle.Render(turnText)

	// Build Ailments content - show all ailment numbers in a row
	var ailmentNumbers []string
	for i := range ailments.remaining {
		if ailments.remaining[i] != REMOVED_AILMENT_VALUE {
			ailmentNumbers = append(ailmentNumbers, fmt.Sprintf("%d", ailments.remaining[i]))
		}
	}

	ailmentsText := "Ailments: "
	if len(ailmentNumbers) > 0 {
		ailmentsText += strings.Join(ailmentNumbers, " ")
	} else {
		ailmentsText += "None"
	}
	ailmentsView := ailmentsStyle.Render(ailmentsText)

	// Join sections horizontally
	grid := lipgloss.JoinHorizontal(lipgloss.Top, livesView, turnView, ailmentsView)

	return grid
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
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// func main() {
// 	player := CreatePlayer()
// 	roundNumber := 0

// 	for player.Ailments.HasAilments() && player.HasLives() {
// 		roundNumber++
// 		turn := Turn{round: roundNumber}
// 		turnStack := CreateTurnStack()

// 		for !turnStack.IsEmpty() {
// 			turnState, _ := turnStack.Top()
// 			turnStack.Pop()

// 			switch turnState {
// 			case GS_TurnStart:
// 				DoTurnStart(player, turn)
// 			case GS_RollDice:
// 				DoRollDice(&turn)
// 			case GS_ViewRoll:
// 				DoViewRoll(&turn)
// 			case GS_DiceChoice:
// 				DoDiceChoice(&turn)
// 			case GS_ViewFinalRoll:
// 				DoViewRoll(&turn)
// 			case GS_TypeExpression:
// 				DoTypeExpression(&turn)
// 			case GS_EvaluateExpression:
// 				DoEvaluateExpression(&turn)
// 			case GS_ApplyResult:
// 				DoApplyResult(&player, &turn)
// 			case GS_ShowTurnResult:
// 				DoShowTurnResult(turn)
// 			case GS_TurnEnd:
// 				DoTurnEnd(player)
// 			}
// 		}
// 	}
// }
