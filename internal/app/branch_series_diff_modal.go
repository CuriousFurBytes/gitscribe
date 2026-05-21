package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const branchSeriesDiffModalTitle = "Branch series diff"

// openBranchSeriesDiffModal puts the model into a loading-state modal that
// shows the diff of the entire branch series against its base. The diff is
// fetched asynchronously via loadBranchSeriesDiffCmd and rendered when a
// branchSeriesDiffLoadedMsg arrives.
func (m *Model) openBranchSeriesDiffModal() {
	m.modal.visible = true
	m.modal.kind = modalBranchSeriesDiff
	m.modal.title = branchSeriesDiffModalTitle
	m.modal.body = ""
	m.modal.loading = true
	m.modal.success = false
	m.modal.refreshRepo = false
	m.modal.selecting = false
	m.modal.returnTo = m.screen
	m.modal.viewport.SetContent("")
	m.modal.viewport.GotoTop()
}

func (m *Model) setBranchSeriesDiffContent(content string, err error) {
	m.modal.loading = false
	if err != nil {
		m.modal.success = false
		m.modal.viewport.SetContent("Could not load branch series diff: " + err.Error())
		m.modal.viewport.GotoTop()
		return
	}
	display := strings.TrimRight(content, "\n")
	if strings.TrimSpace(display) == "" {
		m.modal.success = true
		m.modal.viewport.SetContent("No changes between the current branch and its base.")
		m.modal.viewport.GotoTop()
		return
	}
	m.modal.success = true
	m.modal.viewport.SetContent(m.colorizeDiff(display))
	m.modal.viewport.GotoTop()
}

func (m *Model) renderBranchSeriesDiffModal() string {
	header := m.modal.title
	if m.modal.loading {
		header = lipgloss.JoinHorizontal(lipgloss.Top, m.spinner.View()+" ", m.styles.Accent(header))
	} else if m.modal.success {
		header = m.styles.Success(header)
	} else {
		header = m.styles.Error(header)
	}
	body := m.modal.viewport.View()
	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	panel := m.styles.Panel(branchSeriesDiffModalTitle, content, min(120, max(64, m.width-6)), max(14, m.height-8), m.cfg.Theme.LogsBorder, m.cfg.Theme.LogsTitle)
	footer := m.renderShortcutHints(limitShortcutHints([]shortcutHint{
		{Key: "Esc/Enter/q", Text: "Close"},
		{Key: "↑↓", Text: "Scroll"},
		{Key: "PgUp/PgDn", Text: "Fast scroll"},
	})...)
	return lipgloss.JoinVertical(lipgloss.Left, panel, lipgloss.NewStyle().Width(lipgloss.Width(panel)).Align(lipgloss.Right).Render(footer))
}

func (m *Model) handleBranchSeriesDiffModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		m.closeModal()
		return m, tea.Batch(cmds...)
	case "up", "k":
		m.modal.viewport.LineUp(1)
	case "down", "j":
		m.modal.viewport.LineDown(1)
	case "pgup":
		m.modal.viewport.LineUp(10)
	case "pgdown":
		m.modal.viewport.LineDown(10)
	}
	return m, tea.Batch(cmds...)
}
