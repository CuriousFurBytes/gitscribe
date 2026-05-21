package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/logger"
)

func (m *Model) renderForm(title string, form formState, current screen, allowAI bool) string {
	formWidth := max(56, int(float64(m.width)*0.54))

	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	titleCount := utf8.RuneCountInString(form.Title.Value())
	titleLabel := title + " Title  " + muted.Render(fmt.Sprintf("%d/%d", titleCount, form.Title.CharLimit))
	if form.NoVerify {
		titleLabel += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true).Render("--no-verify")
	}
	if current == screenCommit {
		if summary := m.status.RepoStatus.StagedStats.Summary(); summary != "" {
			titleLabel += "  " + muted.Render(summary)
		}
	}
	titleSection := lipgloss.JoinVertical(lipgloss.Left, titleLabel, "", form.Title.View())

	bodyCount := utf8.RuneCountInString(form.Body.Value())
	bodyLabel := title + " Body  " + muted.Render(fmt.Sprintf("%d", bodyCount))
	bodySection := lipgloss.JoinVertical(lipgloss.Left, bodyLabel, "", form.Body.View())

	shortcuts := []shortcutHint{
		{Key: "Tab", Text: "Focus"},
		{Key: "Esc", Text: "Cancel"},
		{Key: "Ctrl+?", Text: "Help"},
	}
	if current == screenCommit {
		shortcuts = append([]shortcutHint{{Key: "Ctrl+W", Text: "No-verify"}}, shortcuts...)
	}
	if allowAI {
		shortcuts = append([]shortcutHint{{Key: "Ctrl+A", Text: "AI"}}, shortcuts...)
	}
	if form.Focus == focusTitle {
		shortcuts = append([]shortcutHint{{Key: "Enter", Text: "Submit"}}, shortcuts...)
	} else {
		shortcuts = append([]shortcutHint{{Key: "Shift+Enter", Text: "Submit"}}, shortcuts...)
	}

	shortcutsStr := m.renderShortcutHints(limitShortcutHints(shortcuts)...)
	spinnerStr := ""
	if form.Loading {
		spinnerStr = m.spinner.View() + " Generating..."
	}
	spinnerWidth := lipgloss.Width(spinnerStr)
	footerRight := lipgloss.NewStyle().Width(max(0, formWidth-spinnerWidth)).Align(lipgloss.Right).Render(shortcutsStr)
	footer := spinnerStr + footerRight
	lines := []string{titleSection, "", bodySection}
	if current == screenPR {
		if section := m.renderBranchCommitsSection(muted); section != "" {
			lines = append(lines, "", section)
		}
	}
	if form.FeedbackInputVisible {
		feedbackLabel := muted.Render("Feedback for AI (Enter to confirm, Esc to cancel):")
		lines = append(lines, "", feedbackLabel, form.FeedbackInput.View())
	}
	lines = append(lines, footer)
	if statusLine := m.renderBottomStatusLine(formWidth); statusLine != "" {
		lines = append(lines, statusLine)
	}
	if form.Error != "" {
		lines = append(lines, m.styles.Error(form.Error))
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// renderBranchCommitsSection renders the list of commits that would be
// included in a pull request (commits on the current branch not yet on
// the base branch). It returns an empty string when there are no
// commits to list.
func (m *Model) renderBranchCommitsSection(muted lipgloss.Style) string {
	commits := m.status.RepoStatus.BranchCommits
	if len(commits) == 0 {
		return ""
	}
	header := muted.Render(fmt.Sprintf("Commits in this PR (%d)", len(commits)))
	rows := make([]string, 0, len(commits)+1)
	rows = append(rows, header)
	shaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	for _, c := range commits {
		rows = append(rows, fmt.Sprintf("%s  %s", shaStyle.Render(c.ShortSHA), c.Subject))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) handleForm(msg tea.Msg, cmds []tea.Cmd, current screen, form *formState, allowAI bool) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Feedback input mode: intercept all keys
		if form.FeedbackInputVisible {
			switch msg.Type {
			case tea.KeyEnter:
				form.UserFeedback = form.FeedbackInput.Value()
				form.FeedbackInputVisible = false
				form.FeedbackInput.SetValue("")
				form.FeedbackInput.Blur()
				form.Loading = true
				form.Error = ""
				cmds = append(cmds, aiCmd(m.repo.Root, m.status.RepoStatus, m.cfg.AI, current, form.UserFeedback))
				return m, tea.Batch(cmds...)
			case tea.KeyEsc:
				form.FeedbackInputVisible = false
				form.FeedbackInput.SetValue("")
				form.FeedbackInput.Blur()
				return m, tea.Batch(cmds...)
			default:
				form.FeedbackInput, _ = form.FeedbackInput.Update(msg)
				return m, tea.Batch(cmds...)
			}
		}

		// Enter in title field → submit
		if msg.Type == tea.KeyEnter && form.Focus == focusTitle {
			return m.submitForm(cmds, current, form)
		}

		// Shift+Enter in body field → submit (kitty protocol and compatible terminals)
		if msg.String() == "shift+enter" && form.Focus == focusBody {
			return m.submitForm(cmds, current, form)
		}

		switch {
		case msg.Type == tea.KeyCtrlA:
			if allowAI {
				logger.Info("AI generate", "screen", current)
				form.Loading = true
				form.Error = ""
				cmds = append(cmds, aiCmd(m.repo.Root, m.status.RepoStatus, m.cfg.AI, current, form.UserFeedback))
				return m, tea.Batch(cmds...)
			}

		case msg.Type == tea.KeyCtrlR:
			if allowAI {
				logger.Info("AI regenerate", "screen", current)
				form.UserFeedback = ""
				form.Loading = true
				form.Error = ""
				cmds = append(cmds, aiCmd(m.repo.Root, m.status.RepoStatus, m.cfg.AI, current, ""))
				return m, tea.Batch(cmds...)
			}

		case msg.Type == tea.KeyCtrlE:
			if allowAI {
				form.FeedbackInputVisible = true
				form.FeedbackInput.SetValue("")
				form.FeedbackInput.Focus()
				return m, tea.Batch(cmds...)
			}

		case msg.Type == tea.KeyCtrlL:
			form.Title.SetValue("")
			form.Body.SetValue("")
			form.Focus = focusTitle
			form.applyFocus()
			m.persistDraft()
			return m, tea.Batch(cmds...)

		case msg.Type == tea.KeyCtrlW:
			if current == screenCommit {
				form.NoVerify = !form.NoVerify
				logger.Debug("toggle no-verify", "noVerify", form.NoVerify)
				return m, tea.Batch(cmds...)
			}

		case m.matchesKeybinding(config.ActionSubmit, msg):
			return m.submitForm(cmds, current, form)

		case m.matchesKeybinding(config.ActionCancel, msg):
			logger.Info("cancel form", "screen", current)
			form.Error = ""
			form.NoVerify = false
			if m.directMode {
				return m, tea.Quit
			}
			m.screen = screenMain
			m.amendMode = false
			return m, tea.Batch(cmds...)

		case m.matchesKeybinding(config.ActionNextField, msg):
			form.focusNext()
			return m, tea.Batch(cmds...)

		case m.matchesKeybinding(config.ActionPrevField, msg):
			form.focusPrev()
			return m, tea.Batch(cmds...)

		case m.matchesKeybinding(config.ActionFocusTitle, msg):
			form.Focus = focusTitle
			form.applyFocus()
			return m, tea.Batch(cmds...)

		case m.matchesKeybinding(config.ActionFocusBody, msg):
			if form.Focus == focusTitle {
				form.Focus = focusBody
				form.applyFocus()
				return m, tea.Batch(cmds...)
			}
		}
	}

	switch form.Focus {
	case focusTitle:
		form.Title, _ = form.Title.Update(msg)
	case focusBody:
		form.Body, _ = form.Body.Update(msg)
	}
	m.persistDraft()
	return m, tea.Batch(cmds...)
}

