package main

import (
	"dicer/pkg/config"
	"dicer/pkg/math"
	"dicer/pkg/models"
	"dicer/pkg/single"
	"dicer/pkg/stack"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

/*************************************
* Bubble Tea Model
*************************************/
type model struct {
	roundNumber  int
	choices      []string
	selected     map[int]struct{}
	cursor       int
	textInput    textinput.Model
	message      string
	player       models.Player
	turn         *models.Turn
	width        int
	height       int
	instructions string
	debug        string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "( x + y ) / z"
	ti.Focus()
	ti.CharLimit = 24
	ti.Width = 24

	var choices []string
	for i := 0; i < config.NumDice; i++ {
		choices = append(choices, "")
	}

	return model{
		roundNumber: 1,
		selected:    make(map[int]struct{}),
		choices:     choices,
		textInput:   ti,
		player:      models.CreatePlayer(config.MaxLives, config.NumAilments),
		turn:        models.CreateTurn(1),
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
		m.turn.Stack.Push(models.GS_GameOver)
		return
	}

	// Set state for next turn
	next := m.roundNumber + 1
	m.roundNumber = next
	m.turn = models.CreateTurn(next)
	m.textInput.Reset()
	m.selected = make(map[int]struct{})
	m.cursor = 0
}

func (m *model) getCurrentState() (models.TurnPhase, error) {
	return m.turn.Stack.Top()
}

func (m *model) resetDice() {
	m.turn.Dice = nil
}

func (m *model) isValidExpression() bool {
	exp := m.turn.Expression

	isBalanced := math.IsBalancedParens(exp)
	if !isBalanced {
		m.debug = "Parenthesis not balanced"
		return false
	}

	isSpaceDelimted := isSpaceDelimted(exp)
	if !isSpaceDelimted {
		m.debug = "Every character must be separated by a space"
		return false
	}

	numbers := &single.LinkedList{}
	for num := range m.turn.Dice {
		numbers.InsertAtHead(m.turn.Dice[num].Value)
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

	if numOperators != config.NumDice-1 {
		m.debug = "Too many operators"
		return false
	}

	m.debug = ""
	return true
}

func (m *model) submitExpression() {
	m.turn.Expression = m.textInput.Value()

	if !m.isValidExpression() {
		m.turn.Stack.Push(models.GS_ExpressionPhase)
		m.textInput.Reset()
		return
	}

	m.turn.Result = math.EvaluateExpression(m.turn.Expression)
	m.turn.ApplyResult(&m.player)
}

func (m *model) toggleDiceSelection() {
	if _, ok := m.selected[m.cursor]; ok {
		delete(m.selected, m.cursor)
	} else {
		m.selected[m.cursor] = struct{}{}
	}
}

func isSpaceDelimted(exp string) bool {
	stack := stack.StackList[rune]{}

	for index, runeValue := range exp {
		if index == 0 {
			stack.Push(runeValue)
			continue
		}

		lastRune, _ := stack.Top()
		if runeValue != 32 && lastRune != 32 {
			return false
		}
		stack.Push(runeValue)
	}

	return true
}

/*************************************
* State Handlers
*************************************/
type StateHandler func(*model, tea.Msg) (tea.Model, tea.Cmd)

// stateHandlers maps each game state to its handler function
var stateHandlers = map[models.TurnPhase]StateHandler{
	models.GS_TurnStart:       handleTurnStart,
	models.GS_RollPhase:       handleRollPhase,
	models.GS_ExpressionPhase: handleExpressionPhase,
	models.GS_ResultsPhase:    handleResultsPhase,
	models.GS_GameOver:        handleGameOver,
}

func handleTurnStart(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Time to roll!"
	m.instructions = "Press [ r ] to roll the dice"
	return *m, nil
}

func handleRollPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Select which die to re-roll."
	m.instructions = "[ left ] [ right ] to navigate [ space ] to toggle [ enter ] to submit"
	return *m, nil
}

func handleExpressionPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Type your expression! Ensure there is a space between each character. Valid operators include ( ) * / + -"
	m.instructions = "[ enter ] to submit"
	// if m.turn.Expression != "" {
	// 	m.message = m.message + "\nInvalid expression. Try again."
	// }
	m.textInput, _ = m.textInput.Update(msg)
	return *m, nil
}

func handleResultsPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.resetDice()
	enteredText := fmt.Sprintf("You entered %s which evaluates to %d.", m.turn.Expression, m.turn.Result)
	var resultText string
	if m.turn.LostLife {
		resultText = fmt.Sprintf("You lost a life! %d lives remaining.", m.player.Lives)
	} else if m.turn.RemovedAilment {
		resultText = fmt.Sprintf("Hit! You removed %d.", m.turn.Result)
	}
	m.message = enteredText + "\n" + resultText
	m.instructions = "[ space ] to continue"
	return *m, nil
}

func handleGameOver(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.player.Ailments.HasAilments() {
		m.message = "You win! How good."
	}

	if !m.player.HasLives() {
		m.message = "You lose! Bummer."
	}

	m.instructions = "[ enter ] to restart the game"
	return *m, nil
}

func (m model) processGameState(state models.TurnPhase, msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *model) handleEnterKey(state models.TurnPhase) {
	switch state {
	case models.GS_RollPhase:
		m.turn.RollSelectedDice(m.selected)
		m.turn.Stack.Pop()
	case models.GS_ExpressionPhase:
		m.submitExpression()
		m.turn.Stack.Pop()
	}
}

// On [ space ] press
func (m *model) handleSpaceKey(state models.TurnPhase) {
	switch state {
	case models.GS_RollPhase:
		m.toggleDiceSelection()
	case models.GS_ResultsPhase:
		m.endTurn()
	}
}

// On [ r ] press
func (m *model) handleRollKey(state models.TurnPhase) {
	if state == models.GS_TurnStart {
		m.turn.Stack.Pop()
		m.turn.RollDice()
	}
}

// On [ left key ] press
func (m *model) handleLeftKey(state models.TurnPhase) {
	if state == models.GS_RollPhase && m.cursor > 0 {
		m.cursor--
	}
}

// On [ right ] press
func (m *model) handleRightKey(state models.TurnPhase) {
	if state == models.GS_RollPhase && m.cursor < len(m.choices)-1 {
		m.cursor++
	}
}

// Forward [ ] key presses
func (m *model) handleKeyPress(key tea.KeyMsg, state models.TurnPhase) tea.Cmd {
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
		if keyMsg.String() == "enter" && currentState == models.GS_GameOver {
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
	return m.renderGameLayout(m.width, m.height)
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
