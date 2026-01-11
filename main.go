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

func (m *model) IncrementTurn() {
	m.turn = CreateTurn(m.turn.round + 1)
	m.textInput.Reset()
	m.selected = make(map[int]struct{})
	m.cursor = 0
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			switch startingState {
			case GS_DiceChoice:
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			case GS_TurnEnd:
				m.IncrementTurn()
			}
		}
	}

	// Game state
	currentState, _ := m.turn.stack.Top()

	if !m.player.Ailments.HasAilments() || !m.player.HasLives() {
		if m.player.Lives > 0 {
			m.message = "You win!"
		} else {
			m.message = "You lose!"
		}
		return m, nil
	}

	switch currentState {
	case GS_TurnStart:
		m.message = "Press r to roll dice"
	case GS_DiceChoice:
		m.message = "Select which die to reroll"
	case GS_TypeExpression:
		m.message = "Type your expression using ( ) * / + -"
		m.textInput, _ = m.textInput.Update(msg)
		return m, nil
	case GS_EvaluateExpression:
		m.message = "Evaluating expression..."
		m.turn.expression = m.textInput.Value()
		// verify if valid
		// if not put other value back on stack

		// evaluate
		m.turn.EvaluateExpression()

		// apply
		m.turn.ApplyResult(&m.player)
		m.turn.stack.Pop()
		if m.turn.lostLife == true {
			m.message = fmt.Sprintf("You lost a life. %d missed", m.turn.result)
		} else if m.turn.removedAilment {
			m.message = fmt.Sprintf("You hit! %d", m.turn.result)
		}

		m.message = m.message + " Press space to continue!"
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
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
