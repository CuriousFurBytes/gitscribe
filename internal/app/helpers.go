package app

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

var timeNow = time.Now

func titleTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return titleTickMsg{}
	})
}

func withTimeout(timeout time.Duration, fn func(context.Context) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return fn(ctx)
	}
}

func (m *Model) beginOperation(label string) {
	m.operationStatus = operationStatus{label: label}
}

func (m *Model) clearOperation() {
	m.operationStatus = operationStatus{}
}

func (m *Model) closeModal() {
	m.modal.visible = false
	m.modal.kind = modalNone
	m.modal.title = ""
	m.modal.body = ""
	m.modal.loading = false
	m.modal.success = false
	m.modal.refreshRepo = false
	m.modal.selecting = false
	m.modal.hooks = nil
	m.modal.hookIndex = 0
	m.modal.viewport.SetContent("")
	m.modal.input.SetValue("")
}

func normalizeKeybinding(key string) string {
	key = strings.TrimSpace(key)
	// Lowercase modifier combinations (ctrl+x, alt+x, shift+x) but preserve
	// case for single characters so 'a' and 'A' can be bound separately.
	if strings.Contains(key, "+") {
		return strings.ToLower(key)
	}
	return key
}

func normalizedKeybindings(keys []string) []string {
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		key = normalizeKeybinding(key)
		if key == "" {
			continue
		}
		normalized = append(normalized, key)
	}
	return normalized
}

func (m *Model) keybindingsForAction(action string) []string {
	if keys, ok := m.cfg.Keybindings[action]; ok && len(keys) > 0 {
		return normalizedKeybindings(keys)
	}
	return normalizedKeybindings(config.DefaultKeybindings()[action])
}

func (m *Model) matchesKeybinding(action string, msg tea.KeyMsg) bool {
	key := normalizeKeybinding(msg.String())
	for _, candidate := range m.keybindingsForAction(action) {
		if candidate == key {
			return true
		}
	}
	return false
}

func (m *Model) colorizeDiff(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"), strings.HasPrefix(line, "@@"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(true).Render(line)
		case strings.HasPrefix(line, "index "), strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentAltFG)).Render(line)
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.SuccessFG)).Render(line)
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ErrorFG)).Render(line)
		case strings.HasPrefix(line, "Staged changes"), strings.HasPrefix(line, "Unstaged changes"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.WarningFG)).Bold(true).Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) colorizeLog(content string) string {
	if strings.Contains(content, "diff --git") || strings.Contains(content, "@@") {
		return m.colorizeDiff(content)
	}
	return m.colorizeGitOutput(content)
}

func (m *Model) colorizeGitOutput(content string) string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG))
	success := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.SuccessFG))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ErrorFG))
	warning := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.WarningFG))

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "http://"):
			lines[i] = accent.Render(line)
		case strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "]"):
			lines[i] = accent.Render(line)
		case strings.Contains(trimmed, "insertion") || strings.Contains(trimmed, "deletion"):
			lines[i] = colorizeStatSummary(line, m.cfg.Theme.SuccessFG, m.cfg.Theme.ErrorFG)
		case strings.HasPrefix(trimmed, "create mode"):
			lines[i] = success.Render(line)
		case strings.HasPrefix(trimmed, "delete mode"):
			lines[i] = errStyle.Render(line)
		case strings.HasPrefix(trimmed, "rename"):
			lines[i] = warning.Render(line)
		case strings.Contains(trimmed, "->") && strings.Contains(trimmed, ".."):
			lines[i] = accent.Render(line)
		case strings.Contains(line, "|") && (strings.Contains(line, "+") || strings.Contains(line, "-")):
			lines[i] = colorizeFileStat(line, m.cfg.Theme.SuccessFG, m.cfg.Theme.ErrorFG)
		}
	}
	return strings.Join(lines, "\n")
}

func colorizeStatSummary(line, successColor, errorColor string) string {
	parts := strings.Split(line, ",")
	for i, part := range parts {
		if strings.Contains(part, "insertion") {
			parts[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(successColor)).Render(part)
		} else if strings.Contains(part, "deletion") {
			parts[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(part)
		}
	}
	return strings.Join(parts, ",")
}

func colorizeFileStat(line, successColor, errorColor string) string {
	idx := strings.LastIndex(line, "|")
	if idx < 0 {
		return line
	}
	left := line[:idx+1]
	right := line[idx+1:]
	suc := lipgloss.NewStyle().Foreground(lipgloss.Color(successColor))
	err := lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor))
	var colored strings.Builder
	for _, ch := range right {
		switch ch {
		case '+':
			colored.WriteString(suc.Render("+"))
		case '-':
			colored.WriteString(err.Render("-"))
		default:
			colored.WriteRune(ch)
		}
	}
	return left + colored.String()
}

func trimLog(text string, maxLines int) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}

func truncateANSI(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(text) <= width {
		return text
	}
	if width == 1 {
		return ansi.Cut(text, 0, 1)
	}
	return ansi.Cut(text, 0, width-1) + "…"
}

func overlayCentered(base string, overlay string, width int, height int) string {
	baseLines := normalizeLines(lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base), height, width)
	overlayLines := strings.Split(overlay, "\n")
	overlayWidth := 0
	for _, line := range overlayLines {
		overlayWidth = max(overlayWidth, ansi.StringWidth(line))
	}
	startX := max(0, (width-overlayWidth)/2)
	startY := max(0, (height-len(overlayLines))/2)

	for i, line := range overlayLines {
		y := startY + i
		if y < 0 || y >= len(baseLines) {
			continue
		}
		baseLines[y] = overlayLineAt(baseLines[y], line, startX, width)
	}
	return strings.Join(baseLines, "\n")
}

func overlayLineAt(base string, overlay string, start int, totalWidth int) string {
	base = padANSI(base, totalWidth)
	baseWidth := ansi.StringWidth(base)
	end := min(baseWidth, start+ansi.StringWidth(overlay))
	return ansi.Cut(base, 0, start) + overlay + ansi.Cut(base, end, baseWidth)
}

func normalizeLines(content string, height int, width int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) < height {
		padding := make([]string, height-len(lines))
		lines = append(lines, padding...)
	} else if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = padANSI(line, width)
	}
	return lines
}

func padANSI(text string, width int) string {
	visible := ansi.StringWidth(text)
	if visible >= width {
		return ansi.Cut(text, 0, width)
	}
	return text + strings.Repeat(" ", width-visible)
}

func clamp(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func verticalScrollbar(height, total, start int, trackColor, thumbColor string) []string {
	bar := make([]string, height)
	if total <= height {
		return bar
	}
	thumbHeight := max(1, height*height/total)
	maxStart := total - height
	thumbPos := 0
	if maxStart > 0 {
		thumbPos = (start * (height - thumbHeight)) / maxStart
	}
	track := lipgloss.NewStyle().Foreground(lipgloss.Color(trackColor))
	thumb := lipgloss.NewStyle().Foreground(lipgloss.Color(thumbColor))
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			bar[i] = thumb.Render("█")
		} else {
			bar[i] = track.Render("│")
		}
	}
	return bar
}
