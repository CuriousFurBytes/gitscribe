package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
	"github.com/CuriousFurBytes/gitscribe/internal/repo"
)

func TestChangePathsFiltersByRequestedAction(t *testing.T) {
	changes := []git.FileChange{
		{Path: "staged-delete", StagedStatus: git.StatusDeleted},
		{Path: "unstaged-delete", UnstagedStatus: git.StatusDeleted},
		{Path: "new-file", IsUntracked: true, UnstagedStatus: git.StatusUntracked},
		{Path: "mixed-file", StagedStatus: git.StatusModified, UnstagedStatus: git.StatusModified},
	}

	stagePaths := changePaths(changes, false)
	if got, want := len(stagePaths), 3; got != want {
		t.Fatalf("stagePaths length = %d, want %d (%v)", got, want, stagePaths)
	}
	if stagePaths[0] != "unstaged-delete" || stagePaths[1] != "new-file" || stagePaths[2] != "mixed-file" {
		t.Fatalf("unexpected stage paths: %v", stagePaths)
	}

	unstagePaths := changePaths(changes, true)
	if got, want := len(unstagePaths), 2; got != want {
		t.Fatalf("unstagePaths length = %d, want %d (%v)", got, want, unstagePaths)
	}
	if unstagePaths[0] != "staged-delete" || unstagePaths[1] != "mixed-file" {
		t.Fatalf("unexpected unstage paths: %v", unstagePaths)
	}
}

func TestMainPanelWidthsUsesTwentyPercentForTree(t *testing.T) {
	leftWidth, rightWidth := mainPanelWidths(103)
	if leftWidth != 20 {
		t.Fatalf("leftWidth = %d, want 20", leftWidth)
	}
	if rightWidth != 82 {
		t.Fatalf("rightWidth = %d, want 82", rightWidth)
	}
}

func TestBuildTreeRowsNestsFilesUnderDirectories(t *testing.T) {
	rows := buildTreeRows("gitscribe", []git.FileChange{
		{Path: "internal/app/model.go", UnstagedStatus: git.StatusModified},
		{Path: "internal/app/new_file.go", IsUntracked: true, UnstagedStatus: git.StatusUntracked},
	})

	if len(rows) < 5 {
		t.Fatalf("expected nested rows, got %d", len(rows))
	}

	expected := map[string]int{
		"internal":                 1,
		"internal/app":             2,
		"internal/app/model.go":    3,
		"internal/app/new_file.go": 3,
	}
	for _, row := range rows {
		wantLevel, ok := expected[row.Path]
		if !ok {
			continue
		}
		if row.Level != wantLevel {
			t.Fatalf("row %s level = %d, want %d", row.Path, row.Level, wantLevel)
		}
	}
}

func TestRenderHelpModalMentionsHooksAndShellShortcuts(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	help := m.renderHelpModal()
	for _, want := range []string{"Ctrl+G", ":", "pre-commit", "shell"} {
		if !strings.Contains(strings.ToLower(help), strings.ToLower(want)) {
			t.Fatalf("renderHelpModal missing %q: %s", want, help)
		}
	}
}

func TestCtrlGOpensPreCommitModal(t *testing.T) {
	cfg := config.Defaults()
	cfg.Hooks = []config.HookConfig{{Name: "Lint", Command: []string{"make", "lint"}}}
	m := newReadyTestModelWithConfig(cfg)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	got := model.(*Model)

	if !got.modal.visible {
		t.Fatalf("expected modal to be visible")
	}
	if got.modal.kind != modalHooks {
		t.Fatalf("modal kind = %q, want %q", got.modal.kind, modalHooks)
	}
	if got.modal.title != "Hooks" {
		t.Fatalf("modal title = %q, want %q", got.modal.title, "Hooks")
	}
	view := got.View()
	for _, want := range []string{"Pre-commit", "Lint"} {
		if !strings.Contains(view, want) {
			t.Fatalf("hook modal missing %q", want)
		}
	}
}

func TestColonOpensShellModal(t *testing.T) {
	m := newReadyTestModel()

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	got := model.(*Model)

	if !got.modal.visible {
		t.Fatalf("expected modal to be visible")
	}
	if got.modal.kind != modalShell {
		t.Fatalf("modal kind = %q, want %q", got.modal.kind, modalShell)
	}
	if got.modal.title != "Shell" {
		t.Fatalf("modal title = %q, want %q", got.modal.title, "Shell")
	}
}

