package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestOpenPRURLModalSetsFields(t *testing.T) {
	m := newReadyTestModel()
	m.openPRURLModal("https://github.com/owner/repo/pull/42")

	if !m.modal.visible {
		t.Fatalf("expected PR URL modal visible")
	}
	if m.modal.kind != modalPRURL {
		t.Fatalf("kind = %q, want %q", m.modal.kind, modalPRURL)
	}
	if !strings.Contains(m.modal.body, "https://github.com/owner/repo/pull/42") {
		t.Fatalf("body should contain URL, got %q", m.modal.body)
	}
}

func TestRenderPRURLModalShowsURLAndHints(t *testing.T) {
	m := newReadyTestModel()
	m.openPRURLModal("https://github.com/owner/repo/pull/7")
	out := m.renderPRURLModal()

	if !strings.Contains(out, "https://github.com/owner/repo/pull/7") {
		t.Fatalf("expected URL in render: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "esc") {
		t.Fatalf("expected close hint in render: %s", out)
	}
}

func TestPRURLModalEscClosesModal(t *testing.T) {
	m := newReadyTestModel()
	m.openPRURLModal("https://github.com/owner/repo/pull/9")

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).modal.visible {
		t.Fatalf("expected PR URL modal to close on Esc")
	}
}

func TestPRURLModalEnterClosesModal(t *testing.T) {
	m := newReadyTestModel()
	m.openPRURLModal("https://github.com/owner/repo/pull/9")

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.(*Model).modal.visible {
		t.Fatalf("expected PR URL modal to close on Enter")
	}
}

func TestOperationResultShowsPRURLModalWhenConfigured(t *testing.T) {
	cfg := config.Defaults()
	cfg.Logs.AutoCloseOnSuccess = true
	cfg.PullRequest.AfterCreate = "modal"
	m := newReadyTestModelWithConfig(cfg)

	msg := operationResultMsg{
		title:           "Pull request",
		success:         true,
		successReturnTo: screenMain,
		clearPR:         true,
		prURL:           "https://github.com/owner/repo/pull/42",
	}
	m.updateOperationResult(msg, nil)

	if !m.modal.visible {
		t.Fatalf("expected PR URL modal to be visible after PR create with after_create=modal")
	}
	if m.modal.kind != modalPRURL {
		t.Fatalf("kind = %q, want %q", m.modal.kind, modalPRURL)
	}
	if !strings.Contains(m.modal.body, "pull/42") {
		t.Fatalf("expected modal body to contain the URL, got %q", m.modal.body)
	}
}

func TestOperationResultSkipsModalWhenAfterCreateNone(t *testing.T) {
	cfg := config.Defaults()
	cfg.Logs.AutoCloseOnSuccess = true
	cfg.PullRequest.AfterCreate = "none"
	m := newReadyTestModelWithConfig(cfg)

	msg := operationResultMsg{
		title:           "Pull request",
		success:         true,
		successReturnTo: screenMain,
		clearPR:         true,
		prURL:           "https://github.com/owner/repo/pull/42",
	}
	m.updateOperationResult(msg, nil)

	if m.modal.visible && m.modal.kind == modalPRURL {
		t.Fatalf("expected no PR URL modal when after_create=none")
	}
}
