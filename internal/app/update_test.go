package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/ai"
	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestUpdateWindowSizeMakesReady(t *testing.T) {
	m := newReadyTestModel()
	m.ready = false
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !model.(*Model).ready {
		t.Fatalf("expected ready=true after window size")
	}
}

func TestUpdateTitleTickIncrements(t *testing.T) {
	m := newReadyTestModel()
	m.titleAnimationUntil = timeNow().Add(10 * time.Second)
	before := m.titleAnimationFrame
	m.Update(titleTickMsg{})
	if m.titleAnimationFrame != before+1 {
		t.Fatalf("expected frame increment")
	}
}

func TestUpdateRepoLoadedError(t *testing.T) {
	m := newReadyTestModel()
	model, _ := m.Update(repoLoadedMsg{err: errors.New("boom")})
	if model.(*Model).notice == "" {
		t.Fatalf("expected notice on repo error")
	}
}

func TestUpdateDiffLoadedSetsContent(t *testing.T) {
	m := newReadyTestModel()
	if row, ok := m.selectedTreeRow(); ok {
		m.Update(diffLoadedMsg{path: row.Path, content: "diff body", mode: "Working"})
		if m.diffContent != "diff body" {
			t.Fatalf("diff content = %q", m.diffContent)
		}
	}
}

func TestUpdateDiffLoadedMismatchedPathIgnored(t *testing.T) {
	m := newReadyTestModel()
	before := m.diffContent
	m.Update(diffLoadedMsg{path: "nope/elsewhere", content: "x"})
	if m.diffContent != before {
		t.Fatalf("expected diff unchanged for mismatched path")
	}
}

func TestUpdateDiffLoadedError(t *testing.T) {
	m := newReadyTestModel()
	row, _ := m.selectedTreeRow()
	m.Update(diffLoadedMsg{path: row.Path, err: errors.New("oops")})
	if m.notice == "" {
		t.Fatalf("expected notice")
	}
}

func TestUpdateHistoryLoadedEmptyShowsPlaceholder(t *testing.T) {
	m := newReadyTestModel()
	m.Update(historyLoadedMsg{entries: nil})
	if !strings.Contains(m.historyViewport.View(), "No commits") {
		t.Fatalf("expected placeholder")
	}
}

func TestUpdateHistoryLoadedError(t *testing.T) {
	m := newReadyTestModel()
	m.Update(historyLoadedMsg{err: errors.New("oops")})
	if m.notice == "" {
		t.Fatalf("expected notice")
	}
}

func TestUpdateHistoryLoadedTrimsIndex(t *testing.T) {
	m := newReadyTestModel()
	m.historyIndex = 5
	m.Update(historyLoadedMsg{entries: []git.CommitHistoryEntry{{Hash: "a"}}})
	if m.historyIndex != 0 {
		t.Fatalf("expected historyIndex clamped, got %d", m.historyIndex)
	}
}

func TestUpdateCommitPatchLoaded(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "abc"}}
	m.Update(commitPatchLoadedMsg{hash: "abc", content: "diff --git body"})
	if !strings.Contains(m.historyViewport.View(), "body") {
		t.Fatalf("expected patch content in viewport")
	}
}

func TestUpdateCommitPatchLoadedMismatched(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "abc"}}
	m.Update(commitPatchLoadedMsg{hash: "xyz", content: "ignored"})
}

func TestUpdateCommitPatchLoadedError(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "abc"}}
	m.Update(commitPatchLoadedMsg{hash: "abc", err: errors.New("nope")})
	if m.notice == "" {
		t.Fatalf("expected notice on patch error")
	}
}

func TestUpdateAIGeneratedAppliesResponse(t *testing.T) {
	cfg := config.Defaults()
	cfg.AI.Enabled = true
	cfg.AI.Mode = "full"
	m := newReadyTestModelWithConfig(cfg)
	m.screen = screenCommit
	m.commitForm.Loading = true
	m.Update(aiGeneratedMsg{target: "commit_message", resp: ai.Response{Title: "feat: ai", Body: "body"}})
	if m.commitForm.Title.Value() != "feat: ai" {
		t.Fatalf("title = %q", m.commitForm.Title.Value())
	}
	if m.commitForm.Loading {
		t.Fatalf("expected loading cleared")
	}
}

func TestUpdateAIGeneratedTitleOnlySkipsBody(t *testing.T) {
	cfg := config.Defaults()
	cfg.AI.Enabled = true
	cfg.AI.Mode = "title_only"
	m := newReadyTestModelWithConfig(cfg)
	m.screen = screenCommit
	m.Update(aiGeneratedMsg{target: "commit_message", resp: ai.Response{Title: "t", Body: "ignored"}})
	if m.commitForm.Body.Value() != "" {
		t.Fatalf("body should stay empty, got %q", m.commitForm.Body.Value())
	}
}

