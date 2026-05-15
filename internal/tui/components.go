// Package tui — shared view helpers and styles.
//
// # Screen pattern
//
// Each screen lives in its own file with three standalone functions:
//
//	initScreenName(m RootModel) tea.Cmd
//	updateScreenName(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd)
//	screenNameView(m RootModel) string
package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/codehia/goflash/internal/types"
)

// ── Color palette — Catppuccin Mocha ────────────────────────────────
// https://catppuccin.com/palette

var (
	// Base layers (what we own)
	colorBase    = lipgloss.Color("#1e1e2e") // card bg
	colorSurface = lipgloss.Color("#313244") // inner elements

	// Borders
	colorBorder    = lipgloss.Color("#585b70") // Surface2
	colorBorderHov = lipgloss.Color("#6c7086") // Overlay0

	// Text
	colorText  = lipgloss.Color("#cdd6f4") // Text
	colorMuted = lipgloss.Color("#9399b2") // Overlay2
	colorHint  = lipgloss.Color("#7f849c") // Overlay1
	colorFaint = lipgloss.Color("#6c7086") // Overlay0

	// Accents
	colorPurple = lipgloss.Color("#cba6f7") // Mauve
	colorTeal   = lipgloss.Color("#94e2d5") // Teal
	colorAmber  = lipgloss.Color("#fab387") // Peach
	colorGreen  = lipgloss.Color("#a6e3a1") // Green
	colorRed    = lipgloss.Color("#f38ba8") // Red
)

// ── Text styles ─────────────────────────────────────────────────────

var (
	boldStyle   = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	hintStyle   = lipgloss.NewStyle().Foreground(colorHint)
	faintStyle  = lipgloss.NewStyle().Foreground(colorFaint)
	purpleStyle = lipgloss.NewStyle().Foreground(colorPurple)
)

// ── Pill (background only, single line) ─────────────────────────────

func makePill(text string, fg, bg color.Color) string {
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Padding(0, 1).Render(text)
}
func purplePill(text string) string { return makePill(text, colorPurple, colorSurface) }
func tealPill(text string) string   { return makePill(text, colorTeal, colorSurface) }

// ── Outer box ───────────────────────────────────────────────────────

// contentWidth is the inner text width of topic row boxes.
// Derived from card inner width minus box overhead (border=2, padding=4, margin=2).
const contentWidth = cardInnerW - 8 // = 70

func borderedBox(borderColor color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 2)
}

// ── Layout helpers ──────────────────────────────────────────────────

func visibleWindow(cursor, total, maxVisible int) (start, end int) {
	if total <= maxVisible {
		return 0, total
	}
	half := maxVisible / 2
	start = cursor - half
	if start < 0 {
		start = 0
	}
	end = start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return
}

func leftRight(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func progressBar(current, total int) string {
	if total == 0 {
		return ""
	}
	pct := float64(current) / float64(total)
	barWidth := cardInnerW - 2 - len(fmt.Sprintf("%d / %d", current, total))
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	filledStr := purpleStyle.Render(strings.Repeat("━", filled))
	emptyStr := faintStyle.Render(strings.Repeat("━", barWidth-filled))
	counter := purpleStyle.Render(fmt.Sprintf("%d / %d", current, total))
	return filledStr + emptyStr + " " + counter
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

type CardParams struct {
	TermW, TermH         int
	BorderColor          color.Color
	BgColor              color.Color
	Header, Body, Footer string
}

func renderCard(p CardParams) string {
	hasHeader := strings.TrimSpace(p.Header) != ""
	hasFooter := strings.TrimSpace(p.Footer) != ""

	// body fills remaining height after header/footer/dividers
	bh := cardInnerH
	if hasHeader {
		bh -= headerH + 1
	}
	if hasFooter {
		bh -= footerH + 1
	}

	zone := func(content string, h int) string {
		return lipgloss.NewStyle().
			Width(cardInnerW).Height(h).
			Background(p.BgColor).
			Render(content)
	}

	divider := zone(faintStyle.Render(strings.Repeat("─", cardInnerW)), 1)

	var sections []string
	if hasHeader {
		sections = append(sections, zone(p.Header, headerH), divider)
	}
	sections = append(sections, zone(p.Body, bh))
	if hasFooter {
		sections = append(sections, divider, zone(p.Footer, footerH))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderColor).
		Background(p.BgColor).
		Render(content)

	return lipgloss.Place(p.TermW, p.TermH, lipgloss.Center, lipgloss.Center, card)
}

// pillRows lays pills into rows that fit within maxW, padding each row to maxW.
func pillRows(pills []string, maxW int) string {
	var rows []string
	var cur []string
	curW := 0
	for _, p := range pills {
		pw := lipgloss.Width(p)
		sep := 0
		if len(cur) > 0 {
			sep = 1
		}
		if len(cur) > 0 && curW+sep+pw > maxW {
			row := strings.Join(cur, " ")
			rows = append(rows, row+strings.Repeat(" ", maxW-lipgloss.Width(row)))
			cur = []string{p}
			curW = pw
		} else {
			cur = append(cur, p)
			curW += sep + pw
		}
	}
	if len(cur) > 0 {
		row := strings.Join(cur, " ")
		rows = append(rows, row+strings.Repeat(" ", maxW-lipgloss.Width(row)))
	}
	return strings.Join(rows, "\n")
}

// ── Topic row ───────────────────────────────────────────────────────

func renderTopicRow(topic types.TopicSummary, selected bool) string {
	cursor := faintStyle.Render("  ")
	if selected {
		cursor = purpleStyle.Render("▶ ")
	}

	const pillColW = 14
	pill := purplePill(fmt.Sprintf("%d cards", topic.CardCount))
	nameCol := lipgloss.NewStyle().Width(contentWidth - pillColW).MaxWidth(contentWidth - pillColW).Render(cursor + boldStyle.Render(topic.Name))
	pillCol := lipgloss.NewStyle().Width(pillColW).Align(lipgloss.Right).Render(pill)
	line1 := lipgloss.JoinHorizontal(lipgloss.Top, nameCol, pillCol)

	var line2 string
	if len(topic.Subtopics) > 0 {
		var pills []string
		for _, n := range topic.Subtopics {
			pills = append(pills, tealPill(n))
		}
		if len(topic.Subtopics) > 4 {
			pills = pills[:4]
			pills = append(pills, hintStyle.Render(fmt.Sprintf("+%d more", len(topic.Subtopics)-4)))
		}
		line2 = pillRows(pills, contentWidth)
	}

	content := line1
	if line2 != "" {
		divider := faintStyle.Render(strings.Repeat("─", contentWidth))
		content = lipgloss.JoinVertical(lipgloss.Left, line1, divider, line2)
	}

	bc := colorBorder
	if selected {
		bc = colorPurple
	}
	row := borderedBox(bc).Render(content)
	return lipgloss.Place(cardInnerW, lipgloss.Height(row), lipgloss.Center, lipgloss.Top, row)
}
