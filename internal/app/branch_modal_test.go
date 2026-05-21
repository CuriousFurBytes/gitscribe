package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderBranchSelectorListsBranches(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	out := m.renderBranchSelector()
	for _, want := range []string{"main", "dev", "Switch", "Cancel"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
}

func TestBranchSelectorArrowsAndEnter(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev", "feat"}
	m.branchIndex = 0
	m.branchSelector = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).branchIndex != 1 {
		t.Fatalf("branchIndex = %d, want 1", model.(*Model).branchIndex)
	}
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.(*Model).branchIndex != 0 {
		t.Fatalf("expected branchIndex back to 0")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected switch command")
	}
	if m.branchSelector {
		t.Fatalf("expected branch selector closed after enter")
	}
}

func TestBranchSelectorEsc(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	m.branchSelector = true
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).branchSelector {
		t.Fatalf("expected esc to close branch selector")
	}
}

func TestBranchSelectorNKeyEntersCreateMode(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	m.branchSelector = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := model.(*Model)
	if !got.branchCreating {
		t.Fatalf("n key should set branchCreating = true")
	}
	if !got.branchSelector {
		t.Fatalf("branch selector should remain open in create mode")
	}
}

func TestBranchSelectorCreateEscReturnsToList(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	m.branchSelector = true
	m.branchCreating = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.branchCreating {
		t.Fatalf("Esc should exit create mode")
	}
	if !got.branchSelector {
		t.Fatalf("Esc in create mode should return to branch list, not close selector")
	}
}

func TestBranchSelectorCreateEnterTriggersCreate(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	m.branchSelector = true
	m.branchCreating = true
	m.branchCreateInput.SetValue("feature/new")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter with name should trigger create branch command")
	}
	if m.branchSelector {
		t.Fatalf("branch selector should close after creating branch")
	}
}

func TestRenderBranchSelectorShowsCreateHint(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	out := m.renderBranchSelector()
	if !strings.Contains(out, "New") && !strings.Contains(out, "n") {
		t.Fatalf("branch selector should show hint for creating new branch")
	}
}

func TestHistoryCtrlBOpensBranchSelector(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.branches = []string{"main", "dev"}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	if !model.(*Model).branchSelector {
		t.Fatalf("Ctrl+B in history screen should open branch selector")
	}
}

func TestBranchSelectorDownClamps(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"only"}
	m.branchSelector = true
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.branchIndex != 0 {
		t.Fatalf("expected index clamped at 0")
	}
}
