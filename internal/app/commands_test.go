package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func mustGit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
	}
	return dir
}

func runCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatalf("nil cmd")
	}
	return cmd()
}

func TestLoadRepoCmdReturnsStatus(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadRepoCmd(dir))
	if _, ok := msg.(repoLoadedMsg); !ok {
		t.Fatalf("expected repoLoadedMsg, got %T", msg)
	}
}

func TestLoadDiffCmdReturnsDiff(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadDiffCmd(dir, config.GitConfig{}, "missing.go", git.FileChange{Path: "missing.go", UnstagedStatus: git.StatusModified}))
	if _, ok := msg.(diffLoadedMsg); !ok {
		t.Fatalf("expected diffLoadedMsg, got %T", msg)
	}
}

func TestLoadDirectoryDiffCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadDirectoryDiffCmd(dir, "."))
	if _, ok := msg.(diffLoadedMsg); !ok {
		t.Fatalf("expected diffLoadedMsg")
	}
}

func TestLoadHistoryCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadHistoryCmd(dir, 5))
	if _, ok := msg.(historyLoadedMsg); !ok {
		t.Fatalf("expected historyLoadedMsg")
	}
}

func TestLoadCommitPatchCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadCommitPatchCmd(dir, "HEAD"))
	if _, ok := msg.(commitPatchLoadedMsg); !ok {
		t.Fatalf("expected commitPatchLoadedMsg")
	}
}

func TestToggleStageCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, toggleStageCmd(dir, git.FileChange{Path: "missing.go", IsUntracked: true, UnstagedStatus: git.StatusUntracked}))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestToggleStagePathsCmdStage(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, toggleStagePathsCmd(dir, []git.FileChange{{Path: "missing.go", IsUntracked: true, UnstagedStatus: git.StatusUntracked}}, false))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestToggleStagePathsCmdUnstage(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, toggleStagePathsCmd(dir, []git.FileChange{{Path: "f.go", StagedStatus: git.StatusModified}}, true))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestSwitchBranchCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, switchBranchCmd(dir, "nope"))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestPushCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, pushCmd(dir))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestPullCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, pullCmd(dir))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestCommitCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, commitCmd(dir, "title", "body", false))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestCreatePRCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, createPRCmd(dir, config.PullRequestConfig{Enabled: true}, "title", "body"))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestLoadPRTemplateCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadPRTemplateCmd(dir, ""))
	if _, ok := msg.(prTemplateLoadedMsg); !ok {
		t.Fatalf("expected prTemplateLoadedMsg")
	}
}

func TestRunPreCommitCmd(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runPreCommitCmd(dir, filepath.Join(dir, ".git")))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestRunHookCmdPreCommit(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runHookCmd(dir, filepath.Join(dir, ".git"), hookOption{Name: "pc", PreCommit: true}))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestRunHookCmdCustom(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runHookCmd(dir, filepath.Join(dir, ".git"), hookOption{Name: "echo", Command: []string{"echo", "hi"}}))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg")
	}
}

func TestRunHookCmdCustomFailure(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runHookCmd(dir, filepath.Join(dir, ".git"), hookOption{Name: "nope", Command: []string{"definitely-not-a-real-command"}}))
	result, ok := msg.(operationResultMsg)
	if !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
	if result.success {
		t.Fatalf("expected failure for missing command")
	}
}

func TestRunShellCommandCmdSuccess(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runShellCommandCmd(dir, "echo hi"))
	if _, ok := msg.(shellCommandResultMsg); !ok {
		t.Fatalf("expected shellCommandResultMsg")
	}
}

func TestRunShellCommandCmdFailure(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, runShellCommandCmd(dir, "exit 1"))
	r, ok := msg.(shellCommandResultMsg)
	if !ok {
		t.Fatalf("expected shellCommandResultMsg, got %T", msg)
	}
	if r.err == nil {
		t.Fatalf("expected error on exit 1")
	}
}

