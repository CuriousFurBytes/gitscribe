package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/logger"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resize()
		logger.Debug("window resize", "width", msg.Width, "height", msg.Height)
		return m, tea.Batch(cmds...)
	case rawFileLoadedMsg:
		return m.updateRawFileLoaded(msg, cmds)
	case previewLoadedMsg:
		return m.updatePreviewLoaded(msg, cmds)
	case copyPathMsg:
		if msg.err != nil {
			m.notice = "Failed to copy path: " + msg.err.Error()
		} else {
			m.notice = "Copied: " + msg.path
		}
		return m, tea.Batch(cmds...)
	case worktreeListLoadedMsg:
		if msg.err != nil {
			m.notice = msg.err.Error()
		} else {
			m.worktreeEntries = msg.entries
			m.worktreeIndex = 0
		}
		return m, tea.Batch(cmds...)
	case stashListLoadedMsg:
		if msg.err != nil {
			m.notice = msg.err.Error()
		} else {
			m.stashEntries = msg.entries
			m.stashIndex = 0
			if len(m.stashEntries) > 0 {
				cmds = append(cmds, loadStashDiffCmd(m.repo.Root, m.stashEntries[0].RefName))
			} else {
				m.stashViewport.SetContent("No stashes.")
			}
		}
		return m, tea.Batch(cmds...)
	case stashDiffLoadedMsg:
		if msg.err != nil {
			m.stashViewport.SetContent("Could not load stash diff: " + msg.err.Error())
		} else if len(m.stashEntries) > 0 && m.stashEntries[m.stashIndex].RefName == msg.ref {
			content := strings.TrimSpace(msg.content)
			if content == "" {
				content = "No patch available."
			}
			m.stashViewport.SetContent(m.colorizeDiff(content))
			m.stashViewport.GotoTop()
		}
		return m, tea.Batch(cmds...)
	case titleTickMsg:
		m.titleAnimationFrame++
		if m.shouldAnimateTitle() || m.operationStatus.label != "" {
			cmds = append(cmds, titleTickCmd())
		}
		return m, tea.Batch(cmds...)
	case repoLoadedMsg:
		return m.updateRepoLoaded(msg, cmds)
	case diffLoadedMsg:
		return m.updateDiffLoaded(msg, cmds)
	case historyLoadedMsg:
		return m.updateHistoryLoaded(msg, cmds)
	case commitPatchLoadedMsg:
		return m.updateCommitPatchLoaded(msg, cmds)
	case operationResultMsg:
		return m.updateOperationResult(msg, cmds)
	case aiGeneratedMsg:
		return m.updateAIGenerated(msg, cmds)
	case prTemplateLoadedMsg:
		return m.updatePRTemplateLoaded(msg, cmds)
	case shellCommandResultMsg:
		m.modal.loading = false
		content := trimLog(msg.output, m.cfg.Logs.MaxLines)
		if strings.TrimSpace(content) == "" {
			if msg.err == nil {
				content = "Command completed successfully."
			} else {
				content = msg.err.Error()
			}
		}
		m.modal.success = msg.err == nil
		m.modal.viewport.SetContent(m.colorizeLog(content))
		m.modal.viewport.GotoTop()
		if msg.err != nil {
			m.notice = ""
		}
		return m, tea.Batch(cmds...)
	case fileEditorFinishedMsg:
		if msg.err != nil {
			m.notice = msg.err.Error()
			return m, tea.Batch(cmds...)
		}
		cmds = append(cmds, loadRepoCmd(m.repo.Root))
		return m, tea.Batch(cmds...)
	case editorFinishedMsg:
		form := &m.commitForm
		if msg.target == screenPR {
			form = &m.prForm
		}
		if msg.err != nil {
			form.Error = msg.err.Error()
			return m, tea.Batch(cmds...)
		}
		form.Error = ""
		if msg.focus == focusTitle {
			form.Title.SetValue(msg.value)
		} else {
			form.Body.SetValue(msg.value)
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		if m.confirm != nil {
			return m.handleConfirmKey(msg, cmds)
		}
		if m.branchSelector {
			return m.handleBranchSelector(msg, cmds)
		}
		if m.modal.visible {
			return m.handleModalKey(msg, cmds)
		}
		if handled, next := m.handleGlobalKey(msg, cmds); handled {
			return m, next
		}
	}

	switch m.screen {
	case screenMain:
		return m.handleMain(msg, cmds)
	case screenHistory:
		return m.handleHistory(msg, cmds)
	case screenCommit:
		return m.handleCommit(msg, cmds)
	case screenPR:
		return m.handlePR(msg, cmds)
	case screenStash:
		return m.handleStash(msg, cmds)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateRepoLoaded(msg repoLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		logger.Error("repo load failed", "err", msg.err)
		m.notice = msg.err.Error()
		return m, tea.Batch(cmds...)
	}

	logger.Debug("repo loaded", "branch", msg.status.Branch, "files", len(msg.status.Files))
	m.status.RepoStatus = msg.status
	m.repo.Branch = msg.status.Branch
	m.branches = msg.branches
	m.setCurrentBranchIndex()
	m.rebuildTree()
	// Clear tab content cache on repo reload
	m.rawContent = ""
	m.previewContent = ""
	m.diffTab = tabDiff
	cmds = append(cmds, m.loadSelectionDiffCmd())
	return m, tea.Batch(cmds...)
}

