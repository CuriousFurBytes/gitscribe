package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) renderConfirmModal() string {
	confirmHints := m.renderShortcutHints(
		shortcutHint{Key: "Enter/Y", Text: "Confirm"},
		shortcutHint{Key: "Esc/N", Text: "Cancel"},
	)
	return m.styles.Modal(m.confirm.Title, m.confirm.Message+"\n\n"+confirmHints, 56, m.cfg.Theme.ConfirmBorder, m.cfg.Theme.ConfirmTitle)
}

func (m *Model) handleConfirmKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.confirm = nil
	case "enter", "y":
		switch m.confirm.Action {
		case confirmCommit:
			m.beginOperation("Committing")
		case confirmPR:
			m.beginOperation("Creating PR")
		case confirmDiscard:
			m.beginOperation("Discarding")
		case confirmReset:
			m.beginOperation("Resetting")
		case confirmStash:
			m.beginOperation("Stashing")
		}
		cmds = append(cmds, m.confirmActionCmd())
		m.confirm = nil
	}
	return m, tea.Batch(cmds...)
}
