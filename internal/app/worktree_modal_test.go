package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestMainWKeyOpensWorktreeModal(t *testing.T) {
	m := newReadyTestModel()
	m.worktreeEntries = []git.WorktreeEntry{
		{Path: "/repo", Branch: "main", IsCurrent: true},
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}})
	got := model.(*Model)
	if !got.modal.visible {
		t.Fatalf("W key should open worktree modal")
	}
	if got.modal.kind != modalWorktree {
		t.Fatalf("modal.kind = %q, want %q", got.modal.kind, modalWorktree)
	}
}

func TestWorktreeModalEscCloses(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).modal.visible {
		t.Fatalf("Esc should close worktree modal")
	}
}

func TestWorktreeModalArrowKeysNavigate(t *testing.T) {
	m := newReadyTestModel()
	m.worktreeEntries = []git.WorktreeEntry{
		{Path: "/a", Branch: "main"},
		{Path: "/b", Branch: "feat"},
	}
	m.worktreeIndex = 0
	m.openWorktreeModal()

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).worktreeIndex != 1 {
		t.Fatalf("down should advance worktree index")
	}
	model, _ = model.(*Model).Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.(*Model).worktreeIndex != 0 {
		t.Fatalf("up should retreat worktree index")
	}
}

func TestWorktreeModalNKeyEntersCreateMode(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := model.(*Model)
	if !got.worktreeCreating {
		t.Fatalf("n should enter worktree create mode")
	}
}

func TestWorktreeModalCreateEscReturnsToList(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	m.worktreeCreating = true
	m.worktreeCreateStep = 0

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.worktreeCreating {
		t.Fatalf("Esc in create mode should exit create mode")
	}
	if !got.modal.visible {
		t.Fatalf("modal should remain visible after Esc in create mode")
	}
}

func TestRenderWorktreeModalListsEntries(t *testing.T) {
	m := newReadyTestModel()
	m.worktreeEntries = []git.WorktreeEntry{
		{Path: "/repo/main", Branch: "main", IsCurrent: true},
		{Path: "/repo/feat", Branch: "feature/x"},
	}
	m.openWorktreeModal()
	out := m.renderWorktreeModal()
	if !strings.Contains(out, "main") {
		t.Fatalf("worktree modal should list branch names, got: %s", out)
	}
}

func TestWorktreeModalCreatePathEnterAdvancesToBranchStep(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	m.worktreeCreating = true
	m.worktreeCreateStep = 0
	m.modal.input.SetValue("../my-tree")

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*Model)
	if got.worktreeCreateStep != 1 {
		t.Fatalf("Enter with path should advance to step 1 (branch input)")
	}
	if got.worktreeCreatePath != "../my-tree" {
		t.Fatalf("worktreeCreatePath = %q, want ../my-tree", got.worktreeCreatePath)
	}
}

func TestWorktreeModalCreateBranchEnterTriggersCreate(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	m.worktreeCreating = true
	m.worktreeCreateStep = 1
	m.worktreeCreatePath = "../my-tree"
	m.modal.input.SetValue("feature/new")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter in branch step should trigger createWorktreeCmd")
	}
}

func TestRenderWorktreeCreateModal(t *testing.T) {
	m := newReadyTestModel()
	m.openWorktreeModal()
	m.worktreeCreating = true
	m.worktreeCreateStep = 0
	out := m.renderWorktreeModal()
	if !strings.Contains(strings.ToLower(out), "worktree") {
		t.Fatalf("create modal should contain worktree text")
	}
}

func TestHistoryWKeyOpensWorktreeModal(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.worktreeEntries = []git.WorktreeEntry{
		{Path: "/repo", Branch: "main"},
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}})
	got := model.(*Model)
	if !got.modal.visible {
		t.Fatalf("W in history screen should open worktree modal")
	}
}
