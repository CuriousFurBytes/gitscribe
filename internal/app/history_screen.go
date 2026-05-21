package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func (m *Model) renderHistory() string {
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Bold(true).Render("History")
	panelHeight := max(10, m.height-5)
	leftWidth, rightWidth := mainPanelWidths(m.width)

	listContent := m.renderHistoryList(panelHeight-2, max(20, leftWidth-4))
	leftPanel := m.styles.Panel("History", listContent, leftWidth, panelHeight, m.cfg.Theme.HistoryListBorder, m.cfg.Theme.HistoryListTitle)
	rightPanel := m.styles.Panel("Changes", m.historyViewport.View(), rightWidth, panelHeight, m.cfg.Theme.HistoryChangesBorder, m.cfg.Theme.HistoryChangesTitle)

	footer := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Right).Render(m.renderShortcutHints(limitShortcutHints([]shortcutHint{
		{Key: "H", Text: "Back"},
		{Key: "↑↓", Text: "Select"},
		{Key: "r", Text: "Refresh"},
		{Key: "Ctrl+G", Text: "Hooks"},
		{Key: ":", Text: "Shell"},
	})...))
	if statusLine := m.renderBottomStatusLine(m.width); statusLine != "" {
		footer = lipgloss.JoinVertical(lipgloss.Left, footer, statusLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel), footer)
}

func (m *Model) renderHistoryList(height, contentWidth int) string {
	if len(m.historyEntries) == 0 {
		if m.loading {
			return m.spinner.View() + " Loading history..."
		}
		return "No commits found."
	}

	total := len(m.historyEntries)
	showScrollbar := total > height
	lineWidth := contentWidth
	if showScrollbar {
		lineWidth = max(1, contentWidth-2)
	}

	start := clamp(m.historyIndex-height+1, 0, max(0, total-height))
	end := min(total, start+height)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		entry := m.historyEntries[i]
		prefix := "  "
		if i == m.historyIndex {
			prefix = m.selectedArrow() + " "
		}
		line := prefix + m.renderHistoryEntry(entry, i == m.historyIndex)
		lines = append(lines, truncateANSI(line, lineWidth))
	}

	if showScrollbar {
		scrollbar := verticalScrollbar(height, total, start, m.cfg.Theme.ShortcutTextFG, m.cfg.Theme.AccentFG)
		fullLines := make([]string, height)
		copy(fullLines, lines)
		for i := range fullLines {
			fullLines[i] = fullLines[i] + " " + scrollbar[i]
		}
		return strings.Join(fullLines, "\n")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderHistoryEntry(entry git.CommitHistoryEntry, selected bool) string {
	hash := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(selected).Render(entry.Hash)
	author := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.WarningFG)).Render(entry.Author)
	subjectStyle := lipgloss.NewStyle()
	if selected {
		subjectStyle = subjectStyle.Bold(true)
	}
	return hash + "  " + author + "  " + subjectStyle.Render(entry.Subject)
}

func (m *Model) handleHistory(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
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
			if m.historyIndex > 0 {
				m.historyIndex--
				cmds = append(cmds, loadCommitPatchCmd(m.repo.Root, m.historyEntries[m.historyIndex].Hash))
			}
		case "down", "j":
			if m.historyIndex < len(m.historyEntries)-1 {
				m.historyIndex++
				cmds = append(cmds, loadCommitPatchCmd(m.repo.Root, m.historyEntries[m.historyIndex].Hash))
			}
		case "r":
			cmds = append(cmds, loadHistoryCmd(m.repo.Root, m.cfg.History.MaxCommits))
		case "pgup", "b":
			m.historyViewport.LineUp(10)
		case "pgdown", "f":
			m.historyViewport.LineDown(10)
		}
	}
	return m, tea.Batch(cmds...)
}
