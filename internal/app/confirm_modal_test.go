package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderConfirmModalShowsMessage(t *testing.T) {
	m := newReadyTestModel()
	m.confirm = &confirmState{Action: confirmCommit, Title: "Confirm commit", Message: "Commit \"feat\"?"}
	out := m.renderConfirmModal()
	for _, want := range []string{"Confirm commit", "Commit \"feat\"?", "Enter/Y", "Esc/N"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in: %s", want, out)
		}
	}
}

func TestConfirmModalEscCancels(t *testing.T) {
	m := newReadyTestModel()
	m.confirm = &confirmState{Action: confirmCommit, Title: "x", Message: "y"}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).confirm != nil {
		t.Fatalf("expected confirm cleared on Esc")
	}
}

func TestConfirmModalNCancels(t *testing.T) {
	m := newReadyTestModel()
	m.confirm = &confirmState{Action: confirmCommit, Title: "x", Message: "y"}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if model.(*Model).confirm != nil {
		t.Fatalf("expected confirm cleared on N")
	}
}

func TestConfirmModalEnterRunsCommit(t *testing.T) {
	m := newReadyTestModel()
	m.commitForm.Title.SetValue("feat: thing")
	m.confirm = &confirmState{Action: confirmCommit, Title: "Confirm commit", Message: "x"}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.(*Model).confirm != nil {
		t.Fatalf("expected confirm cleared after enter")
	}
	if cmd == nil {
		t.Fatalf("expected commit command on confirm")
	}
	if model.(*Model).operationStatus.label != "Committing" {
		t.Fatalf("expected Committing operation status")
	}
}

func TestConfirmModalEnterRunsPR(t *testing.T) {
	m := newReadyTestModel()
	m.prForm.Title.SetValue("feat")
	m.confirm = &confirmState{Action: confirmPR, Title: "PR", Message: "x"}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatalf("expected PR command")
	}
	if model.(*Model).operationStatus.label != "Creating PR" {
		t.Fatalf("expected Creating PR status, got %q", model.(*Model).operationStatus.label)
	}
}