func (m *Model) submitForm(cmds []tea.Cmd, current screen, form *formState) (tea.Model, tea.Cmd) {
	if err := validateFormTitle(form.Title.Value()); err != nil {
		form.Error = err.Error()
		return m, tea.Batch(cmds...)
	}

	if current == screenCommit {
		if m.amendMode {
			logger.Info("amend commit", "title", strings.TrimSpace(form.Title.Value()))
			m.beginOperation("Amending")
			cmds = append(cmds, amendCmd(m.repo.Root, form.Title.Value(), form.Body.Value(), form.NoVerify))
			return m, tea.Batch(cmds...)
		}
		if m.cfg.Commit.RequireConfirmation {
			m.confirm = &confirmState{
				Action:  confirmCommit,
				Title:   "Confirm commit",
				Message: fmt.Sprintf("Commit %q?", strings.TrimSpace(form.Title.Value())),
			}
			return m, tea.Batch(cmds...)
		}
		logger.Info("commit", "title", strings.TrimSpace(form.Title.Value()), "noVerify", form.NoVerify)
		m.beginOperation("Committing")
		cmds = append(cmds, commitCmd(m.repo.Root, form.Title.Value(), form.Body.Value(), form.NoVerify))
		return m, tea.Batch(cmds...)
	}

	// PR screen
	m.beginOperation("Creating PR")
	m.confirm = &confirmState{
		Action:  confirmPR,
		Title:   "Create pull request",
		Message: fmt.Sprintf("Create PR %q?", strings.TrimSpace(form.Title.Value())),
	}
	return m, tea.Batch(cmds...)
}
