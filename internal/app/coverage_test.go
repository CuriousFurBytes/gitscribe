package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestHandleMainSpaceOnEmptyDirShowsNotice(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = []treeRow{{Path: ".", Label: "root/", IsDir: true}}
	m.tree.Index = 0
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !strings.Contains(model.(*Model).notice, "tracked changes") {
		t.Fatalf("expected directory-with-no-changes notice, got %q", model.(*Model).notice)
	}
}

func TestHandleMainSpaceFileWithoutChange(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = []treeRow{{Path: "x", Label: "x", IsDir: false, Change: nil}}
	m.tree.Index = 0
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !strings.Contains(model.(*Model).notice, "Select a file") {
		t.Fatalf("expected select-a-file notice, got %q", model.(*Model).notice)
	}
}

func TestHandleFormFocusTitleAndBodyKeys(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.commitForm.Focus = focusBody
	m.commitForm.applyFocus()

	m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.commitForm.Focus != focusTitle {
		t.Fatalf("shift-tab should focus title")
	}
}

func TestUpdateOperationResultClearPRForm(t *testing.T) {
	m := newReadyTestModel()
	m.prForm.Title.SetValue("draft")
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.Update(operationResultMsg{title: "PR", success: true, clearPR: true, successReturnTo: screenMain})
	if m.prForm.Title.Value() != "" {
		t.Fatalf("expected PR form cleared")
	}
}

func TestUpdateOperationResultAlwaysModal(t *testing.T) {
	m := newReadyTestModel()
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.Update(operationResultMsg{title: "Hook", success: true, alwaysModal: true, modalKind: modalHooks, output: "ok"})
	if !m.modal.visible {
		t.Fatalf("expected modal to stay open with alwaysModal")
	}
}

func TestHandleHooksModalDownClamps(t *testing.T) {
	m := newReadyTestModel()
	m.openHooksModal()
	for i := 0; i < 10; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.modal.hookIndex >= len(m.modal.hooks) {
		t.Fatalf("hookIndex %d out of range", m.modal.hookIndex)
	}
}

func TestHandleHooksModalUpClamps(t *testing.T) {
	m := newReadyTestModel()
	m.openHooksModal()
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.modal.hookIndex != 0 {
		t.Fatalf("expected up at top to stay 0")
	}
}

func TestRenderHistoryEntryNonSelected(t *testing.T) {
	m := newReadyTestModel()
	got := m.renderHistoryEntry(git.CommitHistoryEntry{Hash: "abc", Author: "x", Subject: "y"}, false)
	if !strings.Contains(got, "abc") {
		t.Fatalf("missing hash")
	}
}

func TestUpdateCommitPatchLoadedEmptyContent(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "h"}}
	m.Update(commitPatchLoadedMsg{hash: "h", content: "   "})
	if !strings.Contains(m.historyViewport.View(), "No patch") {
		t.Fatalf("expected placeholder for empty patch")
	}
}

func TestRenderHistoryLoadingNonZeroEntries(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{
		{Hash: "a"}, {Hash: "b"}, {Hash: "c"}, {Hash: "d"}, {Hash: "e"},
	}
	m.historyIndex = 3
	out := m.renderHistoryList(2, 40)
	if out == "" {
		t.Fatalf("expected window of entries")
	}
}

func TestHandleModalUnknownKindNoop(t *testing.T) {
	m := newReadyTestModel()
	m.modal.visible = true
	m.modal.kind = modalKind("unknown")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
}

func TestRenderHistoryWithStatusLine(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.notice = "info"
	out := m.renderHistory()
	if !strings.Contains(out, "info") {
		t.Fatalf("expected notice in status line")
	}
}

func TestLoadSelectionDiffCmdMissingChange(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = []treeRow{{Path: "x", IsDir: false, Change: nil}}
	m.tree.Index = 0
	if m.loadSelectionDiffCmd() != nil {
		t.Fatalf("expected nil cmd for row without change")
	}
}
