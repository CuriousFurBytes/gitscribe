package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestOpenHooksModalIncludesConfigHooks(t *testing.T) {
	cfg := config.Defaults()
	cfg.Hooks = []config.HookConfig{{Name: "Lint", Command: []string{"make", "lint"}}}
	m := newReadyTestModelWithConfig(cfg)
	m.openHooksModal()
	if len(m.modal.hooks) != 2 {
		t.Fatalf("hooks count = %d, want 2", len(m.modal.hooks))
	}
	if m.modal.hooks[1].Name != "Lint" {
		t.Fatalf("second hook = %q", m.modal.hooks[1].Name)
	}
}

func TestRenderHooksModalSelectingShowsList(t *testing.T) {
	cfg := config.Defaults()
	cfg.Hooks = []config.HookConfig{{Name: "Test", Command: []string{"go", "test"}}}
	m := newReadyTestModelWithConfig(cfg)
	m.openHooksModal()
	out := m.renderHooksModal()
	if !strings.Contains(out, "Pre-commit") || !strings.Contains(out, "Test") {
		t.Fatalf("expected both hooks rendered: %s", out)
	}
}

func TestRenderHooksModalAfterSelectionShowsLogs(t *testing.T) {
	m := newReadyTestModel()
	m.openHooksModal()
	m.modal.selecting = false
	m.modal.title = "Pre-commit"
	out := m.renderHooksModal()
	if !strings.Contains(out, "Pre-commit") {
		t.Fatalf("expected logs-style render: %s", out)
	}
}

func TestHooksModalArrowKeysNavigate(t *testing.T) {
	cfg := config.Defaults()
	cfg.Hooks = []config.HookConfig{{Name: "Lint", Command: []string{"make", "lint"}}}
	m := newReadyTestModelWithConfig(cfg)
	m.openHooksModal()
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.modal.hookIndex != 1 {
		t.Fatalf("hookIndex = %d, want 1", m.modal.hookIndex)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.modal.hookIndex != 0 {
		t.Fatalf("hookIndex = %d, want 0", m.modal.hookIndex)
	}
}

func TestHooksModalEnterRunsHook(t *testing.T) {
	m := newReadyTestModel()
	m.openHooksModal()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected hook run command")
	}
	if m.modal.selecting {
		t.Fatalf("expected selecting=false after enter")
	}
}

func TestHooksModalEscClose(t *testing.T) {
	m := newReadyTestModel()
	m.openHooksModal()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).modal.visible {
		t.Fatalf("expected esc to close hooks modal")
	}
}

func TestSelectedHookOutOfRange(t *testing.T) {
	m := newReadyTestModel()
	m.modal.hookIndex = 5
	if _, ok := m.selectedHook(); ok {
		t.Fatalf("expected out-of-range to return false")
	}
}
