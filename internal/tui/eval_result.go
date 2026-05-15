package tui

/*
eval_result.go — Screen 4: AI feedback + reference answer.

Header: progress bar + topic/subtopic pills
Body:   score box (score + bar + feedback) + reference answer box
Footer: action bar (n / q)
*/

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func updateEvalResult(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "n":
			m.cardIndex++
			if m.cardIndex >= len(m.cards) {
				m.currentScreen = ScreenDone
				return m, nil
			}
			m.userAnswer = ""
			m.textarea = buildTextarea()
			m.currentScreen = ScreenCardQuestion
			return m, nil
		}
	}
	return m, nil
}

func evalResultHeader(m RootModel) string {
	return cardQuestionHeader(m)
}

func evalResultBody(m RootModel) string {
	r := m.evalResult
	scoreColor := scoreAccentColor(r.Score)
	innerW := contentWidth - 6 // borderedBox: border(2) + padding(4)

	// Score line: "AI FEEDBACK" label + score right-aligned
	scoreStr := lipgloss.NewStyle().Foreground(scoreColor).Bold(true).Render(fmt.Sprintf("%d / 5", r.Score))
	label := hintStyle.Render("AI FEEDBACK")
	gap := innerW - lipgloss.Width(label) - lipgloss.Width(scoreStr)
	if gap < 1 {
		gap = 1
	}
	scoreLine := label + strings.Repeat(" ", gap) + scoreStr

	// Score bar
	filled := (r.Score * innerW) / 5
	bar := lipgloss.NewStyle().Foreground(scoreColor).Render(strings.Repeat("━", filled)) +
		faintStyle.Render(strings.Repeat("━", innerW-filled))

	// Feedback text
	feedback := lipgloss.Wrap(r.Feedback, innerW, "")
	feedbackContent := lipgloss.NewStyle().Width(innerW).Render(
		scoreLine + "\n" + bar + "\n\n" + mutedStyle.Render(feedback),
	)
	feedbackBox := borderedBox(colorBorder).Render(feedbackContent)
	feedbackCentered := lipgloss.Place(cardInnerW, lipgloss.Height(feedbackBox), lipgloss.Center, lipgloss.Top, feedbackBox)

	// Reference answer (label + content inside the box)
	answer := lipgloss.Wrap(m.cards[m.cardIndex].Answer, innerW, "")
	refContent := lipgloss.NewStyle().Width(innerW).Render(
		hintStyle.Render("REFERENCE ANSWER") + "\n\n" + mutedStyle.Render(answer),
	)
	refBox := borderedBox(colorBorder).Render(refContent)
	refCentered := lipgloss.Place(cardInnerW, lipgloss.Height(refBox), lipgloss.Center, lipgloss.Top, refBox)

	return "\n" + feedbackCentered + "\n\n" + refCentered
}

func evalResultFooter(_ RootModel) string {
	return "\n " + actionBar("n", "next card", "q", "quit")
}

// scoreAccentColor returns a color matching the score (0–5).
func scoreAccentColor(score int) color.Color {
	switch {
	case score >= 4:
		return colorGreen
	case score >= 2:
		return colorAmber
	default:
		return colorRed
	}
}
