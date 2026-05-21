package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestRenderMainShowsTreeAndDiff(t *testing.T) {
	m := newReadyTestModel()
	out := m.renderMain()
	if !strings.Contains(out, "Files") || !strings.Contains(out, "Diff") {
		t.Fatalf("expected Files and Diff panels: %s", out)
	}
}

func TestMainArrowKeysMoveTreeIndex(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "a.go", UnstagedStatus: git.StatusModified},
		{Path: "b.go", UnstagedStatus: git.StatusModified},
	})
	start := m.tree.Index
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.tree.Index <= start {
		t.Fatalf("expected tree index to advance")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.tree.Index != start {
		t.Fatalf("expected tree index to retreat")
	}
}

func TestMainEnterReloadsDiff(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "x.go", UnstagedStatus: git.StatusModified}})
	m.tree.Index = 2
	if m.tree.Index >= len(m.tree.Rows) {
		m.tree.Index = len(m.tree.Rows) - 1
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd
}

func TestMainSpaceOnFileTogglesStage(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "x.go", UnstagedStatus: git.StatusModified}})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd == nil {
		t.Fatalf("expected toggle stage cmd")
	}
}

func TestMainSpaceOnDirectoryTogglesDir(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "pkg/x.go", UnstagedStatus: git.StatusModified}})
	for i, row := range m.tree.Rows {
		if row.IsDir && row.Path != "." {
			m.tree.Index = i
			break
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd == nil {
		t.Fatalf("expected toggle paths cmd")
	}
}

func TestMainCommitKeyBlockedWithoutStaged(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasStaged = false
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if model.(*Model).screen == screenCommit {
		t.Fatalf("expected commit screen blocked when nothing staged")
	}
	if !strings.Contains(model.(*Model).notice, "Stage changes") {
		t.Fatalf("expected stage-changes notice")
	}
}

func TestMainCommitKeyOpensWithStaged(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasStaged = true
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if model.(*Model).screen != screenCommit {
		t.Fatalf("expected commit screen")
	}
}

func TestMainHistoryKey(t *testing.T) {
	m := newReadyTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	if m.screen != screenHistory {
		t.Fatalf("expected history screen")
	}
	if cmd == nil {
		t.Fatalf("expected history load command")
	}
}

func TestMainPullKey(t *testing.T) {
	m := newReadyTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Fatalf("expected pull command")
	}
	if m.operationStatus.label != "Pulling" {
		t.Fatalf("expected Pulling operation")
	}
}

func TestMainRefreshKey(t *testing.T) {
	m := newReadyTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if !m.loading {
		t.Fatalf("expected refresh to set loading")
	}
	if cmd == nil {
		t.Fatalf("expected refresh command")
	}
}

func TestMainBranchesKey(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if !m.branchSelector {
		t.Fatalf("expected branch selector opened")
	}
}

func TestMainQuitKey(t *testing.T) {
	m := newReadyTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected quit cmd")
	}
}

func TestMainSpaceWithoutSelectionShowsNotice(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = nil
	m.tree.Index = -1
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !strings.Contains(model.(*Model).notice, "Select a file") {
		t.Fatalf("expected notice, got %q", model.(*Model).notice)
	}
}

func TestMainTabDiffKey(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "x.go", UnstagedStatus: git.StatusModified}})
	m.diffTab = tabRaw
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.diffTab != tabDiff {
		t.Fatalf("expected tabDiff after pressing 1, got %d", m.diffTab)
	}
}

func TestMainTabRawRequiresSingleFile(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = nil
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.diffTab != tabDiff {
		t.Fatalf("expected tabDiff when no file selected, got %d", m.diffTab)
	}
	if !strings.Contains(m.notice, "single file") {
		t.Fatalf("expected single-file notice, got %q", m.notice)
	}
}

func TestMainTabPreviewRequiresSingleFile(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = nil
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.diffTab != tabDiff {
		t.Fatalf("expected tabDiff when no file selected, got %d", m.diffTab)
	}
}

func TestMainStashKeyWithNothingToStash(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasUncommitted = false
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.(*Model).confirm != nil {
		t.Fatalf("expected no confirm when nothing to stash")
	}
}

func TestMainStashKeyWithChangesOpensStashModal(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasUncommitted = true
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	got := model.(*Model)
	if got.confirm != nil {
		t.Fatalf("expected no confirm dialog after Phase 4 (stash modal replaces it)")
	}
	if !got.modal.visible {
		t.Fatalf("expected stash modal to be visible")
	}
	if got.modal.kind != modalStash {
		t.Fatalf("modal.kind = %q, want %q", got.modal.kind, modalStash)
	}
}

func TestStashModalEscCancels(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasUncommitted = true
	m.openStashModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.modal.visible {
		t.Fatalf("Esc should close stash modal")
	}
}

