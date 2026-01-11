package main

import (
	"dicer/pkg/math"
	"dicer/pkg/stack"
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

func (t *Turn) EvaluateExpression() {
	postfix := math.InfixToPostfix(t.expression)
	val, _ := math.EvaluatePostfixExpression(postfix)
	t.result = val
}

func (t *Turn) ApplyResult(player *Player) {
	if player.Ailments.HasAilment(t.result) {
		player.Ailments.RemoveAilment(t.result)
		t.removedAilment = true
	} else {
		player.Lives--
		t.lostLife = true
	}
}

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
		player:      CreatePlayer(),
		turn:        CreateTurn(1),
		message:     "Press any [ key ] to begin",
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
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

func (m *model) getCurrentState() (GameState, error) {
	return m.turn.stack.Top()
}

func (m *model) incrementTurn() {
	next := m.roundNumber + 1
	m.roundNumber = next
	m.turn = CreateTurn(next)
	m.textInput.Reset()
	m.selected = make(map[int]struct{})
	m.cursor = 0
}

func (m *model) resetDice() {
	m.turn.dice = nil
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
	m.textInput, _ = m.textInput.Update(msg)
	return *m, nil
}

func handleResultsPhase(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) processGameState(state GameState, msg tea.Msg) (tea.Model, tea.Cmd) {
	handler, exists := stateHandlers[state]
	if !exists {
		// Unknown state
		return m, nil
	}

	return handler(&m, msg)
}

func (m model) processGameOver(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.resetDice()
	if !m.player.Ailments.HasAilments() {
		m.message = "You win! Press [ enter ] to restart the game"
	}
	if !m.player.HasLives() {
		m.message = "You lose! Press [ enter ] to restart the game"
	}

	m.turn.stack.Push(GS_GameOver)
	return m, nil
}

/*************************************
* State Updaters
*************************************/
func (m *model) rerollSelectedDice() {
	for i := range m.selected {
		m.turn.dice[i].Roll()
	}
}

func (m *model) submitExpression() {
	m.turn.expression = m.textInput.Value()

	// TODO: Add validation before evaluation
	// if !isValidExpression(m.turn.expression) {
	//     m.turn.stack.Push(GS_TypeExpression)
	//     m.message = "Invalid expression. Try again."
	// }

	m.turn.EvaluateExpression()
	m.turn.ApplyResult(&m.player)
}

/*************************************
* Key handlers
*************************************/
func (m *model) handleEnterKey(state GameState) {
	switch state {
	case GS_RollPhase:
		m.rerollSelectedDice()
		m.turn.stack.Pop()
	case GS_ExpressionPhase:
		m.submitExpression()
		m.turn.stack.Pop()
	}
}

func (m *model) handleSpaceKey(state GameState) {
	switch state {
	case GS_RollPhase:
		m.toggleDiceSelection()
	case GS_ResultsPhase:
		m.incrementTurn()
	}
}

func (m *model) handleRollKey(state GameState) {
	if state == GS_TurnStart {
		m.turn.stack.Pop()
		m.turn.RollDice()
	}
}

func (m *model) handleLeftKey(state GameState) {
	if state == GS_RollPhase && m.cursor > 0 {
		m.cursor--
	}
}

func (m *model) handleRightKey(state GameState) {
	if state == GS_RollPhase && m.cursor < len(m.choices)-1 {
		m.cursor++
	}
}

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
			model := initialModel()
			model.height = m.height
			model.width = m.width
			return model, nil
		}

		if cmd := m.handleKeyPress(keyMsg, currentState); cmd != nil {
			return m, cmd
		}
		// Update state after key handling
		currentState, _ = m.getCurrentState()
	}

	// Process current game state
	// logger.LogInfo("I made it here", currentState)
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
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
