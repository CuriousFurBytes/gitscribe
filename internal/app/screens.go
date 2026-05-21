package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func (m *Model) handleGlobalKey(msg tea.KeyMsg, cmds []tea.Cmd) (bool, tea.Cmd) {
	switch {
	case m.matchesKeybinding(config.ActionOpenHooks, msg):
		m.openHooksModal()
		return true, tea.Batch(cmds...)
	case m.matchesKeybinding(config.ActionOpenShell, msg):
		m.openShellModal()
		return true, tea.Batch(cmds...)
	case m.matchesKeybinding(config.ActionOpenHelp, msg):
		m.openHelpModal()
		return true, tea.Batch(cmds...)
	}

	return false, nil
}

func (m *Model) handleModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.modal.kind {
	case modalHelp:
		switch {
		case m.matchesKeybinding(config.ActionOpenHelp, msg), m.matchesKeybinding(config.ActionClose, msg):
			m.closeModal()
		}
		return m, tea.Batch(cmds...)
	case modalShell:
		return m.handleShellModalKey(msg, cmds)
	case modalHooks:
		next, batched, handled := m.handleHooksModalKey(msg, cmds)
		if handled {
			return next, batched
		}
		return m.handleLogsModalKey(msg, cmds)
	case modalLogs:
		return m.handleLogsModalKey(msg, cmds)
	case modalStash:
		return m.handleStashModalKey(msg, cmds)
	case modalWorktree:
		return m.handleWorktreeModalKey(msg, cmds)
	case modalPRURL:
		return m.handlePRURLModalKey(msg, cmds)
	case modalBranchSeriesDiff:
		return m.handleBranchSeriesDiffModalKey(msg, cmds)
	}
	return m, tea.Batch(cmds...)
}
