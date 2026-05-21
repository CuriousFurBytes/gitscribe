package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func (m *Model) renderStash() string {
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Bold(true).Render("Stash")
	panelHeight := max(10, m.height-5)
	leftWidth, rightWidth := mainPanelWidths(m.width)

	listContent := m.renderStashList(panelHeight-2, max(20, leftWidth-4))
	leftPanel := m.styles.Panel("Stash", listContent, leftWidth, panelHeight, m.cfg.Theme.HistoryListBorder, m.cfg.Theme.HistoryListTitle)
	rightPanel := m.styles.Panel("Changes", m.stashViewport.View(), rightWidth, panelHeight, m.cfg.Theme.HistoryChangesBorder, m.cfg.Theme.HistoryChangesTitle)

	footer := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Right).Render(
		m.renderShortcutHints(limitShortcutHints([]shortcutHint{
			{Key: "Enter", Text: "Apply"},
			{Key: "p", Text: "Pop"},
			{Key: "d", Text: "Drop"},
			{Key: "r", Text: "Refresh"},
			{Key: "Esc/q", Text: "Back"},
		})...),
	)
	if statusLine := m.renderBottomStatusLine(m.width); statusLine != "" {
		footer = lipgloss.JoinVertical(lipgloss.Left, footer, statusLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel), footer)
}

func (m *Model) renderStashList(height, contentWidth int) string {
	if len(m.stashEntries) == 0 {
		if m.loading {
			return m.spinner.View() + " Loading stashes..."
		}
		return "No stashes found."
	}

	start := clamp(m.stashIndex-height+1, 0, max(0, len(m.stashEntries)-height))
	end := min(len(m.stashEntries), start+height)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		entry := m.stashEntries[i]
		prefix := "  "
		if i == m.stashIndex {
			prefix = m.selectedArrow() + " "
		}
		line := prefix + m.renderStashEntry(entry, i == m.stashIndex)
		lines = append(lines, truncateANSI(line, contentWidth))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderStashEntry(entry git.StashEntry, selected bool) string {
	ref := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(selected).Render(
		fmt.Sprintf("stash@{%d}", entry.Index),
	)
	subjectStyle := lipgloss.NewStyle()
	if selected {
		subjectStyle = subjectStyle.Bold(true)
	}
	return ref + "  " + subjectStyle.Render(entry.Message)
}

func (m *Model) handleStash(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.notice = ""
		if m.matchesKeybinding(config.ActionOpenBranches, msg) {
			if len(m.branches) > 0 {
				m.branchSelector = true
				m.setCurrentBranchIndex()
			}
			return m, tea.Batch(cmds...)
		}
		if m.matchesKeybinding(config.ActionOpenWorktrees, msg) {
			m.openWorktreeModal()
			if len(m.worktreeEntries) == 0 {
				cmds = append(cmds, loadWorktreesCmd(m.repo.Root))
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "esc", "q":
			m.screen = screenMain
		case "up", "k":
			if m.stashIndex > 0 {
				m.stashIndex--
				cmds = append(cmds, loadStashDiffCmd(m.repo.Root, m.stashEntries[m.stashIndex].RefName))
			}
		case "down", "j":
			if m.stashIndex < len(m.stashEntries)-1 {
				m.stashIndex++
				cmds = append(cmds, loadStashDiffCmd(m.repo.Root, m.stashEntries[m.stashIndex].RefName))
			}
		case "r":
			cmds = append(cmds, loadStashListCmd(m.repo.Root))
		case "pgup", "b":
			m.stashViewport.LineUp(10)
		case "pgdown", "f":
			m.stashViewport.LineDown(10)
		case "enter":
			if len(m.stashEntries) > 0 {
				entry := m.stashEntries[m.stashIndex]
				m.confirm = &confirmState{
					Action:  confirmStashApply,
					Title:   "Apply stash",
					Message: fmt.Sprintf("Apply %q?", entry.Message),
				}
			}
		case "p":
			if len(m.stashEntries) > 0 {
				entry := m.stashEntries[m.stashIndex]
				m.confirm = &confirmState{
					Action:  confirmStashPop,
					Title:   "Pop stash",
					Message: fmt.Sprintf("Pop %q (apply and drop)?", entry.Message),
				}
			}
		case "d":
			if len(m.stashEntries) > 0 {
				entry := m.stashEntries[m.stashIndex]
				m.confirm = &confirmState{
					Action:  confirmStashDrop,
					Title:   "Drop stash",
					Message: fmt.Sprintf("Drop %q? This cannot be undone.", entry.Message),
				}
			}
		}
	}
	return m, tea.Batch(cmds...)
}
