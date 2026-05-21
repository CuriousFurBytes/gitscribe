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

	filtered := filterBranches(m.branches, m.branchFilter)
	lines := []string{"Select a branch:\n"}
	cursor := "_"
	lines = append(lines, "filter: "+m.branchFilter+cursor+"\n")
	if len(filtered) == 0 {
		lines = append(lines, "  (no branches match)")
	}
	for i, branch := range filtered {
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
		shortcutHint{Key: "Ctrl+N", Text: "New"},
		shortcutHint{Key: "Esc", Text: "Clear/Cancel"},
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

	filtered := filterBranches(m.branches, m.branchFilter)

	switch msg.Type {
	case tea.KeyEsc:
		if m.branchFilter != "" {
			m.branchFilter = ""
			m.setCurrentBranchIndex()
		} else {
			m.branchSelector = false
		}
		return m, tea.Batch(cmds...)
	case tea.KeyBackspace:
		if m.branchFilter != "" {
			runes := []rune(m.branchFilter)
			m.branchFilter = string(runes[:len(runes)-1])
			if m.branchFilter == "" {
				m.setCurrentBranchIndex()
			} else {
				m.clampBranchIndexToFiltered()
			}
		}
		return m, tea.Batch(cmds...)
	case tea.KeyCtrlN:
		m.branchCreating = true
		m.branchCreateInput.SetValue("")
		m.branchCreateInput.Focus()
		return m, tea.Batch(cmds...)
	case tea.KeyUp:
		if m.branchIndex > 0 {
			m.branchIndex--
		}
		return m, tea.Batch(cmds...)
	case tea.KeyDown:
		if m.branchIndex < len(filtered)-1 {
			m.branchIndex++
		}
		return m, tea.Batch(cmds...)
	case tea.KeyEnter:
		if m.branchIndex >= 0 && m.branchIndex < len(filtered) {
			m.branchSelector = false
			m.branchFilter = ""
			cmds = append(cmds, switchBranchCmd(m.repo.Root, filtered[m.branchIndex]))
		}
		return m, tea.Batch(cmds...)
	case tea.KeyRunes, tea.KeySpace:
		// Treat printable runes as filter input
		runes := msg.Runes
		if msg.Type == tea.KeySpace {
			runes = []rune{' '}
		}
		if len(runes) > 0 {
			m.branchFilter += string(runes)
			m.clampBranchIndexToFiltered()
		}
		return m, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

// clampBranchIndexToFiltered ensures the branchIndex stays within the bounds of
// the filtered branch list, falling back to 0 when the filtered list is empty.
func (m *Model) clampBranchIndexToFiltered() {
	filtered := filterBranches(m.branches, m.branchFilter)
	if len(filtered) == 0 {
		m.branchIndex = 0
		return
	}
	if m.branchIndex >= len(filtered) {
		m.branchIndex = len(filtered) - 1
	}
	if m.branchIndex < 0 {
		m.branchIndex = 0
	}
}
