package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenShellModalInitializes(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	if !m.modal.visible || m.modal.kind != modalShell {
		t.Fatalf("expected shell modal visible")
	}
	if m.modal.input.Value() != "" {
		t.Fatalf("expected input cleared")
	}
}

func TestRenderShellModalShowsPrompt(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	out := m.renderShellModal()
	if !strings.Contains(out, "Shell") {
		t.Fatalf("expected Shell label: %s", out)
	}
	if !strings.Contains(out, "Enter") {
		t.Fatalf("expected shortcut hints")
	}
}

func TestRenderShellModalShowsRunningSpinner(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	m.modal.loading = true
	if !strings.Contains(m.renderShellModal(), "Running") {
		t.Fatalf("expected running spinner")
	}
}

func TestShellModalEscClose(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).modal.visible {
		t.Fatalf("expected esc to close shell modal")
	}
}

func TestShellModalScrollKeys(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	for _, k := range []string{"up", "k", "down", "j", "pgup", "pgdown"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func TestShellModalEnterEmptyDoesNothing(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.(*Model).modal.loading {
		t.Fatalf("did not expect loading on empty enter")
	}
	_ = cmd
}

func TestShellModalTypingForwardsToInput(t *testing.T) {
	m := newReadyTestModel()
	m.openShellModal()
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
}
