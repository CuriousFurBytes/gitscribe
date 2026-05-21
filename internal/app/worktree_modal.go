package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openWorktreeModal() {
	m.modal.visible = true
	m.modal.kind = modalWorktree
	m.modal.title = "Worktrees"
	m.modal.loading = false
	m.modal.success = false
	m.modal.selecting = true
	m.worktreeCreating = false
	m.worktreeCreateStep = 0
	m.worktreeCreatePath = ""
	m.modal.input.SetValue("")
	m.modal.input.Blur()
}

func (m *Model) renderWorktreeModal() string {
	if m.worktreeCreating {
		return m.renderWorktreeCreateModal()
	}

	lines := []string{"Select a worktree:\n"}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))

	for i, entry := range m.worktreeEntries {
		prefix := "  "
		if i == m.worktreeIndex {
			prefix = m.selectedArrow() + " "
		}
		label := entry.Branch
		if label == "" {
			label = entry.Path
		}
		if entry.Head != "" {
			label += muted.Render("  " + entry.Head)
		}
		if entry.IsCurrent {
			label += accent.Render(" •")
		}
		if entry.IsLocked {
			label += muted.Render(" 🔒")
		}
		lines = append(lines, prefix+label)
	}

	if len(m.worktreeEntries) == 0 {
		lines = append(lines, "  No worktrees found.")
	}

	lines = append(lines, "\n"+m.renderShortcutHints(
		shortcutHint{Key: "n", Text: "New"},
		shortcutHint{Key: "Esc", Text: "Close"},
	))
	return m.styles.Modal("Worktrees", strings.Join(lines, "\n"), 56, m.cfg.Theme.HelpBorder, m.cfg.Theme.HelpTitle)
}

func (m *Model) renderWorktreeCreateModal() string {
	var prompt string
	if m.worktreeCreateStep == 0 {
		prompt = "Enter path for new worktree:"
	} else {
		prompt = fmt.Sprintf("Enter branch name (path: %s):", m.worktreeCreatePath)
	}

	body := strings.Join([]string{
		prompt,
		"",
		m.modal.input.View(),
		"",
		m.renderShortcutHints(
			shortcutHint{Key: "Enter", Text: "Next"},
			shortcutHint{Key: "Esc", Text: "Back"},
		),
	}, "\n")
	return m.styles.Modal("New Worktree", body, 56, m.cfg.Theme.ConfirmBorder, m.cfg.Theme.ConfirmTitle)
}

func (m *Model) handleWorktreeModalKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if m.worktreeCreating {
		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.modal.input.Value())
			if val == "" {
				return m, tea.Batch(cmds...)
			}
			if m.worktreeCreateStep == 0 {
				m.worktreeCreatePath = val
				m.worktreeCreateStep = 1
				m.modal.input.SetValue("")
				m.modal.input.Placeholder = "Branch name"
			} else {
				branch := val
				path := m.worktreeCreatePath
				m.worktreeCreating = false
				m.worktreeCreateStep = 0
				m.closeModal()
				cmds = append(cmds, createWorktreeCmd(m.repo.Root, path, branch))
			}
		case tea.KeyEsc:
			if m.worktreeCreateStep > 0 {
				m.worktreeCreateStep--
				m.modal.input.SetValue("")
			} else {
				m.worktreeCreating = false
				m.modal.input.SetValue("")
				m.modal.input.Blur()
			}
		default:
			m.modal.input, _ = m.modal.input.Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	switch msg.String() {
	case "esc", "q":
		m.closeModal()
	case "up", "k":
		if m.worktreeIndex > 0 {
			m.worktreeIndex--
		}
	case "down", "j":
		if m.worktreeIndex < len(m.worktreeEntries)-1 {
			m.worktreeIndex++
		}
	case "n":
		m.worktreeCreating = true
		m.worktreeCreateStep = 0
		m.worktreeCreatePath = ""
		m.modal.input.SetValue("")
		m.modal.input.Placeholder = "Worktree path (e.g. ../my-feature)"
		m.modal.input.Focus()
	}
	return m, tea.Batch(cmds...)
}