func TestAICmdReturnsMessageWhenDisabled(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, aiCmd(dir, git.RepoStatus{}, config.AIConfig{Enabled: false}, screenCommit, ""))
	if _, ok := msg.(aiGeneratedMsg); !ok {
		t.Fatalf("expected aiGeneratedMsg")
	}
}

func TestAICmdPRTarget(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, aiCmd(dir, git.RepoStatus{}, config.AIConfig{Enabled: false}, screenPR, ""))
	a, ok := msg.(aiGeneratedMsg)
	if !ok {
		t.Fatalf("expected aiGeneratedMsg")
	}
	if a.target != "pull_request" {
		t.Fatalf("target = %q, want pull_request", a.target)
	}
}

func TestOpenEditorCmdEmptyEditorString(t *testing.T) {
	old := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "   ")
	defer os.Setenv("EDITOR", old)

	msg := runCmd(t, openEditorCmd(screenCommit, focusTitle, ""))
	m, _ := msg.(editorFinishedMsg)
	if m.err == nil {
		t.Fatalf("expected error on blank editor")
	}
}

func TestOpenEditorCmdWithEditorRuns(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("'true' command not available")
	}
	old := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "true")
	defer os.Setenv("EDITOR", old)
	cmd := openEditorCmd(screenCommit, focusBody, "initial body")
	if cmd == nil {
		t.Fatalf("expected non-nil cmd when EDITOR set")
	}
}

func TestOpenEditorCmdWithoutEnv(t *testing.T) {
	old := os.Getenv("EDITOR")
	os.Unsetenv("EDITOR")
	defer os.Setenv("EDITOR", old)

	cmd := openEditorCmd(screenCommit, focusTitle, "initial")
	msg := runCmd(t, cmd)
	m, ok := msg.(editorFinishedMsg)
	if !ok {
		t.Fatalf("expected editorFinishedMsg")
	}
	if m.err == nil {
		t.Fatalf("expected error when EDITOR unset")
	}
}

func TestWithTimeoutPassesContext(t *testing.T) {
	cmd := withTimeout(time.Second, func(ctx context.Context) tea.Msg {
		if _, ok := ctx.Deadline(); !ok {
			t.Errorf("expected ctx deadline")
		}
		return "ok"
	})
	if cmd() != tea.Msg("ok") {
		t.Fatalf("unexpected result")
	}
}

func TestConfirmActionCmdReturnsNilForUnknown(t *testing.T) {
	m := newReadyTestModel()
	m.confirm = &confirmState{Action: "unknown"}
	if m.confirmActionCmd() != nil {
		t.Fatalf("expected nil for unknown action")
	}
}

func TestOpenUserShellCmdReturnsCmd(t *testing.T) {
	old := os.Getenv("SHELL")
	os.Unsetenv("SHELL")
	defer os.Setenv("SHELL", old)
	if openUserShellCmd() == nil {
		t.Fatalf("expected non-nil cmd")
	}
	os.Setenv("SHELL", "/bin/sh")
	if openUserShellCmd() == nil {
		t.Fatalf("expected non-nil cmd when SHELL set")
	}
}

func TestTitleTickCmdReturnsTitleTickMsg(t *testing.T) {
	msg := runCmd(t, titleTickCmd())
	if _, ok := msg.(titleTickMsg); !ok {
		t.Fatalf("expected titleTickMsg, got %T", msg)
	}
}

func TestStashCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, stashCmd(dir, ""))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestDiscardCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, discardCmd(dir, git.FileChange{Path: "missing.go", UnstagedStatus: git.StatusModified}))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestResetCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, resetCmd(dir, git.FileChange{Path: "missing.go", StagedStatus: git.StatusModified}))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestAmendCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, amendCmd(dir, "new title", "", false))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestFetchCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, fetchCmd(dir))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestStageAllCmdReturnsRepoLoadedMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, stageAllCmd(dir))
	if _, ok := msg.(repoLoadedMsg); !ok {
		t.Fatalf("expected repoLoadedMsg, got %T", msg)
	}
}

