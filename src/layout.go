package main

import (
	"dicer/pkg/models"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

/*************************************
* UI Configuration
*************************************/
const SIDEBAR_WIDTH = 15
const DICE_WIDTH = 8

// Colors
const COLOR_BORDER = lipgloss.Color("#555555")
const COLOR_HIGHLIGHT = lipgloss.Color("#FFFF00")
const COLOR_LOGO_GREEN = lipgloss.Color("#9FE2BF")
const COLOR_LOGO_BLUE = lipgloss.Color("#87CEEB")
const COLOR_TEXT = lipgloss.Color("#EAEAEA")
const COLOR_RED = lipgloss.Color("#963c31")
const COLOR_BRIGHT_RED = lipgloss.Color("#C24D3F")
const COLOR_BLUE = lipgloss.Color("#45657A")
const COLOR_YELLOW = lipgloss.Color("#D9C380")
const COLOR_AILMENT_ACTIVE = lipgloss.Color("#56787a")
const COLOR_AILMENT_INACTIVE = lipgloss.Color("#333333")

func (m model) getHeader(width int) string {
	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(COLOR_LOGO_GREEN)).
		Bold(true)

	postfixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(COLOR_LOGO_BLUE)).
		Bold(true)

	logoText := prefixStyle.Render("Dice") + postfixStyle.Render("r")

	headerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(1, 0).
		Border(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	return headerStyle.Render(logoText)
}

func (m model) getStatusSidebar() string {
	createStyle := func(background lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(background).
			Foreground(COLOR_TEXT).
			Padding(1, 2).
			Align(lipgloss.Center).
			Width(SIDEBAR_WIDTH).
			MarginTop(1).
			MarginRight(1).
			Bold(true)
	}

	livesStyle := createStyle(COLOR_RED)
	turnStyle := createStyle(COLOR_BLUE)

	// Build content
	livesText := fmt.Sprintf("Lives: %d", m.player.Lives)
	turnText := fmt.Sprintf("Turn: %d", m.roundNumber)

	// Stack vertically
	sidebar := lipgloss.JoinVertical(
		lipgloss.Left,
		livesStyle.Render(livesText),
		turnStyle.Render(turnText),
	)

	return sidebar
}

func (m model) getAilmentsBar(width int) string {
	ailments := m.player.Ailments
	availableWidth := width - len(ailments.Remaining) - 1
	boxWidth := availableWidth / len(ailments.Remaining)

	createBoxStyle := func(background lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(background).
			Foreground(COLOR_TEXT).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(boxWidth).
			MarginRight(1)
	}

	// Create 9 boxes, one for each ailment (1-9)
	var boxes []string
	for i := 1; i <= len(ailments.Remaining); i++ {
		var boxStyle lipgloss.Style
		if ailments.HasAilment(i) {
			boxStyle = createBoxStyle(COLOR_AILMENT_ACTIVE)
		} else {
			boxStyle = createBoxStyle(COLOR_AILMENT_INACTIVE)
		}
		boxes = append(boxes, boxStyle.Render(fmt.Sprintf("%d", i)))
	}

	// Join boxes horizontally (margins will create the black line effect)
	bar := lipgloss.JoinHorizontal(lipgloss.Top, boxes...)

	// Style the bar container with full width and top border
	barStyle := lipgloss.NewStyle().
		Width(width).
		Padding(1, 0).
		Border(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Align(lipgloss.Center)

	return barStyle.Render(bar)
}

func (m model) getDice(dice []models.Dice) string {
	// Create a box style with border, no background, bold centered text
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(COLOR_BORDER).
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(DICE_WIDTH).
		Bold(true)

	// Create boxes for each die
	var boxes []string
	for _, die := range dice {
		boxes = append(boxes, boxStyle.Render(fmt.Sprintf("%d", die.Value)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
}

func (m model) getChoices() string {
	// Create a box style matching dice width
	createBoxStyle := func(isCursor bool, isSelected bool) lipgloss.Style {
		borderColor := COLOR_BORDER
		if isCursor {
			borderColor = COLOR_HIGHLIGHT
		}

		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Align(lipgloss.Center).
			Width(DICE_WIDTH)

		if isSelected {
			style = style.Bold(true)
		}

		return style
	}

	// Create boxes for each choice
	var boxes []string
	for i := range m.choices {
		isCursor := m.cursor == i
		_, isSelected := m.selected[i]

		boxStyle := createBoxStyle(isCursor, isSelected)

		// Build the content for the box
		content := ""
		if isSelected {
			content += "x"
		}

		boxes = append(boxes, boxStyle.Render(content))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
}

func (m model) getBoard(message string) string {
	contentStyle := lipgloss.NewStyle().
		Padding(1, 2)

	dice := m.getDice(m.turn.Dice)

	turnState, _ := m.turn.Stack.Top()

	choices := ""
	if turnState == models.GS_RollPhase {
		choices = m.getChoices()
	}

	expression := ""
	if turnState == models.GS_ExpressionPhase {
		expression = m.textInput.View()
	}

	mainContent := lipgloss.JoinVertical(
		lipgloss.Top,
		contentStyle.Render(message),
		dice,
		choices,
		expression,
	)

	return mainContent
}

func (m model) getInstructions(width int) string {
	style := lipgloss.NewStyle().
		Width(width/2 - 2).
		Align(lipgloss.Left).
		PaddingLeft(1).
		Foreground(lipgloss.Color(COLOR_YELLOW))

	return style.Render(m.instructions)
}

func (m model) getDebug(width int) string {
	style := lipgloss.NewStyle().
		Width(width/2 - 2).
		Align(lipgloss.Right).
		PaddingRight(1).
		Foreground(lipgloss.Color(COLOR_BRIGHT_RED))

	return style.Render(m.debug)
}

func (m model) getFooter(width int, instructions, debug string) string {
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		instructions,
		debug,
	)

	footerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return footerStyle.Render(footerContent)
}

func (m model) renderGameLayout(width, height int) string {
	header := m.getHeader(m.width)
	board := m.getBoard(m.message)
	sidebar := m.getStatusSidebar()
	ailmentsBar := m.getAilmentsBar(m.width)
	instructions := m.getInstructions(m.width)
	debug := m.getDebug(m.width)
	// Board Math
	boardHeight := height - lipgloss.Height(header) - lipgloss.Height(ailmentsBar) - 2
	boardWidth := width - SIDEBAR_WIDTH - 1

	// Style the main content area
	boardStyle := lipgloss.NewStyle().
		Width(boardWidth).
		Height(boardHeight).
		Padding(1, 2)

	boardContent := boardStyle.Render(board)

	// Create the body: main content + sidebar
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		boardContent,
		sidebar,
	)

	footer := m.getFooter(width, instructions, debug)
	// Combine header, body, and ailments bar
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		ailmentsBar,
		footer,
	)

	fullWindowStyle := lipgloss.NewStyle().
		Width(width).
		Height(height)

	return fullWindowStyle.Render(ui)
}
