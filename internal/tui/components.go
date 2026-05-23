// Package tui — shared view helpers and styles.
//
// # Screen pattern
//
// Each screen lives in its own file:
//
//	updateScreenName(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd)
//	screenNameHeader/Body/Footer(m RootModel) lipgloss.Style
//
// Screens that load data on entry also define initScreenName(m RootModel) tea.Cmd.
package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	cardInnerW     = cardWidth - 2
	cardInnerH     = cardHeight - 2
	contentWidth   = cardWidth - 10
	roundedBorderH = 2 // lipgloss.RoundedBorder adds 1 char left + 1 right
)

// ── Color palette — Catppuccin Mocha ────────────────────────────────
// https://catppuccin.com/palette

var (
	// Base layers (what we own)
	colorBase    = lipgloss.Color("#1e1e2e") // card bg
	colorSurface = lipgloss.Color("#313244") // inner elements

	// Borders
	colorBorder = lipgloss.Color("#585b70") // Surface2

	// Text
	colorText  = lipgloss.Color("#cdd6f4") // Text
	colorMuted = lipgloss.Color("#9399b2") // Overlay2
	colorHint  = lipgloss.Color("#7f849c") // Overlay1
	colorFaint = lipgloss.Color("#6c7086") // Overlay0

	// Accents
	colorFlamingo = lipgloss.Color("#f2cdcd") // Flamingo
	colorSapphire = lipgloss.Color("#74c7ec") // Sapphire
	colorAmber    = lipgloss.Color("#fab387") // Peach
	colorGreen    = lipgloss.Color("#a6e3a1") // Green
	colorRed      = lipgloss.Color("#f38ba8") // Red
)

// ── Text styles ─────────────────────────────────────────────────────

var (
	boldStyle   = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	hintStyle   = lipgloss.NewStyle().Foreground(colorHint)
	faintStyle  = lipgloss.NewStyle().Foreground(colorFaint)
	purpleStyle = lipgloss.NewStyle().Foreground(colorFlamingo)
)

// ── Pill (background only, single line) ─────────────────────────────

func makePill(text string, fg, bg color.Color) string {
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Padding(0, 1).Render(text)
}
func purplePill(text string) string { return makePill(text, colorFlamingo, colorSurface) }
func tealPill(text string) string   { return makePill(text, colorSapphire, colorSurface) }

// ── styledBox ────────────────────────────────────────────────────────

// styledBox returns a lipgloss.Style configured from p.

func styledBox(p CardParams) lipgloss.Style {
	s := lipgloss.NewStyle().Background(p.BgColor)
	if p.BorderColor != nil {
		s = s.Border(lipgloss.RoundedBorder()).BorderForeground(p.BorderColor)
	}
	if len(p.Padding) > 0 {
		s = s.Padding(p.Padding...)
	}
	if len(p.Margins) > 0 {
		s = s.Margin(p.Margins...)
	}
	return s
}

func centerCard(termW, termH int, card string) string {
	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, card)
}

// ── Too small screen ────────────────────────────────────────────────

func renderTooSmall(w, h int) string {
	msg := lipgloss.JoinVertical(lipgloss.Center,
		boldStyle.Render("terminal too small"),
		mutedStyle.Render(fmt.Sprintf("current   %d × %d", w, h)),
		mutedStyle.Render(fmt.Sprintf("required  %d × %d", cardWidth, cardHeight)),
	)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}

// ── Layout helpers ──────────────────────────────────────────────────

func progressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	pct := float64(current) / float64(total)
	counter := fmt.Sprintf("%d / %d", current, total)
	barWidth := width - len(counter) - 1
	filled := min(int(pct*float64(barWidth)), barWidth)
	filledStr := purpleStyle.Render(strings.Repeat("━", filled))
	emptyStr := faintStyle.Render(strings.Repeat("━", barWidth-filled))
	return filledStr + emptyStr + " " + purpleStyle.Render(counter)
}

func actionBar(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		key := purpleStyle.Render(pairs[i])
		label := hintStyle.Render(pairs[i+1])
		parts = append(parts, fmt.Sprintf("[ %s ] %s", key, label))
	}
	return strings.Join(parts, "    ")
}

// ── Card (outer wrapper) ─────────────────────────────────────────────

func renderCard(width int, borderColor color.Color, header, body, footer lipgloss.Style) string {
	outer := styledBox(CardParams{BorderColor: borderColor})
	innerW := width - roundedBorderH
	divider := faintStyle.Render(strings.Repeat("─", innerW))

	h := header.Width(innerW).String()
	b := body.Width(innerW).String()
	f := footer.Width(innerW).String()

	var sections []string
	if strings.TrimSpace(h) != "" {
		sections = append(sections, h)
	}
	if strings.TrimSpace(b) != "" {
		sections = append(sections, divider, b)
	}
	if strings.TrimSpace(f) != "" {
		sections = append(sections, divider, f)
	}

	return outer.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}
