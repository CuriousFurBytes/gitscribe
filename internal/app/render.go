package app

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

var titleAnimationFrames = []string{"◜", "◠", "◝", "◞", "◡", "◟"}

func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.width < m.cfg.UI.MinWidth || m.height < m.cfg.UI.MinHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Terminal too small for GitScribe.\nResize to continue.")
	}

	base := m.renderBaseScreen()
	base = lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, base)
	if m.branchSelector {
		base = overlayCentered(base, m.renderBranchSelector(), m.width, m.height)
	}
	if m.confirm != nil {
		base = overlayCentered(base, m.renderConfirmModal(), m.width, m.height)
	}
	if m.modal.visible {
		base = overlayCentered(base, m.renderModal(), m.width, m.height)
	}
	return base
}

func (m *Model) renderBaseScreen() string {
	switch m.screen {
	case screenHistory:
		return m.renderHistory()
	case screenCommit:
		return m.renderCommitScreen()
	case screenPR:
		return m.renderPRScreen()
	case screenStash:
		return m.renderStash()
	default:
		return m.renderMain()
	}
}

func (m *Model) renderModal() string {
	switch m.modal.kind {
	case modalHelp:
		return m.renderHelpModal()
	case modalShell:
		return m.renderShellModal()
	case modalHooks:
		return m.renderHooksModal()
	case modalStash:
		return m.renderStashModal()
	case modalWorktree:
		return m.renderWorktreeModal()
	default:
		return m.renderScrollableModal()
	}
}

func (m *Model) renderTree(height int, width int) string {
	rows := m.tree.Rows
	if len(rows) == 0 {
		if m.loading {
			return m.spinner.View() + " Loading repository state..."
		}
		return "Working tree clean."
	}

	start := clamp(m.tree.Index-height+1, 0, max(0, len(rows)-height))
	if m.tree.Index < start {
		start = m.tree.Index
	}
	end := min(len(rows), start+height)

	rendered := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		row := rows[i]
		prefix := "  "
		if i == m.tree.Index {
			prefix = m.selectedArrow() + " "
		}

		label := row.Label
		if row.Change != nil {
			if row.Change.IsDeleted || row.Change.StagedStatus == "deleted" || row.Change.UnstagedStatus == "deleted" {
				label = lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.ErrorFG)).Strikethrough(true).Render(row.Label)
			}
			label += " " + m.statusBadges(*row.Change)
		}
		line := prefix + indent(row.Level) + label
		if ansi.StringWidth(line) > width {
			line = truncateANSI(line, width)
		}
		rendered = append(rendered, line)
	}

	return strings.Join(rendered, "\n")
}

func (m *Model) visibleAppTitle() string {
	if m.shouldAnimateTitle() {
		return titleAnimationFrames[m.titleAnimationFrame%len(titleAnimationFrames)] + " " + appTitle
	}
	return appTitle
}

func (m *Model) selectedArrow() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.cfg.Theme.SelectedArrowFG)).Bold(true).Render("›")
}

func (m *Model) statusBadges(change git.FileChange) string {
	badges := []string{}
	if change.IsUntracked {
		badges = append(badges, m.styles.Dot(m.cfg.Theme.UntrackedFG))
	}
	if change.HasUnstaged() && !change.IsUntracked && !change.IsChangedAfterStage {
		badges = append(badges, m.styles.Dot(m.cfg.Theme.UnstagedFG))
	}
	if change.HasStaged() {
		badges = append(badges, m.styles.Dot(m.cfg.Theme.StagedFG))
	}
	if change.IsChangedAfterStage || change.StagedStatus == git.StatusConflicted || change.UnstagedStatus == git.StatusConflicted {
		badges = append(badges, m.styles.Dot(m.cfg.Theme.ChangedFG))
	}
	if len(badges) == 0 && change.HasUnstaged() {
		badges = append(badges, m.styles.Dot(m.cfg.Theme.UnstagedFG))
	}
	return strings.Join(badges, " ")
}

func mainPanelWidths(totalWidth int) (int, int) {
	leftWidth := mainPanelLeftWidth(totalWidth)
	return leftWidth, max(40, totalWidth-leftWidth-1)
}

func mainPanelLeftWidth(totalWidth int) int {
	return max(20, (totalWidth-3)*20/100)
}

func displayName(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	return filepath.Base(path)
}
