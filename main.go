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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

/*************************************
* Game State Flow
*************************************/
type GameState int

const (
	GS_TurnStart          GameState = iota
	GS_DiceChoice                   // User selects what to do with each dice
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
	stack.Push(GS_DiceChoice)
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

func (t *Turn) RollDice() {
	dice := make([]Dice, NUM_DICE)
	dice[0] = CreateAndRollDie()
	dice[1] = CreateAndRollDie()
	dice[2] = CreateAndRollDie()
	t.dice = dice
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
* Bubble Tea
*************************************/
type model struct {
	roundNumber int
	choices     []string
	selected    map[int]struct{}
	cursor      int
	textInput   textinput.Model
	message     string
	player      Player
	turn        *Turn
	width       int
	height      int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "( 5 + 2 ) / 1"
	ti.Focus()
	ti.CharLimit = 24
	ti.Width = 20

	return model{
		roundNumber: 1,
		selected:    make(map[int]struct{}),
		choices:     []string{"", "", ""},
		textInput:   ti,
		player:      CreatePlayer(),
		turn:        CreateTurn(1),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// if m.player.Ailments.HasAilments() && m.player.HasLives() {

	// }
	startingState, _ := m.turn.stack.Top()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// Key listners
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "r" key rolls the dice
		case "r":
			if startingState == GS_TurnStart {
				m.turn.stack.Pop()
				m.turn.RollDice()
			}

		// The "left" and "h" keys move the cursor left
		case "left", "h":
			if startingState == GS_DiceChoice && m.cursor > 0 {
				m.cursor--
			}

		// The "right" and "l" keys move the cursor right
		case "right", "l":
			if startingState == GS_DiceChoice && m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key submits user input
		case "enter":
			switch startingState {
			case GS_DiceChoice:
				for i := range m.selected {
					m.turn.dice[i].Roll()
				}
				m.turn.stack.Pop()
			case GS_TypeExpression:
				m.turn.stack.Pop()
			}

		// The spacebar selects user choices
		case " ":
			if startingState == GS_DiceChoice {
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}
		}
	}

	// Game state
	currentState, _ := m.turn.stack.Top()
	var cmd tea.Cmd

	switch currentState {
	case GS_TurnStart:
		m.message = "Press r to roll dice"
	case GS_DiceChoice:
		m.message = "Select which die to reroll"
	case GS_TypeExpression:
		m.message = "Type your expression using ( ) * / + -"
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case GS_EvaluateExpression:
		m.message = "Evaluating expression..."
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	header := m.getHeader(m.width)
	sidebar := m.getStatusSidebar(m.player.Lives, m.roundNumber)
	mainContent := m.getMainContent(m.message, m.turn)
	ailmentsBar := m.getAilmentsBar(m.width, m.player.Ailments)

	return m.renderGameLayout(m.width, m.height, header, sidebar, mainContent, ailmentsBar)
}

/*************************************
* Turn Functions
*************************************/

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
