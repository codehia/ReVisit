package tui

/*
card_question.go — Screen 2: show current card question.

Header: progress bar + topic pill
Body:   question in bordered box + hint
Footer: action bar (space / q)
*/

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func updateCardQuestion(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", "space":
			if len(m.cards) == 0 {
				return m, nil
			}
			card := m.cards[m.cardIndex]
			return m, func() tea.Msg {
				return CardSelectedMsg{cardID: *card.ID}
			}
		}
	}
	return m, nil
}

func cardQuestionHeader(m RootModel) string {
	bar := progressBar(m.cardIndex+1, len(m.cards))
	pills := tealPill(m.topicName)
	if m.cardIndex < len(m.cards) {
		for _, s := range m.cards[m.cardIndex].Subtopics {
			pills += " " + tealPill(s)
		}
	}
	return "\n " + bar + "\n " + pills
}

func cardQuestionBody(m RootModel) string {
	if len(m.cards) == 0 {
		return "\n " + mutedStyle.Render("No cards due.")
	}
	card := m.cards[m.cardIndex]
	question := boldStyle.Render(lipgloss.Wrap(card.Question, contentWidth-4, ""))
	box := borderedBox(colorBorder).Render(
		lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(question),
	)
	centered := lipgloss.Place(cardInnerW, lipgloss.Height(box), lipgloss.Center, lipgloss.Top, box)
	hint := hintStyle.Render("think through your answer before continuing")
	hintLine := lipgloss.NewStyle().Width(cardInnerW).Align(lipgloss.Center).Render(hint)
	return "\n" + centered + "\n\n" + hintLine
}

func cardQuestionFooter(_ RootModel) string {
	return "\n " + actionBar("space", "answer", "q", "topics")
}

