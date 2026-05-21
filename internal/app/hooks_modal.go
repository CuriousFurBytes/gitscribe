package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openHooksModal() {
	m.modal.visible = true
	m.modal.kind = modalHooks
	m.modal.title = "Hooks"
	m.modal.body = ""
	m.modal.loading = false
	m.modal.success = false
	m.modal.refreshRepo = false
	m.modal.hooks = m.availableHooks()
	m.modal.hookIndex = 0
	m.modal.selecting = true
	m.modal.viewport.SetContent("")
	m.modal.viewport.GotoTop()
}

func (m *Model) renderHooksModal() string {
	if !m.modal.selecting {
		return m.renderScrollableModal()
	}
	lines := []string{m.styles.Muted("Select a hook to run."), ""}
	for i, hook := range m.modal.hooks {
		prefix := "  "
		if i == m.modal.hookIndex {
			prefix = m.selectedArrow() + " "
		}
		lines = append(lines, prefix+hook.Name)
	}
	lines = append(lines, "", m.renderShortcutHints(limitShortcutHints([]shortcutHint{
		{Key: "Enter", Text: "Run"},
		{Key: "↑↓", Text: "Select"},
		{Key: "Esc", Text: "Close"},
	})...))
	return m.styles.Modal("Hooks", strings.Join(lines, "\n"), min(68, m.width-8), m.cfg.Theme.LogsBorder, m.cfg.Theme.LogsTitle)
}

func (m *Model) availableHooks() []hookOption {
	hooks := []hookOption{{Name: "Pre-commit", PreCommit: true}}
	for _, hook := range m.cfg.Hooks {
		hooks = append(hooks, hookOption{
			Name:    hook.Name,
			Command: append([]string(nil), hook.Command...),
		})
	}
	return hooks
}

func (m *Model) selectedHook() (hookOption, bool) {
	if m.modal.hookIndex < 0 || m.modal.hookIndex >= len(m.modal.hooks) {
		return hookOption{}, false
	}
	return m.modal.hooks[m.modal.hookIndex], true
}

func (m *Model) handleHooksModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd, bool) {
	if !m.modal.selecting {
		return m, tea.Batch(cmds...), false
	}
	switch msg.String() {
	case "esc", "q":
		m.closeModal()
	case "up", "k":
		if m.modal.hookIndex > 0 {
			m.modal.hookIndex--
		}
	case "down", "j":
		if m.modal.hookIndex < len(m.modal.hooks)-1 {
			m.modal.hookIndex++
		}
	case "enter":
		hook, ok := m.selectedHook()
		if ok {
			m.modal.selecting = false
			m.modal.loading = true
			m.modal.title = hook.Name
			m.modal.body = "Running hook..."
			m.modal.viewport.SetContent("")
			cmds = append(cmds, runHookCmd(m.repo.Root, m.repo.GitDir, hook))
		}
	}
	return m, tea.Batch(cmds...), true
}
