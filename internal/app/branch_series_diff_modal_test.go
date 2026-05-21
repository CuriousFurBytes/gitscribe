package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPRScreenCtrlDOpensBranchSeriesDiffModal(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	got := model.(*Model)

	if !got.modal.visible {
		t.Fatalf("expected modal visible after Ctrl+D on PR screen")
	}
	if got.modal.kind != modalBranchSeriesDiff {
		t.Fatalf("modal.kind = %q, want %q", got.modal.kind, modalBranchSeriesDiff)
	}
	if !got.modal.loading {
		t.Fatalf("expected modal.loading = true while diff is being fetched")
	}
}

func TestCtrlDOnCommitScreenDoesNotOpenModal(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	got := model.(*Model)
	if got.modal.visible && got.modal.kind == modalBranchSeriesDiff {
		t.Fatalf("did not expect branch series diff modal on commit screen")
	}
}

func TestBranchSeriesDiffLoadedMsgPopulatesViewport(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.openBranchSeriesDiffModal()

	diff := strings.Join([]string{
		"diff --git a/foo.go b/foo.go",
		"index 111..222 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1 +1 @@",
		"-old",
		"+new",
	}, "\n")

	model, _ := m.Update(branchSeriesDiffLoadedMsg{content: diff})
	got := model.(*Model)

	if got.modal.loading {
		t.Fatalf("expected loading = false after diff loaded")
	}
	if !strings.Contains(got.modal.viewport.View(), "foo.go") {
		t.Fatalf("expected viewport to contain diff content, got: %q", got.modal.viewport.View())
	}
}

func TestBranchSeriesDiffLoadedMsgEmptyShowsPlaceholder(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.openBranchSeriesDiffModal()

	model, _ := m.Update(branchSeriesDiffLoadedMsg{content: ""})
	got := model.(*Model)

	if got.modal.loading {
		t.Fatalf("expected loading = false even when diff is empty")
	}
	if !strings.Contains(got.modal.viewport.View(), "No changes") {
		t.Fatalf("expected empty-diff placeholder, got: %q", got.modal.viewport.View())
	}
}

func TestBranchSeriesDiffLoadedMsgErrorShowsMessage(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.openBranchSeriesDiffModal()

	model, _ := m.Update(branchSeriesDiffLoadedMsg{err: assertErr("boom")})
	got := model.(*Model)

	if got.modal.loading {
		t.Fatalf("expected loading = false after error")
	}
	if !strings.Contains(got.modal.viewport.View(), "boom") {
		t.Fatalf("expected error in viewport, got: %q", got.modal.viewport.View())
	}
}

func TestBranchSeriesDiffModalEscClosesModal(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.openBranchSeriesDiffModal()
	// Simulate loaded state so we are not stuck in loading.
	m.Update(branchSeriesDiffLoadedMsg{content: "diff --git a/x b/x\n"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.modal.visible {
		t.Fatalf("expected modal closed after Esc")
	}
}

func TestRenderBranchSeriesDiffModalIncludesTitleAndHints(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.openBranchSeriesDiffModal()
	m.Update(branchSeriesDiffLoadedMsg{content: "diff --git a/x b/x\n"})

	out := m.renderBranchSeriesDiffModal()
	if !strings.Contains(out, "Branch series diff") {
		t.Fatalf("expected modal title in render, got: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "esc") {
		t.Fatalf("expected close hint in render, got: %s", out)
	}
}
