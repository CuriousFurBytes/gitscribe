package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestRenderPRScreenShowsLabel(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	out := m.renderPRScreen()
	if !strings.Contains(out, "Pull Request") {
		t.Fatalf("expected Pull Request label: %s", out)
	}
}

func TestPRSubmitOpensConfirmation(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.prForm.Title.SetValue("feat: pr")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if model.(*Model).confirm == nil {
		t.Fatalf("expected PR confirm dialog")
	}
}

func TestPRPrevField(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.prForm.Focus = focusBody
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.(*Model).prForm.Focus != focusTitle {
		t.Fatalf("expected prev-field to land on title")
	}
}

func TestActiveFormReturnsPRForm(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	if m.activeForm() != &m.prForm {
		t.Fatalf("expected PR form active")
	}
}

func TestActiveFormReturnsCommitForm(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	if m.activeForm() != &m.commitForm {
		t.Fatalf("expected commit form active")
	}
}

func TestPRDisabledKeyShowsNotice(t *testing.T) {
	cfg := config.Defaults()
	cfg.PullRequest.Enabled = false
	m := newReadyTestModelWithConfig(cfg)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if !strings.Contains(model.(*Model).notice, "disabled") {
		t.Fatalf("expected disabled notice, got %q", model.(*Model).notice)
	}
}

func TestPROpensFromMainWhenEnabled(t *testing.T) {
	cfg := config.Defaults()
	cfg.PullRequest.Enabled = true
	m := newReadyTestModelWithConfig(cfg)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if model.(*Model).screen != screenPR {
		t.Fatalf("expected PR screen")
	}
}
