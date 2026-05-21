package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openShellModal() {
	m.modal.visible = true
	m.modal.kind = modalShell
	m.modal.title = "Shell"
	m.modal.loading = false
	m.modal.success = false
	m.modal.selecting = false
	m.modal.input.SetValue("")
	m.modal.input.Focus()
	m.modal.viewport.SetContent("Run a shell command without leaving GitScribe.")
	m.modal.viewport.GotoTop()
}

func (m *Model) renderShellModal() string {
	output := m.modal.viewport.View()
	if m.modal.loading {
		output = m.spinner.View() + " Running...\n\n" + output
	}
	body := strings.Join([]string{
		m.styles.Muted("Run shell commands inside the current repository."),
		"",
		m.modal.input.View(),
		"",
		output,
		"",
		m.renderShortcutHints(limitShortcutHints([]shortcutHint{
			{Key: "Enter", Text: "Run"},
			{Key: "Esc", Text: "Close"},
			{Key: "↑↓", Text: "Scroll"},
		})...),
	}, "\n")
	return m.styles.Modal("Shell", body, min(92, m.width-8), m.cfg.Theme.LogsBorder, m.cfg.Theme.LogsTitle)
}

func (m *Model) handleShellModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc", "q":
		m.closeModal()
	case "enter":
		command := strings.TrimSpace(m.modal.input.Value())
		if command != "" && !m.modal.loading {
			m.modal.loading = true
			cmds = append(cmds, runShellCommandCmd(m.repo.Root, command))
		}
	case "up", "k":
		m.modal.viewport.LineUp(1)
	case "down", "j":
		m.modal.viewport.LineDown(1)
	case "pgup":
		m.modal.viewport.LineUp(10)
	case "pgdown":
		m.modal.viewport.LineDown(10)
	default:
		m.modal.input, cmd = m.modal.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
