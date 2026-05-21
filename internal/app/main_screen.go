package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/logger"
)

func (m *Model) renderMain() string {
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Bold(true).Render(m.visibleAppTitle())
	panelHeight := max(10, m.height-5)
	leftWidth, rightWidth := mainPanelWidths(m.width)

	treeContent := m.renderTree(panelHeight-2, leftWidth-4)

	diffHeader := m.renderDiffTabBar()
	if m.diffModeLabel != "" && m.diffTab == tabDiff {
		diffHeader += " (" + m.diffModeLabel + ")"
	}
	if pct := m.diffViewport.ScrollPercent(); pct > 0 {
		muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))
		diffHeader += "  " + muted.Render(fmt.Sprintf("↕ %d%%", int(pct*100)))
	}

	leftPanel := m.styles.Panel("Files", treeContent, leftWidth, panelHeight, m.cfg.Theme.MainFilesBorder, m.cfg.Theme.MainFilesTitle)
	rightPanel := m.styles.Panel(diffHeader, m.diffViewport.View(), rightWidth, panelHeight, m.cfg.Theme.MainDiffBorder, m.cfg.Theme.MainDiffTitle)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	shortcuts := []shortcutHint{
		{Key: "Space", Text: "Stage"},
		{Key: "c", Text: "Commit"},
		{Key: "a", Text: "AI Commit"},
		{Key: "s", Text: "Stash"},
		{Key: "?", Text: "Help"},
	}

	branchPill := m.renderBranchPill()
	rightFooter := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(max(0, m.width-lipgloss.Width(branchPill)-2)).
		Render(m.renderShortcutHints(limitShortcutHints(shortcuts)...))
	footer := lipgloss.JoinHorizontal(lipgloss.Top, branchPill, " ", rightFooter)
	if statusLine := m.renderBottomStatusLine(m.width); statusLine != "" {
		footer = lipgloss.JoinVertical(lipgloss.Left, footer, statusLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *Model) diffTabsVisible() (showAll bool) {
	change, ok := m.selectedFileChange()
	if !ok {
		return false
	}
	if change.IsDeleted || change.StagedStatus == "deleted" || change.UnstagedStatus == "deleted" {
		return false
	}
	return true
}

func (m *Model) renderDiffTabBar() string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.AccentFG)).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ShortcutTextFG))

	diffLabel := fmt.Sprintf("[1] Diff")
	if m.diffTab == tabDiff {
		diffLabel = accent.Render(diffLabel)
	} else {
		diffLabel = muted.Render(diffLabel)
	}

	if !m.diffTabsVisible() {
		return diffLabel
	}

	rawLabel := fmt.Sprintf("[2] Raw")
	if m.diffTab == tabRaw {
		rawLabel = accent.Render(rawLabel)
	} else {
		rawLabel = muted.Render(rawLabel)
	}

	previewLabel := fmt.Sprintf("[3] Preview")
	if m.diffTab == tabPreview {
		previewLabel = accent.Render(previewLabel)
	} else {
		previewLabel = muted.Render(previewLabel)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, diffLabel, "  ", rawLabel, "  ", previewLabel)
}

