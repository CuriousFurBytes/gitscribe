package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openStashModal() {
	m.modal.visible = true
	m.modal.kind = modalStash
	m.modal.title = "Stash changes"
	m.modal.loading = false
	m.modal.success = false
	m.modal.selecting = false
	m.modal.input.SetValue("")
	m.modal.input.Placeholder = "Stash name (optional)"
	m.modal.input.Focus()
}

func (m *Model) renderStashModal() string {
	inputLine := m.modal.input.View()
	body := strings.Join([]string{
		"Enter a name for this stash (leave blank for default).",
		"",
		inputLine,
		"",
		m.renderShortcutHints(
			shortcutHint{Key: "Enter", Text: "Stash"},
			shortcutHint{Key: "Esc", Text: "Cancel"},
		),
	}, "\n")
	return m.styles.Modal("Stash changes", body, min(60, m.width-8), m.cfg.Theme.ConfirmBorder, m.cfg.Theme.ConfirmTitle)
}

func (m *Model) handleStashModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := strings.TrimSpace(m.modal.input.Value())
		m.closeModal()
		cmds = append(cmds, stashCmd(m.repo.Root, name))
		return m, tea.Batch(cmds...)
	case tea.KeyEsc:
		m.closeModal()
		return m, tea.Batch(cmds...)
	default:
		m.modal.input, _ = m.modal.input.Update(msg)
		return m, tea.Batch(cmds...)
	}
}
