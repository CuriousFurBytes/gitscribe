package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

const maxBottomHints = 5

func (m *Model) renderBottomStatusLine(width int) string {
	parts := []string{}
	if extra := git.AheadBehindLabel(m.status.RepoStatus); extra != "" {
		parts = append(parts, m.styles.Muted(extra))
	}
	if m.operationStatus.label != "" {
		parts = append(parts, strings.TrimSpace(m.spinner.View())+" "+m.operationStatus.label)
	}
	if m.notice != "" {
		parts = append(parts, m.styles.Error(singleLine(m.notice)))
	}
	if len(parts) == 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(parts, "   "))
}

func (m *Model) renderShortcutHints(hints ...shortcutHint) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutKeyFG)).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))
	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		parts = append(parts, keyStyle.Render(hint.Key)+" "+textStyle.Render(hint.Text))
	}
	return strings.Join(parts, "   ")
}

func limitShortcutHints(hints []shortcutHint) []shortcutHint {
	if len(hints) <= maxBottomHints {
		return hints
	}
	pinned := hints[len(hints)-1]
	return append(hints[:maxBottomHints-1:maxBottomHints-1], pinned)
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