func TestUnstageAllCmdReturnsRepoLoadedMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, unstageAllCmd(dir))
	if _, ok := msg.(repoLoadedMsg); !ok {
		t.Fatalf("expected repoLoadedMsg, got %T", msg)
	}
}

func TestIgnoreCmdReturnsOperationResultMsg(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, ignoreCmd(dir, "ignored-file.txt"))
	if _, ok := msg.(operationResultMsg); !ok {
		t.Fatalf("expected operationResultMsg, got %T", msg)
	}
}

func TestLoadRawFileCmdReturnsRawFileLoadedMsg(t *testing.T) {
	dir := mustGit(t)
	// Write a file to load
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	msg := runCmd(t, loadRawFileCmd(dir, "hello.txt"))
	r, ok := msg.(rawFileLoadedMsg)
	if !ok {
		t.Fatalf("expected rawFileLoadedMsg, got %T", msg)
	}
	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if r.content != "hello\n" {
		t.Fatalf("content = %q, want %q", r.content, "hello\n")
	}
}

func TestLoadRawFileCmdMissingFile(t *testing.T) {
	dir := mustGit(t)
	msg := runCmd(t, loadRawFileCmd(dir, "does-not-exist.txt"))
	r, ok := msg.(rawFileLoadedMsg)
	if !ok {
		t.Fatalf("expected rawFileLoadedMsg")
	}
	if r.err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestLoadPreviewCmdReturnsPreviewMsg(t *testing.T) {
	dir := mustGit(t)
	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	msg := runCmd(t, loadPreviewCmd(dir, "hello.go", "dracula"))
	if _, ok := msg.(previewLoadedMsg); !ok {
		t.Fatalf("expected previewLoadedMsg, got %T", msg)
	}
}

func TestLoadPreviewCmdMarkdown(t *testing.T) {
	dir := mustGit(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	msg := runCmd(t, loadPreviewCmd(dir, "README.md", "dracula"))
	r, ok := msg.(previewLoadedMsg)
	if !ok {
		t.Fatalf("expected previewLoadedMsg, got %T", msg)
	}
	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
}

func TestCopyPathCmdReturnsCopyPathMsg(t *testing.T) {
	msg := runCmd(t, copyPathCmd("some/path/file.go"))
	if _, ok := msg.(copyPathMsg); !ok {
		t.Fatalf("expected copyPathMsg, got %T", msg)
	}
}

func TestConfirmActionCmdHandlesDiscard(t *testing.T) {
	dir := mustGit(t)
	m := newReadyTestModel()
	m.repo.Root = dir
	m.status.RepoStatus.Files = []git.FileChange{{Path: "f.go", UnstagedStatus: git.StatusModified}}
	m.rebuildTree()
	// Select the file row (not the root dir row)
	for i, row := range m.tree.Rows {
		if !row.IsDir && row.Change != nil {
			m.tree.Index = i
			break
		}
	}
	m.confirm = &confirmState{Action: confirmDiscard}
	cmd := m.confirmActionCmd()
	if cmd == nil {
		t.Fatalf("expected non-nil cmd for confirmDiscard")
	}
}

func TestConfirmActionCmdHandlesStash(t *testing.T) {
	dir := mustGit(t)
	m := newReadyTestModel()
	m.repo.Root = dir
	m.confirm = &confirmState{Action: confirmStash}
	cmd := m.confirmActionCmd()
	if cmd == nil {
		t.Fatalf("expected non-nil cmd for confirmStash")
	}
}

func TestOpenFileInEditorCmdWithEmptyEditorUsesEnv(t *testing.T) {
	dir := mustGit(t)
	cmd := openFileInEditorCmd(dir, "file.go", "")
	if cmd == nil {
		t.Fatalf("expected non-nil cmd")
	}
}
