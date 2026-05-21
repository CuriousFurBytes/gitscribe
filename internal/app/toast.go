package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// successToastText returns the short toast string for a successful operation,
// or "" when the operation should not produce a toast.
func successToastText(msg operationResultMsg) string {
	if !msg.success {
		return ""
	}
	switch {
	case msg.clearCommit:
		return "Commit created"
	case msg.clearPR:
		return "Pull request created"
	}
	return ""
}

// toastExpireCmd schedules a toastExpireMsg after toastDuration. It uses
// time.AfterFunc-driven tea.Tick so the message arrives on the Bubble Tea
// dispatch loop.
func toastExpireCmd() tea.Cmd {
	return tea.Tick(toastDuration, func(time.Time) tea.Msg {
		return toastExpireMsg{}
	})
}

func (m *Model) renderToast(width int) string {
	_ = width
	if !m.toast.visible() {
		return ""
	}
	color := m.cfg.Theme.SuccessFG
	if m.toast.kind == toastInfo {
		color = m.cfg.Theme.AccentFG
	}
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Padding(0, 1)
	return style.Render(m.toast.text)
}
