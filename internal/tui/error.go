package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func updateError(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd) {
	return m, nil
}

func errorHeader(_ RootModel) lipgloss.Style {
	refError := lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render("ERROR!")
	return styledBox(CardParams{BgColor: colorBase, Padding: []int{1, 1}}).SetString(refError)
}

func errorBody(m RootModel) lipgloss.Style {
	refErrorBody := lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render(m.errMsg)
	return styledBox(CardParams{BgColor: colorBase, Padding: []int{1, 1}}).SetString(refErrorBody)
}

func errorFooter(_ RootModel) lipgloss.Style { return lipgloss.NewStyle() }
