package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// filterBranches returns the subset of branches whose names contain query as a
// case-insensitive substring. A blank query returns the input slice unchanged.
// A nil input slice yields a nil result.
func filterBranches(branches []string, query string) []string {
	q := strings.TrimSpace(query)
	if q == "" {
		return branches
	}
	if branches == nil {
		return nil
	}
	needle := strings.ToLower(q)
	out := make([]string, 0, len(branches))
	for _, b := range branches {
		if strings.Contains(strings.ToLower(b), needle) {
			out = append(out, b)
		}
	}
	return out
}

func (m *Model) renderBranchSelector() string {
	if m.branchCreating {
		baseName := ""
		if m.branchIndex >= 0 && m.branchIndex < len(m.branches) {
			baseName = m.branches[m.branchIndex]
		}
		body := strings.Join([]string{
			"New branch name (based on: " + baseName + "):",
			"",
			m.branchCreateInput.View(),
			"",
			m.renderShortcutHints(
				shortcutHint{Key: "Enter", Text: "Create"},
				shortcutHint{Key: "Esc", Text: "Back"},
			),
		}, "\n")
		return m.styles.Modal("New Branch", body, 52, m.cfg.Theme.HelpBorder, m.cfg.Theme.HelpTitle)
	}

	lines := []string{"Select a branch:\n"}
	for i, branch := range m.branches {
		prefix := "  "
		if i == m.branchIndex {
			prefix = m.selectedArrow() + " "
		}
		suffix := ""
		if branch == m.repo.Branch {
			suffix = " •"
		}
		lines = append(lines, prefix+branch+suffix)
	}
	lines = append(lines, "\n"+m.renderShortcutHints(
		shortcutHint{Key: "Enter", Text: "Switch"},
		shortcutHint{Key: "n", Text: "New"},
		shortcutHint{Key: "Esc", Text: "Cancel"},
	))
	return m.styles.Modal("Branches", strings.Join(lines, "\n"), 48, m.cfg.Theme.HelpBorder, m.cfg.Theme.HelpTitle)
}

func (m *Model) handleBranchSelector(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if m.branchCreating {
		switch msg.Type {
		case tea.KeyEnter:
			name := strings.TrimSpace(m.branchCreateInput.Value())
			if name == "" {
				return m, tea.Batch(cmds...)
			}
			base := ""
			if m.branchIndex >= 0 && m.branchIndex < len(m.branches) {
				base = m.branches[m.branchIndex]
			}
			m.branchCreating = false
			m.branchSelector = false
			m.branchCreateInput.SetValue("")
			m.branchCreateInput.Blur()
			cmds = append(cmds, createBranchCmd(m.repo.Root, name, base))
		case tea.KeyEsc:
			m.branchCreating = false
			m.branchCreateInput.SetValue("")
			m.branchCreateInput.Blur()
		default:
			m.branchCreateInput, _ = m.branchCreateInput.Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	switch msg.String() {
	case "esc":
		m.branchSelector = false
	case "up", "k":
		if m.branchIndex > 0 {
			m.branchIndex--
		}
	case "down", "j":
		if m.branchIndex < len(m.branches)-1 {
			m.branchIndex++
		}
	case "n":
		m.branchCreating = true
		m.branchCreateInput.SetValue("")
		m.branchCreateInput.Focus()
	case "enter":
		if m.branchIndex >= 0 && m.branchIndex < len(m.branches) {
			m.branchSelector = false
			cmds = append(cmds, switchBranchCmd(m.repo.Root, m.branches[m.branchIndex]))
		}
	}
	return m, tea.Batch(cmds...)
}
