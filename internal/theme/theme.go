package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

type Styles struct {
	theme config.ThemeConfig
}

func New(cfg config.ThemeConfig) Styles {
	return Styles{theme: cfg}
}

func (s Styles) Panel(title string, content string, width int, height int, borderColor string, titleColor string) string {
	box := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render(lipgloss.NewStyle().
			Width(max(0, width-4)).
			Height(max(0, height-2)).
			Render(content))

	titleText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(titleColor)).
		Bold(true).
		Render(" " + title + " ")

	return overlayLine(box, titleText, 2, 0)
}

func (s Styles) CompactPanel(title string, content string, width int, borderColor string, titleColor string) string {
	body := lipgloss.NewStyle().
		Width(max(0, width-4)).
		Render(content)
	return s.Panel(title, body, width, 1, borderColor, titleColor)
}

func (s Styles) Modal(title string, content string, width int, borderColor string, titleColor string) string {
	body := lipgloss.NewStyle().Width(max(20, width)).Render(content)
	return s.Panel(title, body, max(20, width)+4, lipgloss.Height(body)+2, borderColor, titleColor)
}

func (s Styles) BranchPill(branch string) string {
	left := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.BranchBG)).Render("")
	center := lipgloss.NewStyle().
		Background(lipgloss.Color(s.theme.BranchBG)).
		Foreground(lipgloss.Color(s.theme.BranchFG)).
		Render("   " + branch + "  ")
	right := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.BranchBG)).Render("")
	return left + center + right
}

func (s Styles) Badge(label string, fg string, bg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Padding(0, 1).
		Render(strings.ToUpper(label))
}

func (s Styles) Error(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.ErrorFG)).Bold(true).Render(text)
}

func (s Styles) Success(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.SuccessFG)).Bold(true).Render(text)
}

func (s Styles) Muted(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.StatusFG)).Render(text)
}

func (s Styles) Text(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.TextFG)).Render(text)
}

func (s Styles) Accent(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.AccentFG)).Bold(true).Render(text)
}

func (s Styles) AccentAlt(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.AccentAltFG)).Render(text)
}

func (s Styles) Warning(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.WarningFG)).Render(text)
}

func (s Styles) ThemeBadge(label string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.BadgeFG)).
		Bold(true).
		Render(label)
}

func (s Styles) Dot(color string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("")
}

func overlayLine(base string, overlay string, x int, y int) string {
	lines := strings.Split(base, "\n")
	if y < 0 || y >= len(lines) {
		return base
	}
	lines[y] = replaceAtWidth(lines[y], overlay, x)
	return strings.Join(lines, "\n")
}

func replaceAtWidth(base string, overlay string, start int) string {
	baseWidth := ansi.StringWidth(base)
	if start < 0 {
		start = 0
	}
	if start > baseWidth {
		start = baseWidth
	}
	end := min(baseWidth, start+ansi.StringWidth(overlay))
	return ansi.Cut(base, 0, start) + overlay + ansi.Cut(base, end, baseWidth)
}
