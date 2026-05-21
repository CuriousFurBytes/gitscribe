package app

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openAppLogsModal() {
	m.modal.visible = true
	m.modal.kind = modalLogs
	m.modal.title = "Application Logs"
	m.modal.loading = false
	m.modal.success = true
	m.modal.refreshRepo = false
	m.modal.returnTo = m.screen
	m.modal.selecting = false

	content := ""
	if m.logFilePath != "" {
		data, err := os.ReadFile(m.logFilePath)
		if err == nil {
			content = string(data)
		} else {
			content = "Could not read log file: " + err.Error()
		}
	} else {
		content = "No log file available."
	}

	content = trimLog(content, m.cfg.Logs.MaxLines)
	m.modal.viewport.SetContent(content)
	m.modal.viewport.GotoBottom()
}

func (m *Model) openLogsModal(msg operationResultMsg) {
	m.modal.visible = true
	if msg.modalKind != modalNone {
		m.modal.kind = msg.modalKind
	} else {
		m.modal.kind = modalLogs
	}
	m.modal.title = msg.title
	m.modal.body = ""
	m.modal.loading = false
	m.modal.success = msg.success
	m.modal.refreshRepo = msg.refreshRepo
	m.modal.returnTo = m.screen
	m.modal.selecting = false
	content := trimLog(msg.output, m.cfg.Logs.MaxLines)
	if strings.TrimSpace(content) == "" {
		if msg.success {
			content = "Operation completed successfully."
		} else if strings.TrimSpace(msg.stderr) != "" {
			content = trimLog(msg.stderr, m.cfg.Logs.MaxLines)
		} else if msg.err != nil {
			content = msg.err.Error()
		}
	}
	m.modal.viewport.SetContent(m.colorizeLog(content))
	m.modal.viewport.GotoTop()
}

func (m *Model) renderScrollableModal() string {
	header := m.modal.title
	if m.modal.success {
		header = m.styles.Success(header)
	} else if !m.modal.loading {
		header = m.styles.Error(header)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.modal.viewport.View())
	if m.modal.loading {
		content = lipgloss.JoinVertical(lipgloss.Left, header, m.spinner.View()+" "+m.modal.body, "", m.modal.viewport.View())
	}
	panel := m.styles.Panel(m.modal.title, content, min(110, max(64, m.width-6)), max(14, m.height-8), m.cfg.Theme.LogsBorder, m.cfg.Theme.LogsTitle)
	footer := m.renderShortcutHints(limitShortcutHints([]shortcutHint{
		{Key: "Esc/Enter/q", Text: "Close"},
		{Key: "↑↓", Text: "Scroll"},
		{Key: "PgUp/PgDn", Text: "Fast scroll"},
	})...)
	return lipgloss.JoinVertical(lipgloss.Left, panel, lipgloss.NewStyle().Width(lipgloss.Width(panel)).Align(lipgloss.Right).Render(footer))
}

func (m *Model) handleLogsModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		shouldRefresh := m.modal.refreshRepo
		m.closeModal()
		if shouldRefresh {
			cmds = append(cmds, loadRepoCmd(m.repo.Root))
		}
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
