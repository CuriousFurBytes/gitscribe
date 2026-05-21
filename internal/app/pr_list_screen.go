package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	ghcli "github.com/CuriousFurBytes/gitscribe/internal/github"
	"github.com/CuriousFurBytes/gitscribe/internal/logger"
)

// renderPRList draws the open pull requests screen. It mirrors the stash and
// history screens: a centered header, a single bordered list panel, and a
// shortcut hint footer.
func (m *Model) renderPRList() string {
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Bold(true).Render("Open Pull Requests")
	panelHeight := max(10, m.height-5)
	panelWidth := max(60, m.width-2)

	listContent := m.renderPRListBody(panelHeight-2, max(40, panelWidth-4))
	panel := m.styles.Panel("Open Pull Requests", listContent, panelWidth, panelHeight, m.cfg.Theme.HistoryListBorder, m.cfg.Theme.HistoryListTitle)

	footer := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Right).Render(
		m.renderShortcutHints(limitShortcutHints([]shortcutHint{
			{Key: "Enter", Text: "Open in browser"},
			{Key: "↑↓", Text: "Select"},
			{Key: "r", Text: "Refresh"},
			{Key: "Esc/q", Text: "Back"},
		})...),
	)
	if statusLine := m.renderBottomStatusLine(m.width); statusLine != "" {
		footer = lipgloss.JoinVertical(lipgloss.Left, footer, statusLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, panel, footer)
}

func (m *Model) renderPRListBody(height, contentWidth int) string {
	if len(m.prList) == 0 {
		if m.loading {
			return m.spinner.View() + " Loading pull requests..."
		}
		return "No open PRs."
	}

	start := clamp(m.prListIndex-height+1, 0, max(0, len(m.prList)-height))
	end := min(len(m.prList), start+height)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		entry := m.prList[i]
		prefix := "  "
		if i == m.prListIndex {
			prefix = m.selectedArrow() + " "
		}
		line := prefix + m.renderPRListEntry(entry, i == m.prListIndex)
		lines = append(lines, truncateANSI(line, contentWidth))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderPRListEntry(entry ghcli.PullRequestSummary, selected bool) string {
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(selected)
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.WarningFG))
	branchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))
	titleStyle := lipgloss.NewStyle()
	if selected {
		titleStyle = titleStyle.Bold(true)
	}

	parts := []string{
		numberStyle.Render(fmt.Sprintf("#%d", entry.Number)),
		titleStyle.Render(entry.Title),
		authorStyle.Render("@" + entry.Author),
		branchStyle.Render(entry.HeadRefName + " -> " + entry.BaseRefName),
	}
	if entry.IsDraft {
		draftStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG)).Faint(true)
		parts = append(parts, draftStyle.Render("[draft]"))
	}
	return strings.Join(parts, "  ")
}

func (m *Model) handlePRList(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
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

		switch msg.String() {
		case "esc", "q":
			m.screen = screenMain
		case "up", "k":
			if m.prListIndex > 0 {
				m.prListIndex--
			}
		case "down", "j":
			if m.prListIndex < len(m.prList)-1 {
				m.prListIndex++
			}
		case "r":
			m.loading = true
			cmds = append(cmds, loadOpenPRsCmd(m.repo.Root, m.cfg.PullRequest.ListLimit))
		case "enter":
			if len(m.prList) == 0 {
				return m, tea.Batch(cmds...)
			}
			if m.prListIndex < 0 || m.prListIndex >= len(m.prList) {
				return m, tea.Batch(cmds...)
			}
			url := strings.TrimSpace(m.prList[m.prListIndex].URL)
			if url == "" {
				m.notice = "Selected PR has no URL."
				return m, tea.Batch(cmds...)
			}
			if err := openURLInBrowser(url); err != nil {
				logger.Error("open browser failed", "err", err)
				m.notice = "Could not open browser: " + err.Error()
			} else {
				m.notice = "Opened PR in browser: " + url
			}
		}
	}
	return m, tea.Batch(cmds...)
}

// updatePRListLoaded handles the prListLoadedMsg, updating state and any
// notice when the gh wrapper returns an error.
func (m *Model) updatePRListLoaded(msg prListLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		logger.Error("pr list load failed", "err", msg.err)
		m.notice = msg.err.Error()
		return m, tea.Batch(cmds...)
	}
	m.prList = msg.entries
	if m.prListIndex >= len(m.prList) {
		m.prListIndex = max(0, len(m.prList)-1)
	}
	return m, tea.Batch(cmds...)
}
