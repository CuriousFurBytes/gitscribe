package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestRenderCommitScreenShowsTitleSection(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.commitForm.Title.SetValue("feat: x")
	out := m.renderCommitScreen()
	if !strings.Contains(out, "Commit Title") {
		t.Fatalf("expected Commit Title label: %s", out)
	}
}

func TestCommitSubmitWithoutTitleSetsError(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if model.(*Model).commitForm.Error == "" {
		t.Fatalf("expected validation error")
	}
}

func TestCommitSubmitWithConfirmationOpensDialog(t *testing.T) {
	cfg := config.Defaults()
	cfg.Commit.RequireConfirmation = true
	m := newReadyTestModelWithConfig(cfg)
	m.screen = screenCommit
	m.commitForm.Title.SetValue("feat: ok")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if model.(*Model).confirm == nil {
		t.Fatalf("expected confirm dialog")
	}
}

func TestCommitSubmitDirectStartsOperation(t *testing.T) {
	cfg := config.Defaults()
	cfg.Commit.RequireConfirmation = false
	m := newReadyTestModelWithConfig(cfg)
	m.screen = screenCommit
	m.commitForm.Title.SetValue("feat: ok")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected commit command")
	}
	if model.(*Model).operationStatus.label != "Committing" {
		t.Fatalf("expected Committing operation")
	}
}

func TestCommitCancelKeyReturnsToMain(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).screen != screenMain {
		t.Fatalf("expected screen reset to main")
	}
}

func TestCommitTabSwitchesFocus(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model.(*Model).commitForm.Focus != focusBody {
		t.Fatalf("expected focus to move to body")
	}
}

func TestCommitAIKey(t *testing.T) {
	cfg := config.Defaults()
	cfg.AI.Enabled = true
	m := newReadyTestModelWithConfig(cfg)
	m.screen = screenCommit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd == nil {
		t.Fatalf("expected AI command")
	}
	if !m.commitForm.Loading {
		t.Fatalf("expected loading flag")
	}
}
