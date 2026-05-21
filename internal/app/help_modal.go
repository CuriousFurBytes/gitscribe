package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openHelpModal() {
	m.modal.visible = true
	m.modal.kind = modalHelp
	m.modal.title = "Help"
	m.modal.loading = false
	m.modal.success = false
	m.modal.selecting = false
}

func (m *Model) renderHelpModal() string {
	type item struct {
		key  string
		desc string
	}
	sections := []struct {
		title string
		items []item
	}{
		{
			title: "Workspace",
			items: []item{
				{key: "↑/↓  j/k", desc: "Navigate file tree"},
				{key: "Space", desc: "Stage or unstage file or directory"},
				{key: "Ctrl+A", desc: "Stage all / unstage all"},
				{key: "1 / 2 / 3", desc: "Switch diff tab: Diff / Raw / Preview"},
				{key: "e", desc: "Open file in $EDITOR"},
				{key: "Ctrl+O", desc: "Copy file path to clipboard"},
			},
		},
		{
			title: "Git operations",
			items: []item{
				{key: "f", desc: "Fetch"},
				{key: "p / P", desc: "Pull / Push"},
				{key: "s", desc: "Stash changes (opens name input)"},
				{key: "S", desc: "Open stash screen"},
				{key: "d", desc: "Discard unstaged changes (with confirm)"},
				{key: "D", desc: "Reset file to HEAD (with confirm)"},
				{key: "i", desc: "Add file to .gitignore"},
			},
		},
		{
			title: "Commit flows",
			items: []item{
				{key: "c", desc: "Open commit screen"},
				{key: "a", desc: "Open commit screen with AI pre-fill"},
				{key: "A", desc: "Amend last commit"},
				{key: "w", desc: "Commit without hooks (--no-verify)"},
				{key: "Ctrl+P", desc: "Open pull request screen"},
				{key: "H", desc: "Open commit history"},
				{key: "b / Ctrl+B", desc: "Switch branches"},
				{key: "W", desc: "Open worktrees"},
			},
		},
		{
			title: "Modals",
			items: []item{
				{key: "Ctrl+G", desc: "Run hooks (including pre-commit)"},
				{key: ":", desc: "Open shell"},
				{key: "L", desc: "View application logs"},
				{key: "?", desc: "Toggle this help modal"},
				{key: "Esc / q", desc: "Close modal or quit"},
			},
		},
		{
			title: "Commit / PR form",
			items: []item{
				{key: "Enter (title)", desc: "Submit"},
				{key: "Shift+Enter (body)", desc: "Submit"},
				{key: "Ctrl+A", desc: "Generate AI message"},
				{key: "Ctrl+R", desc: "Regenerate AI message"},
				{key: "Ctrl+E", desc: "Regenerate with feedback"},
				{key: "Ctrl+L", desc: "Clear title and body"},
				{key: "Ctrl+W", desc: "Toggle --no-verify (commit only)"},
				{key: "Tab", desc: "Switch focus between fields"},
				{key: "Esc", desc: "Cancel and return to main"},
			},
		},
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	lines := []string{}
	for _, section := range sections {
		lines = append(lines, m.styles.Accent(section.title))
		for _, entry := range section.items {
			lines = append(lines, "  "+keyStyle.Render(entry.key)+"  "+descStyle.Render(entry.desc))
		}
		lines = append(lines, "")
	}
	lines = append(lines, m.renderShortcutHints(shortcutHint{Key: "Esc/?", Text: "Close"}))

	return m.styles.Modal("Help", strings.Join(lines, "\n"), min(104, m.width-8), m.cfg.Theme.HelpBorder, m.cfg.Theme.HelpTitle)
}
