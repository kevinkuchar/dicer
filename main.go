package main

import (
	"dicer/pkg/math"
	"dicer/pkg/single"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

/*************************************
* Game State Flow
*************************************/
type GameState int

const (
	GS_TurnStart       GameState = iota
	GS_RollPhase                 // Re-roll selection phase
	GS_ExpressionPhase           // Gather user input for expression
	GS_ResultsPhase              // Show results of expression
	GS_GameOver                  // Game over state
)

/*************************************
* Game Configuration
*************************************/
const MAX_LIVES = 1
const NUM_AILMENTS = 9
const NUM_DICE = 3
const REMOVED_AILMENT_VALUE = -1

/*************************************
* Dice
*************************************/
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
* Bubble Tea Model
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
	debug       string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "( x + y ) / z"
	ti.Focus()
	ti.CharLimit = 24
	ti.Width = 24

	var choices []string
	for i := 0; i < NUM_DICE; i++ {
		choices = append(choices, "")
	}

	return model{
		roundNumber: 1,
		selected:    make(map[int]struct{}),
		choices:     choices,
		textInput:   ti,
		player:      CreatePlayer(MAX_LIVES, NUM_AILMENTS),
		turn:        CreateTurn(1),
		message:     "Press any [ key ] to begin",
	}
}

func newModel(current *model) model {
	model := initialModel()
	model.height = current.height
	model.width = current.width
	return model
}

/*************************************
* Model Utilities
*************************************/
func (m *model) checkGameOver() bool {
	if !m.player.Ailments.HasAilments() {
		return true
	}
	if !m.player.HasLives() {
		return true
	}
	return false
}

func (m *model) endTurn() {
	// Check game over condition first
	if isGameOver := m.checkGameOver(); isGameOver != false {
		m.turn.stack.Push(GS_GameOver)
		return
	}

	// Set state for next turn
	next := m.roundNumber + 1
	m.roundNumber = next
	m.turn = CreateTurn(next)
	m.textInput.Reset()
	m.selected = make(map[int]struct{})
	m.cursor = 0
}

func (m *model) getCurrentState() (GameState, error) {
	return m.turn.stack.Top()
}

func (m *model) resetDice() {
	m.turn.dice = nil
}

func (m *model) isValidExpression() bool {
	exp := m.turn.expression

	isBalanced := math.IsBalancedParens(exp)
	if !isBalanced {
		m.debug = "Parenthesis not balanced"
		return false
	}

	numbers := &single.LinkedList{}
	for num := range m.turn.dice {
		numbers.InsertAtHead(m.turn.dice[num].value)
	}

	numOperators := 0

	for _, char := range exp {
		if math.IsOperand(string(char)) {
			numbers.RemoveVal(int(char - '0'))
		}
		if math.IsOperator(string(char)) {
			numOperators++
		}
	}

	if numbers.Head != nil {
		m.debug = "Expression doesn't include all dice rolls"
		return false
	}

	if numOperators != NUM_DICE-1 {
		m.debug = "Too many operators"
		return false
	}

	m.debug = ""
	return true
}

func (m *model) submitExpression() {
	m.turn.expression = m.textInput.Value()

	if !m.isValidExpression() {
		m.turn.stack.Push(GS_ExpressionPhase)
		m.textInput.Reset()
		return
	}

	m.turn.result = math.EvaluateExpression(m.turn.expression)
	m.turn.ApplyResult(&m.player)
}

func (m *model) toggleDiceSelection() {
	if _, ok := m.selected[m.cursor]; ok {
		delete(m.selected, m.cursor)
	} else {
		m.selected[m.cursor] = struct{}{}
	}
}

/*************************************
* State Handlers
*************************************/
type StateHandler func(*model, tea.Msg) (tea.Model, tea.Cmd)

// stateHandlers maps each game state to its handler function
var stateHandlers = map[GameState]StateHandler{
	GS_TurnStart:       handleTurnStart,
	GS_RollPhase:       handleRollPhase,
	GS_ExpressionPhase: handleExpressionPhase,
	GS_ResultsPhase:    handleResultsPhase,
	GS_GameOver:        handleGameOver,
}

func handleTurnStart(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Press [ r ] to roll the dice"
	return *m, nil
}

func handleRollPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Select which die to re-roll. [ left arrow ] and [ right arrow ] to navigate [ space ] to toggle and [ enter ] to submit"
	return *m, nil
}

func handleExpressionPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Type your expression and press [ enter ] to submit. Valid operators include ( ) * / + -"

	if m.turn.expression != "" {
		m.message = m.message + "\nInvalid expression. Try again."
	}
	m.textInput, _ = m.textInput.Update(msg)
	return *m, nil
}

func handleResultsPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.resetDice()
	enteredText := fmt.Sprintf("You entered %s which evaluated to %d.", m.turn.expression, m.turn.result)
	var resultText string
	if m.turn.lostLife {
		resultText = fmt.Sprintf("You lost a life! %d lives remaining.", m.player.Lives)
	} else if m.turn.removedAilment {
		resultText = fmt.Sprintf("Hit! You removed %d.", m.turn.result)
	}
	m.message = enteredText + "\n" + resultText + "\n" + "Press [ space ] to continue"
	return *m, nil
}

func handleGameOver(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.player.Ailments.HasAilments() {
		m.message = "You win! Press [ enter ] to restart the game"
	}
	if !m.player.HasLives() {
		m.message = "You lose! Press [ enter ] to restart the game"
	}

	return *m, nil
}

func (m model) processGameState(state GameState, msg tea.Msg) (tea.Model, tea.Cmd) {
	handler, exists := stateHandlers[state]
	if !exists {
		// Unknown state
		return m, nil
	}

	return handler(&m, msg)
}

/*************************************
* Key handlers
*************************************/
// On [ enter ] press
func (m *model) handleEnterKey(state GameState) {
	switch state {
	case GS_RollPhase:
		m.turn.RollSelectedDice(m.selected)
		m.turn.stack.Pop()
	case GS_ExpressionPhase:
		m.submitExpression()
		m.turn.stack.Pop()
	}
}

// On [ space ] press
func (m *model) handleSpaceKey(state GameState) {
	switch state {
	case GS_RollPhase:
		m.toggleDiceSelection()
	case GS_ResultsPhase:
		m.endTurn()
	}
}

// On [ r ] press
func (m *model) handleRollKey(state GameState) {
	if state == GS_TurnStart {
		m.turn.stack.Pop()
		m.turn.RollDice()
	}
}

// On [ left key ] press
func (m *model) handleLeftKey(state GameState) {
	if state == GS_RollPhase && m.cursor > 0 {
		m.cursor--
	}
}

// On [ right ] press
func (m *model) handleRightKey(state GameState) {
	if state == GS_RollPhase && m.cursor < len(m.choices)-1 {
		m.cursor++
	}
}

// Forward [ ] key presses
func (m *model) handleKeyPress(key tea.KeyMsg, state GameState) tea.Cmd {
	switch key.String() {
	case "ctrl+c", "q":
		return tea.Quit

	case "r":
		m.handleRollKey(state)

	case "left", "h":
		m.handleLeftKey(state)

	case "right", "l":
		m.handleRightKey(state)

	case "enter":
		m.handleEnterKey(state)

	case " ":
		m.handleSpaceKey(state)
	}

	return nil
}

/*************************************
* Bubble Tea Functions
*************************************/
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize early
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Get current state
	currentState, err := m.getCurrentState()
	if err != nil {
		return m, tea.Quit
	}

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Reset Game
		if keyMsg.String() == "enter" && currentState == GS_GameOver {
			model := newModel(&m)
			return model, nil
		}

		if cmd := m.handleKeyPress(keyMsg, currentState); cmd != nil {
			return m, cmd
		}
		// Update state after key handling
		currentState, _ = m.getCurrentState()
	}

	// Process current game state
	return m.processGameState(currentState, msg)
}

func (m model) View() string {
	header := m.getHeader(m.width)
	sidebar := m.getStatusSidebar(m.player.Lives, m.roundNumber)
	inputWindow := m.getInputWindow(m.message)
	ailmentsBar := m.getAilmentsBar(m.width, m.player.Ailments)

	return m.renderGameLayout(m.width, m.height, header, sidebar, inputWindow, ailmentsBar)
}

/*************************************
* Main Loop
*************************************/
func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
