package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) renderPRScreen() string {
	return m.renderForm("Pull Request", m.prForm, screenPR, m.cfg.AI.Enabled)
}

func (m *Model) handlePR(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	return m.handleForm(msg, cmds, screenPR, &m.prForm, m.cfg.AI.Enabled)
}