func (m *Model) updateRawFileLoaded(msg rawFileLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		logger.Error("raw file load failed", "path", msg.path, "err", msg.err)
		m.notice = "Could not read file: " + msg.err.Error()
		m.diffTab = tabDiff
		m.diffViewport.SetContent(m.diffContent)
		return m, tea.Batch(cmds...)
	}
	logger.Debug("raw file loaded", "path", msg.path)
	m.rawContent = msg.content
	if m.diffTab == tabRaw {
		m.diffViewport.SetContent(m.rawContent)
		m.diffViewport.GotoTop()
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) updatePreviewLoaded(msg previewLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		logger.Error("preview load failed", "path", msg.path, "err", msg.err)
		m.notice = "Could not render preview: " + msg.err.Error()
		m.diffTab = tabDiff
		m.diffViewport.SetContent(m.diffContent)
		return m, tea.Batch(cmds...)
	}
	logger.Debug("preview loaded", "path", msg.path)
	m.previewContent = msg.content
	if m.diffTab == tabPreview {
		m.diffViewport.SetContent(m.previewContent)
		m.diffViewport.GotoTop()
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) updateDiffLoaded(msg diffLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if row, ok := m.selectedTreeRow(); ok && row.Path != msg.path {
		return m, tea.Batch(cmds...)
	}
	if msg.err != nil {
		m.notice = msg.err.Error()
		m.setDiff("", "")
		return m, tea.Batch(cmds...)
	}
	m.setDiff(msg.content, msg.mode)
	return m, tea.Batch(cmds...)
}

func (m *Model) updateHistoryLoaded(msg historyLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notice = msg.err.Error()
		return m, tea.Batch(cmds...)
	}
	m.historyEntries = msg.entries
	if m.historyIndex >= len(m.historyEntries) {
		m.historyIndex = max(0, len(m.historyEntries)-1)
	}
	if len(m.historyEntries) == 0 {
		m.historyViewport.SetContent("No commits found.")
		return m, tea.Batch(cmds...)
	}
	cmds = append(cmds, loadCommitPatchCmd(m.repo.Root, m.historyEntries[m.historyIndex].Hash))
	return m, tea.Batch(cmds...)
}

func (m *Model) updateCommitPatchLoaded(msg commitPatchLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if len(m.historyEntries) > 0 && m.historyEntries[m.historyIndex].Hash != msg.hash {
		return m, tea.Batch(cmds...)
	}
	if msg.err != nil {
		m.notice = msg.err.Error()
		m.historyViewport.SetContent("No patch available for this commit")
		return m, tea.Batch(cmds...)
	}
	content := strings.TrimSpace(msg.content)
	if content == "" {
		content = "No patch available for this commit"
	}
	m.historyViewport.SetContent(m.colorizeDiff(content))
	m.historyViewport.GotoTop()
	return m, tea.Batch(cmds...)
}

func (m *Model) updateOperationResult(msg operationResultMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.clearOperation()

	if msg.success {
		logger.Info("operation succeeded", "title", msg.title)
		m.notice = ""
		if msg.clearCommit {
			m.commitForm = newForm("Commit title", "Commit body", 72)
			m.amendMode = false
		}
		if msg.clearPR {
			m.prForm = newForm("PR title", "PR body", 72)
		}
		if msg.clearCommit || msg.clearPR {
			m.persistDraft()
		}
		if m.directMode && (msg.clearCommit || msg.clearPR) {
			return m, tea.Quit
		}
		// PR-create success: honor pull_request.after_create policy.
		if msg.prURL != "" {
			switch m.cfg.PullRequest.AfterCreate {
			case "modal":
				m.screen = msg.successReturnTo
				m.openPRURLModal(msg.prURL)
				if msg.refreshRepo {
					cmds = append(cmds, loadRepoCmd(m.repo.Root))
				}
				return m, tea.Batch(cmds...)
			case "browser":
				m.screen = msg.successReturnTo
				if err := openURLInBrowser(msg.prURL); err != nil {
					logger.Error("open browser failed", "err", err)
					m.notice = "Could not open browser: " + err.Error()
				} else {
					m.notice = "Opened PR in browser: " + msg.prURL
				}
				if msg.refreshRepo {
					cmds = append(cmds, loadRepoCmd(m.repo.Root))
				}
				return m, tea.Batch(cmds...)
			}
			// "none" falls through to default behavior below.
		}
		if m.cfg.Logs.AutoCloseOnSuccess && !msg.alwaysModal {
			m.closeModal()
			m.screen = msg.successReturnTo
			if msg.refreshRepo {
				cmds = append(cmds, loadRepoCmd(m.repo.Root))
			}
			return m, tea.Batch(cmds...)
		}
		m.screen = msg.successReturnTo
	} else {
		logger.Error("operation failed", "title", msg.title, "err", msg.err)
		m.notice = ""
		m.screen = msg.failureReturnTo
	}

	m.openLogsModal(msg)
	return m, tea.Batch(cmds...)
}

func (m *Model) updateAIGenerated(msg aiGeneratedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	m.activeForm().Loading = false
	if msg.err != nil {
		logger.Error("AI generation failed", "err", msg.err)
		m.activeForm().Error = msg.err.Error()
		return m, tea.Batch(cmds...)
	}
	logger.Info("AI generation succeeded", "target", msg.target)
	form := m.activeForm()
	form.Error = ""
	form.Title.SetValue(msg.resp.Title)
	if m.cfg.AI.Mode != "title_only" {
		form.Body.SetValue(msg.resp.Body)
	}
	m.persistDraft()
	return m, tea.Batch(cmds...)
}

func (m *Model) updatePRTemplateLoaded(msg prTemplateLoadedMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.prForm.Error = msg.err.Error()
		return m, tea.Batch(cmds...)
	}
	if strings.TrimSpace(m.prForm.Body.Value()) == "" {
		m.prForm.Body.SetValue(msg.body)
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) shouldAnimateTitle() bool {
	return !m.titleAnimationUntil.IsZero() && timeNow().Before(m.titleAnimationUntil)
}
