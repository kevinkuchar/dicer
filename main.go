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
	GS_TurnStart          GameState = iota
	GS_DiceChoice                   // Re-roll selection phase
	GS_TypeExpression               // Gather user input for expression
	GS_EvaluateExpression           // Validate input and evaluate outcome of expression
	GS_TurnEnd                      // Create a new turn
	GS_GameOver                     // Game over
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

	stack.Push(GS_TurnEnd)
	stack.Push(GS_EvaluateExpression)
	stack.Push(GS_TypeExpression)
	stack.Push(GS_DiceChoice)
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
	ti.Placeholder = "( 5 + 2 ) / 1"
	ti.Focus()
	ti.CharLimit = 24
	ti.Width = 20

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
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

/*************************************
* State Handlers
*************************************/
type StateHandler func(*model, tea.Msg) (tea.Model, tea.Cmd)

// stateHandlers maps each game state to its handler function
var stateHandlers = map[GameState]StateHandler{
	GS_TurnStart:          handleTurnStart,
	GS_DiceChoice:         handleDiceChoice,
	GS_TypeExpression:     handleTypeExpression,
	GS_EvaluateExpression: handleEvaluateExpression,
	GS_TurnEnd:            handleTurnEnd,
	GS_GameOver:           handleGameOver,
}

func handleTurnStart(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Press r to roll dice"
	return *m, nil
}

func handleDiceChoice(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Select which die to reroll"
	return *m, nil
}

func handleTypeExpression(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.message = "Type your expression using ( ) * / + -"
	m.textInput, _ = m.textInput.Update(msg)
	return *m, nil
}

func handleEvaluateExpression(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.turn.expression = m.textInput.Value()

	// TODO: Add validation before evaluation
	// if !isValidExpression(m.turn.expression) {
	//     m.turn.stack.Push(GS_TypeExpression)
	//     m.message = "Invalid expression. Try again."
	//     return *m, nil
	// }

	m.turn.EvaluateExpression()
	m.turn.ApplyResult(&m.player)

	if m.checkGameOver() {
		return handleGameOver(m, msg)
	}

	m.message = m.formatEvaluationResult()
	m.turn.stack.Pop()
	return *m, nil
}

func handleTurnEnd(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.resetDice()
	return *m, nil
}

func handleGameOver(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	m.resetDice()
	m.message = "GAME OVER!!"
	return *m, nil
}

/*************************************
* Model Utilities
*************************************/
func (m *model) incrementTurn() {
	m.turn = CreateTurn(m.turn.round + 1)
	m.textInput.Reset()
	m.selected = make(map[int]struct{})
	m.cursor = 0
}

func (m *model) getCurrentState() (GameState, error) {
	return m.turn.stack.Top()
}

func (m *model) checkGameOver() bool {

	if !m.player.Ailments.HasAilments() {
		return true
		// 	return "You win!"
	}
	// }
	if !m.player.HasLives() {
		return true
		// return "You lose
	}
	return false
}

func (m *model) rerollSelectedDice() {
	for i := range m.selected {
		m.turn.dice[i].Roll()
	}
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

func (m *model) formatEvaluationResult() string {
	var msg string
	if m.turn.lostLife {
		msg = fmt.Sprintf("You lost a life. %d missed", m.turn.result)
	} else if m.turn.removedAilment {
		msg = fmt.Sprintf("You hit! %d", m.turn.result)
	}
	return msg + " Press space to continue!"
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
func (m *model) handleEnterKey(state GameState) {
	switch state {
	case GS_DiceChoice:
		m.rerollSelectedDice()
		m.turn.stack.Pop()
	case GS_TypeExpression:
		m.turn.stack.Pop()
	}
}

func (m *model) handleSpaceKey(state GameState) {
	switch state {
	case GS_DiceChoice:
		m.toggleDiceSelection()
	case GS_TurnEnd:
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
	if state == GS_DiceChoice && m.cursor > 0 {
		m.cursor--
	}
}

func (m *model) handleRightKey(state GameState) {
	if state == GS_DiceChoice && m.cursor < len(m.choices)-1 {
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

	// Check game over condition first
	// if gameOverMsg := m.checkGameOver(); gameOverMsg != "" {
	// 	m.message = gameOverMsg
	// 	return m, nil
	// }

	// Get current state
	currentState, err := m.getCurrentState()
	if err != nil {
		// Handle stack error - could transition to game over
		return m, nil
	}

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
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
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