func TestUpdateAIGeneratedError(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.commitForm.Loading = true
	m.Update(aiGeneratedMsg{err: errors.New("boom")})
	if m.commitForm.Error == "" {
		t.Fatalf("expected error set on form")
	}
}

func TestUpdatePRTemplateLoaded(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.Update(prTemplateLoadedMsg{body: "Template"})
	if m.prForm.Body.Value() != "Template" {
		t.Fatalf("expected template body applied")
	}
}

func TestUpdatePRTemplateError(t *testing.T) {
	m := newReadyTestModel()
	m.Update(prTemplateLoadedMsg{err: errors.New("missing")})
	if m.prForm.Error == "" {
		t.Fatalf("expected pr form error set")
	}
}

func TestUpdateShellCommandResultSuccess(t *testing.T) {
	m := newReadyTestModel()
	m.modal.loading = true
	m.Update(shellCommandResultMsg{output: "ok"})
	if m.modal.loading {
		t.Fatalf("expected loading cleared")
	}
}

func TestUpdateShellCommandResultError(t *testing.T) {
	m := newReadyTestModel()
	m.Update(shellCommandResultMsg{err: errors.New("nope")})
	if !strings.Contains(m.modal.viewport.View(), "nope") {
		t.Fatalf("expected error displayed")
	}
}

func TestUpdateEditorFinishedAppliesTitle(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.Update(editorFinishedMsg{target: screenCommit, focus: focusTitle, value: "edited"})
	if m.commitForm.Title.Value() != "edited" {
		t.Fatalf("expected title applied")
	}
}

func TestUpdateEditorFinishedAppliesBody(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPR
	m.Update(editorFinishedMsg{target: screenPR, focus: focusBody, value: "body text"})
	if m.prForm.Body.Value() != "body text" {
		t.Fatalf("expected body applied")
	}
}

func TestUpdateEditorFinishedError(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	m.Update(editorFinishedMsg{target: screenCommit, focus: focusTitle, err: errors.New("nope")})
	if m.commitForm.Error == "" {
		t.Fatalf("expected error set on commit form")
	}
}

func TestUpdateRoutesGlobalKeyToHelp(t *testing.T) {
	m := newReadyTestModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !model.(*Model).modal.visible {
		t.Fatalf("expected ? to open help modal")
	}
}

func TestUpdateOperationResultCommitSuccessShowsToast(t *testing.T) {
	m := newReadyTestModel()
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.screen = screenCommit
	model, _ := m.Update(operationResultMsg{
		title:           "Commit",
		output:          "ok",
		success:         true,
		successReturnTo: screenMain,
		clearCommit:     true,
	})
	got := model.(*Model)
	if !strings.Contains(got.View(), "Commit created") {
		t.Fatalf("expected toast 'Commit created' in View, got:\n%s", got.View())
	}
}

func TestUpdateOperationResultPRSuccessShowsToast(t *testing.T) {
	m := newReadyTestModel()
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.screen = screenPR
	model, _ := m.Update(operationResultMsg{
		title:           "Pull request",
		output:          "ok",
		success:         true,
		successReturnTo: screenMain,
		clearPR:         true,
	})
	got := model.(*Model)
	if !strings.Contains(got.View(), "Pull request created") {
		t.Fatalf("expected toast 'Pull request created' in View, got:\n%s", got.View())
	}
}

func TestUpdateOperationResultFailureSurfacesStderr(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenCommit
	model, _ := m.Update(operationResultMsg{
		title:           "Commit",
		output:          "",
		stderr:          "error: pre-commit hook failed",
		success:         false,
		failureReturnTo: screenCommit,
		err:             errors.New("git commit: exit code 1"),
	})
	got := model.(*Model)
	if !strings.Contains(got.View(), "error: pre-commit hook failed") {
		t.Fatalf("expected View to surface stderr text, got:\n%s", got.View())
	}
}

func TestUpdateOperationResultAutoCloseRefreshes(t *testing.T) {
	m := newReadyTestModel()
	m.cfg.Logs.AutoCloseOnSuccess = true
	m.screen = screenCommit
	m.modal.visible = true
	_, _ = m.Update(operationResultMsg{title: "Commit", success: true, refreshRepo: true, successReturnTo: screenMain, clearCommit: true})
	if m.modal.visible {
		t.Fatalf("expected modal auto-closed")
	}
	if m.screen != screenMain {
		t.Fatalf("expected return to main screen")
	}
}
