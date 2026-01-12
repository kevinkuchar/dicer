package main

import (
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
const COLOR_BLUE = lipgloss.Color("#45657A")
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

func (m model) getStatusSidebar(lives int, turnNumber int) string {
	createStyle := func(background lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(background).
			Foreground(COLOR_TEXT).
			Padding(1, 2).
			Align(lipgloss.Center).
			Width(SIDEBAR_WIDTH).
			MarginTop(1).
			Bold(true)
	}

	livesStyle := createStyle(COLOR_RED)
	turnStyle := createStyle(COLOR_BLUE)

	// Build content
	livesText := fmt.Sprintf("Lives: %d", lives)
	turnText := fmt.Sprintf("Turn: %d", turnNumber)

	// Stack vertically
	sidebar := lipgloss.JoinVertical(
		lipgloss.Left,
		livesStyle.Render(livesText),
		turnStyle.Render(turnText),
	)

	return sidebar
}

func (m model) getAilmentsBar(width int, ailments *Ailments) string {
	availableWidth := width - len(ailments.remaining) - 1
	boxWidth := availableWidth / len(ailments.remaining)

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
	for i := 1; i <= len(ailments.remaining); i++ {
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

func (m model) getDice(dice []Dice) string {
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
		boxes = append(boxes, boxStyle.Render(fmt.Sprintf("%d", die.value)))
	}

	// Join boxes horizontally with a small gap
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

	// Join boxes horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
}

func (m model) getInputWindow(message string) string {
	contentStyle := lipgloss.NewStyle().
		Padding(1, 2)

	dice := m.getDice(m.turn.dice)

	turnState, _ := m.turn.stack.Top()

	choices := ""
	if turnState == GS_RollPhase {
		choices = m.getChoices()
	}

	expression := ""
	if turnState == GS_ExpressionPhase {
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

func (m model) renderGameLayout(width, height int, header, sidebar, mainContent, ailmentsBar string) string {
	sidebarWidth := SIDEBAR_WIDTH
	headerHeight := lipgloss.Height(header)
	ailmentsBarHeight := lipgloss.Height(ailmentsBar)
	availableHeight := height - headerHeight - ailmentsBarHeight
	contentWidth := width - sidebarWidth

	// Style the main content area
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(availableHeight).
		Padding(1, 2)

	styledContent := contentStyle.Render(mainContent)

	// Create the body: main content + sidebar
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styledContent,
		sidebar,
	)

	debug := m.debug

	// Combine header, body, and ailments bar
	ui := lipgloss.JoinVertical(lipgloss.Left, header, body, ailmentsBar, debug)

	mainStyle := lipgloss.NewStyle().
		Width(width).
		Height(height)

	return mainStyle.Render(ui)
}
