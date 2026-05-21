package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openPRURLModal opens a modal displaying the URL of a freshly-created pull
// request. The modal is dismissed via Esc or Enter.
func (m *Model) openPRURLModal(url string) {
	m.modal.visible = true
	m.modal.kind = modalPRURL
	m.modal.title = "Pull request created"
	m.modal.body = url
	m.modal.loading = false
	m.modal.success = true
	m.modal.refreshRepo = false
	m.modal.selecting = false
	m.modal.returnTo = m.screen
}

func (m *Model) renderPRURLModal() string {
	url := strings.TrimSpace(m.modal.body)

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(true)

	lines := []string{
		m.styles.Success("Pull request created successfully."),
		"",
		labelStyle.Render("URL:"),
		urlStyle.Render(url),
		"",
		m.renderShortcutHints(shortcutHint{Key: "Esc/Enter", Text: "Close"}),
	}

	width := min(96, max(48, m.width-8))
	return m.styles.Modal("Pull request created", strings.Join(lines, "\n"), width, m.cfg.Theme.ConfirmBorder, m.cfg.Theme.ConfirmTitle)
}

func (m *Model) handlePRURLModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		m.closeModal()
	}
	return m, tea.Batch(cmds...)
}