func TestPullStartsTitleSpinnerState(t *testing.T) {
	m := newReadyTestModel()
	m.titleAnimationUntil = time.Now().Add(-time.Second)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	got := model.(*Model)

	if got.operationStatus.label != "Pulling" {
		t.Fatalf("operationStatus = %q, want %q", got.operationStatus.label, "Pulling")
	}
	if strings.Contains(got.visibleAppTitle(), "Pulling") {
		t.Fatalf("expected title to stay clean, got %q", got.visibleAppTitle())
	}
	if !strings.Contains(got.View(), "Pulling") {
		t.Fatalf("expected bottom bar to include operation status")
	}
}

func TestOperationResultOpensLogsModalWithoutSwitchingScreens(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit

	model, _ := m.Update(operationResultMsg{
		title:           "Commit",
		output:          "hook failed",
		success:         false,
		failureReturnTo: screenCommit,
		err:             assertErr("commit failed"),
	})
	got := model.(*Model)

	if got.screen != screenCommit {
		t.Fatalf("screen = %q, want %q", got.screen, screenCommit)
	}
	if !got.modal.visible {
		t.Fatalf("expected logs modal to be visible")
	}
	if got.modal.kind != modalLogs {
		t.Fatalf("modal kind = %q, want %q", got.modal.kind, modalLogs)
	}
	if got.notice != "" {
		t.Fatalf("expected failure details to stay modal-only, got notice %q", got.notice)
	}
}

func TestAnimatedTitleStopsAfterFiveSeconds(t *testing.T) {
	m := newReadyTestModel()
	m.titleAnimationUntil = time.Now().Add(-time.Second)

	title := m.visibleAppTitle()
	if strings.Contains(title, spinnerDot) {
		t.Fatalf("expected expired title animation to stop, got %q", title)
	}
}

func TestAnimatedTitleShowsMotionDuringFirstFiveSeconds(t *testing.T) {
	m := newReadyTestModel()
	m.titleAnimationUntil = time.Now().Add(5 * time.Second)

	title := m.visibleAppTitle()
	if title == appTitle {
		t.Fatalf("expected animated title, got %q", title)
	}
}

func TestOperationResultSuccessAutoClosesLogs(t *testing.T) {
	m := newReadyTestModel()
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.screen = screenCommit

	model, _ := m.Update(operationResultMsg{
		title:           "Commit",
		output:          "done",
		success:         true,
		successReturnTo: screenMain,
		clearCommit:     true,
	})
	got := model.(*Model)

	if got.modal.visible {
		t.Fatalf("expected success log modal to stay closed")
	}
	if got.screen != screenMain {
		t.Fatalf("screen = %q, want %q", got.screen, screenMain)
	}
	if got.commitForm.Title.Value() != "" {
		t.Fatalf("expected commit form to be reset, got %q", got.commitForm.Title.Value())
	}
}

func TestHelpModalClosesOnQuestionMark(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	got := model.(*Model)
	if got.modal.visible {
		t.Fatalf("expected help modal to close")
	}
}

func TestShellModalEnterReturnsCommand(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	m.modal.input.SetValue("git status")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected shell modal enter to return a command")
	}
}

func TestCustomCommitKeybindingOpensCommitScreen(t *testing.T) {
	cfg := config.Defaults()
	cfg.Keybindings = map[string][]string{
		config.ActionOpenCommit: {"x"},
	}
	m := newReadyTestModelWithConfig(cfg)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	got := model.(*Model)
	if got.screen != screenCommit {
		t.Fatalf("screen = %q, want %q", got.screen, screenCommit)
	}
}

func TestFolderToggleKeysDoNotCollapseDirectories(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "internal/app/model.go", UnstagedStatus: git.StatusModified},
	})
	m.tree.Index = 1
	before := len(m.tree.Rows)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	got := model.(*Model)

	if len(got.tree.Rows) != before {
		t.Fatalf("expected directories to stay expanded, got %d rows before and %d after", before, len(got.tree.Rows))
	}
}

func TestSpaceOnDirectoryDoesNotCollapseTree(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "internal/app/model.go", UnstagedStatus: git.StatusModified},
	})
	before := len(m.tree.Rows)
	for i, row := range m.tree.Rows {
		if row.IsDir && row.Path != "." {
			m.tree.Index = i
			break
		}
	}

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	got := model.(*Model)

	if len(got.tree.Rows) != before {
		t.Fatalf("space on directory: expected %d rows, got %d", before, len(got.tree.Rows))
	}
}

func TestUnstageReloadDoesNotCollapseTree(t *testing.T) {
	staged := []git.FileChange{
		{Path: "internal/app/model.go", StagedStatus: git.StatusModified},
	}
	m := newReadyTestModelWithFiles(staged)
	before := len(m.tree.Rows)

	unstaged := []git.FileChange{
		{Path: "internal/app/model.go", UnstagedStatus: git.StatusModified},
	}
	model, _ := m.Update(repoLoadedMsg{
		status: git.RepoStatus{Files: unstaged},
	})
	got := model.(*Model)

	if len(got.tree.Rows) != before {
		t.Fatalf("after unstage reload: expected %d rows, got %d", before, len(got.tree.Rows))
	}
}