func (m *Model) handleMain(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.notice = ""
		switch {
		case m.matchesKeybinding(config.ActionQuit, msg):
			return m, tea.Quit
		}
		switch msg.String() {
		case "up", "k":
			if m.tree.Index > 0 {
				m.tree.Index--
				cmds = append(cmds, m.loadSelectionDiffCmd())
				m.rawContent = ""
				m.previewContent = ""
				cmds = m.maybeLoadTabContent(cmds)
			}
		case "down", "j":
			if m.tree.Index < len(m.tree.Rows)-1 {
				m.tree.Index++
				cmds = append(cmds, m.loadSelectionDiffCmd())
				m.rawContent = ""
				m.previewContent = ""
				cmds = m.maybeLoadTabContent(cmds)
			}
		case "enter":
			cmds = append(cmds, m.loadSelectionDiffCmd())
		case " ":
			row, ok := m.selectedTreeRow()
			if !ok {
				m.notice = "Select a file to stage or unstage."
				return m, tea.Batch(cmds...)
			}
			if row.IsDir {
				changes := m.changesUnderPath(row.Path)
				if len(changes) == 0 {
					m.notice = "Select a directory with tracked changes."
					return m, tea.Batch(cmds...)
				}
				cmds = append(cmds, toggleStagePathsCmd(m.repo.Root, changes, shouldUnstageDirectory(changes)))
				return m, tea.Batch(cmds...)
			}
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to stage or unstage."
				return m, tea.Batch(cmds...)
			}
			cmds = append(cmds, toggleStageCmd(m.repo.Root, *change))
		}

		switch {
		case m.matchesKeybinding(config.ActionTabDiff, msg):
			m.diffTab = tabDiff
			m.diffViewport.SetContent(m.diffContent)
			logger.Debug("switched to diff tab")

		case m.matchesKeybinding(config.ActionTabRaw, msg):
			m.diffTab = tabRaw
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Raw view requires a single file selection."
				m.diffTab = tabDiff
				return m, tea.Batch(cmds...)
			}
			if m.rawContent != "" {
				m.diffViewport.SetContent(m.rawContent)
			} else {
				cmds = append(cmds, loadRawFileCmd(m.repo.Root, change.Path))
			}
			logger.Debug("switched to raw tab", "path", change.Path)

		case m.matchesKeybinding(config.ActionTabPreview, msg):
			m.diffTab = tabPreview
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Preview requires a single file selection."
				m.diffTab = tabDiff
				return m, tea.Batch(cmds...)
			}
			if m.previewContent != "" {
				m.diffViewport.SetContent(m.previewContent)
			} else {
				cmds = append(cmds, loadPreviewCmd(m.repo.Root, change.Path, m.cfg.Git.PreviewSyntaxTheme))
			}
			logger.Debug("switched to preview tab", "path", change.Path)

		case m.matchesKeybinding(config.ActionOpenEditor, msg):
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to open in editor."
				return m, tea.Batch(cmds...)
			}
			if m.diffTab == tabRaw && m.cfg.Git.EditorCommand != "" {
				cmds = append(cmds, openFileInEditorCmd(m.repo.Root, change.Path, m.cfg.Git.EditorCommand))
			} else {
				cmds = append(cmds, openFileInEditorCmd(m.repo.Root, change.Path, ""))
			}
			return m, tea.Batch(cmds...)

		case m.matchesKeybinding(config.ActionRefresh, msg):
			m.loading = true
			cmds = append(cmds, loadRepoCmd(m.repo.Root))
			logger.Debug("refresh")

		case m.matchesKeybinding(config.ActionOpenHistory, msg):
			m.screen = screenHistory
			logger.Info("open history")
			if len(m.historyEntries) == 0 {
				cmds = append(cmds, loadHistoryCmd(m.repo.Root, m.cfg.History.MaxCommits))
			} else if len(m.historyEntries) > 0 {
				cmds = append(cmds, loadCommitPatchCmd(m.repo.Root, m.historyEntries[m.historyIndex].Hash))
			}

		case m.matchesKeybinding(config.ActionOpenCommit, msg):
			if !m.status.RepoStatus.HasStaged {
				m.notice = "Stage changes before opening the commit screen."
				return m, tea.Batch(cmds...)
			}
			logger.Info("open commit")
			m.screen = screenCommit
			m.amendMode = false
			m.commitForm.Focus = focusTitle
			m.commitForm.Title.Focus()
			m.commitForm.Body.Blur()

		case m.matchesKeybinding(config.ActionCommitAI, msg):
			if !m.status.RepoStatus.HasStaged {
				m.notice = "Stage changes before generating a commit message."
				return m, tea.Batch(cmds...)
			}
			if !m.cfg.AI.Enabled {
				m.notice = "AI is disabled in config."
				return m, tea.Batch(cmds...)
			}
			logger.Info("open commit with AI")
			m.screen = screenCommit
			m.amendMode = false
			m.commitForm.Focus = focusTitle
			m.commitForm.Title.Focus()
			m.commitForm.Body.Blur()
			m.commitForm.Loading = true
			m.commitForm.Error = ""
			cmds = append(cmds, aiCmd(m.repo.Root, m.status.RepoStatus, m.cfg.AI, screenCommit, ""))

		case m.matchesKeybinding(config.ActionCommitNoVerify, msg):
			if !m.status.RepoStatus.HasStaged {
				m.notice = "Stage changes before committing."
				return m, tea.Batch(cmds...)
			}
			logger.Info("open commit no-verify")
			m.screen = screenCommit
			m.amendMode = false
			m.commitForm.Focus = focusTitle
			m.commitForm.Title.Focus()
			m.commitForm.Body.Blur()
			m.commitForm.NoVerify = true

		case m.matchesKeybinding(config.ActionAmend, msg):
			logger.Info("open amend commit")
			m.screen = screenCommit
			m.amendMode = true
			m.commitForm.Focus = focusTitle
			m.commitForm.Title.Focus()
			m.commitForm.Body.Blur()

		case m.matchesKeybinding(config.ActionOpenPullRequest, msg):
			if !m.cfg.PullRequest.Enabled {
				m.notice = "Pull request creation is disabled in config."
				return m, tea.Batch(cmds...)
			}
			logger.Info("open PR")
			m.screen = screenPR
			m.prForm.Focus = focusTitle
			m.prForm.Title.Focus()
			m.prForm.Body.Blur()
			cmds = append(cmds, loadPRTemplateCmd(m.repo.Root, m.cfg.PullRequest.TemplatePath))

		case m.matchesKeybinding(config.ActionPull, msg):
			m.beginOperation("Pulling")
			logger.Info("pull")
			cmds = append(cmds, pullCmd(m.repo.Root))

		case m.matchesKeybinding(config.ActionPush, msg):
			m.beginOperation("Pushing")
			logger.Info("push")
			cmds = append(cmds, pushCmd(m.repo.Root))

		case m.matchesKeybinding(config.ActionFetch, msg):
			m.beginOperation("Fetching")
			logger.Info("fetch")
			cmds = append(cmds, fetchCmd(m.repo.Root))

		case m.matchesKeybinding(config.ActionStash, msg):
			if !m.status.RepoStatus.HasUncommitted {
				m.notice = "Nothing to stash."
				return m, tea.Batch(cmds...)
			}
			m.openStashModal()

		case m.matchesKeybinding(config.ActionDiscard, msg):
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to discard."
				return m, tea.Batch(cmds...)
			}
			m.confirm = &confirmState{
				Action:  confirmDiscard,
				Title:   "Discard changes",
				Message: fmt.Sprintf("Discard all changes to %q?", change.DisplayPath),
			}

		case m.matchesKeybinding(config.ActionReset, msg):
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to reset."
				return m, tea.Batch(cmds...)
			}
			if !change.HasStaged() {
				m.notice = "File has no staged changes to reset."
				return m, tea.Batch(cmds...)
			}
			m.confirm = &confirmState{
				Action:  confirmReset,
				Title:   "Reset file",
				Message: fmt.Sprintf("Reset staged changes for %q?", change.DisplayPath),
			}

		case m.matchesKeybinding(config.ActionStageAll, msg):
			if m.status.RepoStatus.HasStaged {
				cmds = append(cmds, unstageAllCmd(m.repo.Root))
			} else {
				cmds = append(cmds, stageAllCmd(m.repo.Root))
			}

		case m.matchesKeybinding(config.ActionCopyPath, msg):
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to copy its path."
				return m, tea.Batch(cmds...)
			}
			cmds = append(cmds, copyPathCmd(change.Path))

		case m.matchesKeybinding(config.ActionIgnore, msg):
			change, ok := m.selectedFileChange()
			if !ok {
				m.notice = "Select a file to ignore."
				return m, tea.Batch(cmds...)
			}
			m.confirm = &confirmState{
				Action:  confirmIgnore,
				Title:   "Ignore file",
				Message: fmt.Sprintf("Add %q to .gitignore?", change.DisplayPath),
			}

		case m.matchesKeybinding(config.ActionOpenBranches, msg):
			if len(m.branches) > 0 {
				m.branchSelector = true
				m.setCurrentBranchIndex()
			}

		case m.matchesKeybinding(config.ActionOpenLogs, msg):
			m.openAppLogsModal()

		case m.matchesKeybinding(config.ActionOpenWorktrees, msg):
			m.openWorktreeModal()
			if len(m.worktreeEntries) == 0 {
				cmds = append(cmds, loadWorktreesCmd(m.repo.Root))
			}

		case m.matchesKeybinding(config.ActionOpenStash, msg):
			m.screen = screenStash
			cmds = append(cmds, loadStashListCmd(m.repo.Root))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) maybeLoadTabContent(cmds []tea.Cmd) []tea.Cmd {
	change, ok := m.selectedFileChange()
	if !ok {
		return cmds
	}
	switch m.diffTab {
	case tabRaw:
		cmds = append(cmds, loadRawFileCmd(m.repo.Root, change.Path))
	case tabPreview:
		cmds = append(cmds, loadPreviewCmd(m.repo.Root, change.Path, m.cfg.Git.PreviewSyntaxTheme))
	}
	return cmds
}
