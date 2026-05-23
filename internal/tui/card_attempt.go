package tui

/*
card_attempt.go — Screen 3: type answer to current card.

Header: progress bar + topic/subtopic pills (same as question screen)
Body:   textarea
Footer: action bar (ctrl+s / esc)
*/

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/codehia/goflash/internal/ai"
)

func updateCardAttempt(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentScreen = ScreenCardQuestion
			return m, nil
		case "ctrl+s":
			m.userAnswer = m.textarea.Value()
			card := m.cards[m.cardIndex]
			userAnswer := m.userAnswer
			return m, func() tea.Msg {
				result, err := ai.Evaluate(card.Question, card.Answer, userAnswer)
				return EvalResultMsg{result: result, err: err}
			}
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func cardAttemptHeader(m RootModel) lipgloss.Style {
	return cardQuestionHeader(m)
}

func cardAttemptBody(m RootModel) lipgloss.Style {
	var content string
	if m.cardIndex < len(m.cards) {
		box := styledBox(CardParams{BorderColor: colorFlamingo, Padding: []int{0, 2}}).Render(m.textarea.View())
		centered := lipgloss.Place(cardInnerW, lipgloss.Height(box), lipgloss.Center, lipgloss.Top, box)
		content = "\n" + hintStyle.Render("your answer") + "\n\n" + centered
	}
	return styledBox(CardParams{BgColor: colorBase}).SetString(content)
}

func cardAttemptFooter(_ RootModel) lipgloss.Style {
	return styledBox(CardParams{BgColor: colorBase, Padding: []int{1, 1}}).SetString(actionBar("ctrl+s", "submit", "esc", "back"))
}