func TestStashModalEnterWithNameTriggersStash(t *testing.T) {
	m := newReadyTestModel()
	m.openStashModal()
	m.modal.input.SetValue("my-stash")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter with name should trigger stash command")
	}
}

func TestStashModalEnterWithoutNameTriggersStash(t *testing.T) {
	m := newReadyTestModel()
	m.openStashModal()
	m.modal.input.SetValue("")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter without name should trigger stash command")
	}
}

func TestMainDiscardKeyOpensConfirm(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "x.go", UnstagedStatus: git.StatusModified}})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if model.(*Model).confirm == nil {
		t.Fatalf("expected confirm dialog for discard")
	}
	if model.(*Model).confirm.Action != confirmDiscard {
		t.Fatalf("expected confirmDiscard, got %q", model.(*Model).confirm.Action)
	}
}

func TestMainAmendKeyOpensCommitScreen(t *testing.T) {
	m := newReadyTestModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if model.(*Model).screen != screenCommit {
		t.Fatalf("expected commit screen for amend")
	}
	if !model.(*Model).amendMode {
		t.Fatalf("expected amendMode to be true")
	}
}

func TestMainFetchKey(t *testing.T) {
	m := newReadyTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if cmd == nil {
		t.Fatalf("expected fetch command")
	}
	if m.operationStatus.label != "Fetching" {
		t.Fatalf("expected Fetching operation, got %q", m.operationStatus.label)
	}
}

func TestMainOpenLogsKeyOpensModal(t *testing.T) {
	m := newReadyTestModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	if !model.(*Model).modal.visible {
		t.Fatalf("expected logs modal to be visible")
	}
	if model.(*Model).modal.kind != modalLogs {
		t.Fatalf("expected modalLogs, got %q", model.(*Model).modal.kind)
	}
}

func TestRenderDiffTabBarShowsTabs(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "ok.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	out := m.renderDiffTabBar()
	for _, want := range []string{"Diff", "Raw", "Preview"} {
		if !strings.Contains(out, want) {
			t.Fatalf("tab bar missing %q: %s", want, out)
		}
	}
}

func TestDiffTabBarHidesRawAndPreviewWhenNoFileSelected(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = nil
	out := m.renderDiffTabBar()
	if strings.Contains(out, "Raw") {
		t.Fatalf("Raw tab should be hidden when no file selected")
	}
	if strings.Contains(out, "Preview") {
		t.Fatalf("Preview tab should be hidden when no file selected")
	}
}

func TestDiffTabBarHidesRawAndPreviewForDeletedFile(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "gone.go", IsDeleted: true, StagedStatus: git.StatusDeleted},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	out := m.renderDiffTabBar()
	if strings.Contains(out, "Raw") {
		t.Fatalf("Raw tab should be hidden for deleted file")
	}
	if strings.Contains(out, "Preview") {
		t.Fatalf("Preview tab should be hidden for deleted file")
	}
}

func TestDiffTabBarShowsAllTabsWithRegularFileSelected(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "ok.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	out := m.renderDiffTabBar()
	for _, want := range []string{"Diff", "Raw", "Preview"} {
		if !strings.Contains(out, want) {
			t.Fatalf("tab bar missing %q for regular file: %s", want, out)
		}
	}
}

func TestDeletedFileTreeRendersWithoutPanic(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "deleted.go", IsDeleted: true, StagedStatus: git.StatusDeleted},
	})
	out := m.renderTree(20, 80)
	if !strings.Contains(out, "deleted.go") {
		t.Fatalf("expected deleted file in tree output")
	}
}

func TestMainAICommitKeyWithoutStagedShowsNotice(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasStaged = false
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if model.(*Model).screen == screenCommit {
		t.Fatalf("expected commit screen blocked when nothing staged")
	}
}

func TestMainAICommitKeyWithStagedOpensCommit(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus.HasStaged = true
	m.cfg.AI.Enabled = true
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if model.(*Model).screen != screenCommit {
		t.Fatalf("expected commit screen with AI")
	}
}

func TestMaybeLoadTabContentReturnsCmds(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "a.go", UnstagedStatus: git.StatusModified},
		{Path: "b.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	m.diffTab = tabRaw

	var cmds []tea.Cmd
	cmds = m.maybeLoadTabContent(cmds)
	if len(cmds) == 0 {
		t.Fatalf("expected maybeLoadTabContent to return non-empty cmds for Raw tab")
	}
}

func TestCursorMoveInRawTabProducesCmd(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "a.go", UnstagedStatus: git.StatusModified},
		{Path: "b.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	m.diffTab = tabRaw
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd when moving cursor in Raw tab")
	}
}
