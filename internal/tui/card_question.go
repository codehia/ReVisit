package tui

/*
card_question.go — Screen 2: show current card question.

Header: progress bar + topic pill
Body:   question in bordered box + hint
Footer: action bar (space / q)
*/

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/codehia/ReVisit/internal/store"
)

func initCards(m RootModel) tea.Cmd {
	return func() tea.Msg {
		cards, err := store.GetCardsForTopic(m.db, m.selectedTopicID)
		if err != nil {
			return cardsLoadedMsg{err: err}
		}
		return cardsLoadedMsg{cards: cards}
	}
}

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

func cardQuestionHeader(m RootModel) lipgloss.Style {
	bar := progressBar(m.cardIndex+1, len(m.cards), cardInnerW-2)
	var pills strings.Builder
	pills.WriteString(tealPill(m.topicName))
	if m.cardIndex < len(m.cards) {
		for _, s := range m.cards[m.cardIndex].Subtopics {
			pills.WriteString(" " + tealPill(s))
		}
	}
	return styledBox(CardParams{BgColor: colorBase, Padding: []int{1, 1}}).SetString(bar + "\n" + pills.String())
}

func cardQuestionBody(m RootModel) lipgloss.Style {
	var content string
	if len(m.cards) == 0 {
		content = "\n" + mutedStyle.Render("No cards due.")
	} else {
		card := m.cards[m.cardIndex]
		question := boldStyle.Render(lipgloss.Wrap(card.Question, contentWidth-4, ""))
		box := styledBox(CardParams{BorderColor: colorBorder, Padding: []int{0, 2}}).Render(
			lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(question),
		)
		centered := lipgloss.Place(cardInnerW, lipgloss.Height(box), lipgloss.Center, lipgloss.Top, box)
		hint := hintStyle.Render("think through your answer before continuing")
		hintLine := lipgloss.NewStyle().Width(cardInnerW).Align(lipgloss.Center).Render(hint)
		content = "\n" + centered + "\n\n" + hintLine
	}
	return styledBox(CardParams{BgColor: colorBase}).SetString(content)
}

func cardQuestionFooter(_ RootModel) lipgloss.Style {
	return styledBox(CardParams{BgColor: colorBase, Padding: []int{1, 1}}).SetString(actionBar("space", "answer", "q", "topics"))
}
