package app

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenLogsModalSuccess(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Commit", output: "", success: true})
	if !m.modal.visible {
		t.Fatalf("expected logs modal visible")
	}
	if m.modal.kind != modalLogs {
		t.Fatalf("kind = %q", m.modal.kind)
	}
	if !strings.Contains(m.modal.viewport.View(), "successfully") {
		t.Fatalf("expected success message in viewport")
	}
}

func TestOpenLogsModalErrorContent(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Push", success: false, err: errors.New("boom")})
	if !strings.Contains(m.modal.viewport.View(), "boom") {
		t.Fatalf("expected error to land in viewport")
	}
}

func TestOpenLogsModalWithKind(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Pre-commit", success: false, output: "log\nlog2", modalKind: modalHooks})
	if m.modal.kind != modalHooks {
		t.Fatalf("kind override = %q, want %q", m.modal.kind, modalHooks)
	}
}

func TestRenderScrollableModalSuccessHeader(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Pull", success: true, output: "ok"})
	out := m.renderScrollableModal()
	if !strings.Contains(out, "Pull") {
		t.Fatalf("expected title in render: %s", out)
	}
}

func TestRenderScrollableModalLoading(t *testing.T) {
	m := newReadyTestModel()
	m.modal.visible = true
	m.modal.kind = modalLogs
	m.modal.loading = true
	m.modal.title = "Working"
	m.modal.body = "details"
	out := m.renderScrollableModal()
	if !strings.Contains(out, "Working") {
		t.Fatalf("expected loading content: %s", out)
	}
}

func TestLogsModalScrollKeys(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Logs", output: strings.Repeat("a\n", 100), success: true})
	for _, k := range []string{"down", "j", "pgdown", "up", "k", "pgup"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func TestLogsModalEnterRefreshesWhenRequested(t *testing.T) {
	m := newReadyTestModel()
	m.openLogsModal(operationResultMsg{title: "Commit", success: true, refreshRepo: true})
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.(*Model).modal.visible {
		t.Fatalf("modal should close")
	}
	if cmd == nil {
		t.Fatalf("expected refresh command on close")
	}
}
