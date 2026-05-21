package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenHelpModalSetsFields(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	if !m.modal.visible {
		t.Fatalf("expected help modal to be visible")
	}
	if m.modal.kind != modalHelp {
		t.Fatalf("kind = %q, want %q", m.modal.kind, modalHelp)
	}
	if m.modal.title != "Help" {
		t.Fatalf("title = %q", m.modal.title)
	}
}

func TestRenderHelpModalIncludesSections(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	for _, want := range []string{"Workspace", "Commit flows", "Modals", "Esc/?"} {
		if !strings.Contains(out, want) {
			t.Fatalf("renderHelpModal missing %q: %s", want, out)
		}
	}
}

func TestRenderHelpModalContainsWorkspaceAndModals(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	for _, want := range []string{"Workspace", "Commit flows", "Ctrl+G", "Esc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("renderHelpModal missing %q", want)
		}
	}
}

func TestHelpModalCapitalizesCtrl(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	if strings.Contains(out, "ctrl+") {
		t.Fatalf("renderHelpModal contains lowercase ctrl+ prefix")
	}
}

func TestHelpModalIncludesStashHint(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	if !strings.Contains(strings.ToLower(out), "stash") {
		t.Fatalf("renderHelpModal missing stash hint")
	}
}

func TestHelpModalIncludesWorktreeHint(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	if !strings.Contains(strings.ToLower(out), "worktree") {
		t.Fatalf("renderHelpModal missing worktree hint")
	}
}

func TestHelpModalIncludesCtrlBBranchHint(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	out := m.renderHelpModal()
	if !strings.Contains(out, "Ctrl+B") {
		t.Fatalf("renderHelpModal missing Ctrl+B branch hint")
	}
}

func TestHelpModalEscClose(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).modal.visible {
		t.Fatalf("expected help modal closed on Esc")
	}
}
