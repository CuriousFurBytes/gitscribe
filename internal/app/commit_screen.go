package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) renderCommitScreen() string {
	label := "Commit"
	if m.amendMode {
		label = "Amend Commit"
	}
	return m.renderForm(label, m.commitForm, screenCommit, m.cfg.AI.Enabled)
}

func (m *Model) handleCommit(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	return m.handleForm(msg, cmds, screenCommit, &m.commitForm, m.cfg.AI.Enabled)
}