func TestQuitActionMatchesEscape(t *testing.T) {
	m := newReadyTestModel()
	if !m.matchesKeybinding(config.ActionQuit, tea.KeyMsg{Type: tea.KeyEsc}) {
		t.Fatalf("expected escape to match quit action")
	}
}

func TestCommitViewShowsCharacterCounts(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.commitForm.Title.SetValue("feat")
	m.commitForm.Body.SetValue("body")

	view := m.View()
	for _, want := range []string{"4/72", "Commit Title", "Commit Body"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q", want)
		}
	}
}

func TestEditorKeyOpensFileFromTree(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "main.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir {
			m.tree.Index = i
			break
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatalf("e key on file tree should return a command")
	}
}

func TestUpdateRepoLoadedRefreshesState(t *testing.T) {
	m := newReadyTestModel()

	model, _ := m.Update(repoLoadedMsg{
		status: git.RepoStatus{
			Branch: "feature/modal",
			Files: []git.FileChange{
				{Path: "internal/app/model.go", UnstagedStatus: git.StatusModified},
			},
		},
		branches: []string{"main", "feature/modal"},
	})
	got := model.(*Model)

	if got.repo.Branch != "feature/modal" {
		t.Fatalf("branch = %q, want %q", got.repo.Branch, "feature/modal")
	}
	if len(got.tree.Rows) == 0 {
		t.Fatalf("expected tree rows to be rebuilt")
	}
}

func TestRenderViewIncludesHelpModalContent(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()

	view := m.View()
	for _, want := range []string{"Help", "Modals", "Ctrl+G", ":"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q", want)
		}
	}
}

func TestShouldUnstageDirectoryRequiresStagedOnlyFiles(t *testing.T) {
	if !shouldUnstageDirectory([]git.FileChange{{Path: "staged", StagedStatus: git.StatusModified}}) {
		t.Fatalf("expected staged-only directory to unstage")
	}
	if shouldUnstageDirectory([]git.FileChange{{Path: "mixed", StagedStatus: git.StatusModified, UnstagedStatus: git.StatusModified}}) {
		t.Fatalf("expected mixed directory to stay in stage mode")
	}
}

func TestTrimLogKeepsTail(t *testing.T) {
	got := trimLog("one\ntwo\nthree\nfour", 2)
	if got != "three\nfour" {
		t.Fatalf("trimLog() = %q, want %q", got, "three\nfour")
	}
}

func TestStatusBadgesMarksUntrackedFiles(t *testing.T) {
	m := newReadyTestModel()
	got := m.statusBadges(git.FileChange{IsUntracked: true, UnstagedStatus: git.StatusUntracked})
	if got == "" {
		t.Fatalf("expected untracked badge")
	}
}

func testRepoInfo() repo.Info {
	return repo.Info{Root: "/tmp/repo", GitDir: "/tmp/repo/.git", Branch: "main"}
}

func newReadyTestModel() *Model {
	cfg := config.Defaults()
	return newReadyTestModelWithConfig(cfg)
}

func newReadyTestModelWithConfig(cfg config.Config) *Model {
	cfg.Logs.AutoCloseOnSuccess = false
	cfg.UI.ShowIntroAnimation = true

	m := New(cfg, repo.Info{Root: "/tmp/repo", GitDir: "/tmp/repo/.git", Branch: "main"}, Options{})
	m.ready = true
	m.width = 120
	m.height = 40
	m.resize()
	m.status.RepoStatus = git.RepoStatus{
		Branch:    "main",
		HasStaged: true,
		Files: []git.FileChange{
			{Path: "README.md", UnstagedStatus: git.StatusModified},
		},
	}
	m.rebuildTree()
	return m
}

func newReadyTestModelWithFiles(files []git.FileChange) *Model {
	cfg := config.Defaults()
	cfg.Logs.AutoCloseOnSuccess = false
	cfg.UI.ShowIntroAnimation = true

	m := New(cfg, repo.Info{Root: "/tmp/repo", GitDir: "/tmp/repo/.git", Branch: "main"}, Options{})
	m.ready = true
	m.width = 120
	m.height = 40
	m.resize()
	m.status.RepoStatus = git.RepoStatus{
		Branch:    "main",
		HasStaged: true,
		Files:     files,
	}
	m.rebuildTree()
	return m
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
