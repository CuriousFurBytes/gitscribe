package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/CuriousFurBytes/gitscribe/internal/ai"
	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/execx"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
	ghcli "github.com/CuriousFurBytes/gitscribe/internal/github"
	"github.com/CuriousFurBytes/gitscribe/internal/logger"
	"github.com/CuriousFurBytes/gitscribe/internal/templates"
)

func loadRepoCmd(repoRoot string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		status, err := git.LoadStatus(ctx, repoRoot)
		if err != nil {
			return repoLoadedMsg{err: err}
		}
		branches, err := git.ListBranches(ctx, repoRoot)
		return repoLoadedMsg{status: status, branches: branches, err: err}
	})
}

func loadDiffCmd(repoRoot string, cfg config.GitConfig, targetPath string, change git.FileChange) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		content, mode, err := git.LoadDiff(ctx, repoRoot, cfg, change)
		return diffLoadedMsg{path: targetPath, content: content, mode: mode, err: err}
	})
}

func loadDirectoryDiffCmd(repoRoot string, path string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		content, mode, err := git.LoadDirectoryDiff(ctx, repoRoot, path)
		return diffLoadedMsg{path: path, content: content, mode: mode, err: err}
	})
}

func loadHistoryCmd(repoRoot string, maxCommits int) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		entries, err := git.LoadHistory(ctx, repoRoot, maxCommits)
		return historyLoadedMsg{entries: entries, err: err}
	})
}

func loadCommitPatchCmd(repoRoot string, hash string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		content, err := git.LoadCommitPatch(ctx, repoRoot, hash)
		return commitPatchLoadedMsg{hash: hash, content: content, err: err}
	})
}

func toggleStageCmd(repoRoot string, change git.FileChange) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.ToggleStage(ctx, repoRoot, change)
		return operationResultMsg{
			title:           "Stage",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func toggleStagePathsCmd(repoRoot string, changes []git.FileChange, unstage bool) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.ToggleStagePaths(ctx, repoRoot, changePaths(changes, unstage), unstage)
		title := "Stage"
		if unstage {
			title = "Unstage"
		}
		return operationResultMsg{
			title:           title,
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func loadStashListCmd(repoRoot string) tea.Cmd {
	return withTimeout(10*time.Second, func(ctx context.Context) tea.Msg {
		entries, err := git.ListStashes(ctx, repoRoot)
		return stashListLoadedMsg{entries: entries, err: err}
	})
}

func loadStashDiffCmd(repoRoot, ref string) tea.Cmd {
	return withTimeout(15*time.Second, func(ctx context.Context) tea.Msg {
		content, err := git.ShowStash(ctx, repoRoot, ref)
		return stashDiffLoadedMsg{ref: ref, content: content, err: err}
	})
}

func applyStashCmd(repoRoot, ref string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.ApplyStash(ctx, repoRoot, ref)
		return operationResultMsg{
			title:           "Apply stash",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenStash,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func popStashCmd(repoRoot, ref string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.PopStash(ctx, repoRoot, ref)
		return operationResultMsg{
			title:           "Pop stash",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenStash,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func dropStashCmd(repoRoot, ref string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.DropStash(ctx, repoRoot, ref)
		return operationResultMsg{
			title:           "Drop stash",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenStash,
			failureReturnTo: screenStash,
			refreshRepo:     false,
			err:             err,
		}
	})
}

func loadWorktreesCmd(repoRoot string) tea.Cmd {
	return withTimeout(10*time.Second, func(ctx context.Context) tea.Msg {
		entries, err := git.ListWorktrees(ctx, repoRoot)
		return worktreeListLoadedMsg{entries: entries, err: err}
	})
}

func createWorktreeCmd(repoRoot, path, branch string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.CreateWorktree(ctx, repoRoot, path, branch)
		return operationResultMsg{
			title:           "Create worktree",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     false,
			err:             err,
		}
	})
}

func createBranchCmd(repoRoot, name, base string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.CreateBranch(ctx, repoRoot, name, base)
		return operationResultMsg{
			title:           "Create branch",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func switchBranchCmd(repoRoot string, branch string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.SwitchBranch(ctx, repoRoot, branch)
		return operationResultMsg{
			title:           "Branch switch",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func pushCmd(repoRoot string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.Push(ctx, repoRoot)
		return operationResultMsg{
			title:           "Push",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			err:             err,
		}
	})
}

func pullCmd(repoRoot string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.Pull(ctx, repoRoot)
		return operationResultMsg{
			title:           "Pull",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func commitCmd(repoRoot string, title string, body string, noVerify bool) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.Commit(ctx, repoRoot, strings.TrimSpace(title), strings.TrimRight(body, "\n"), noVerify)
		return operationResultMsg{
			title:           "Commit",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenCommit,
			refreshRepo:     true,
			clearCommit:     err == nil,
			err:             err,
		}
	})
}

func createPRCmd(repoRoot string, cfg config.PullRequestConfig, title string, body string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		result, err := ghcli.CreatePR(ctx, repoRoot, cfg, strings.TrimSpace(title), strings.TrimRight(body, "\n"))
		return operationResultMsg{
			title:           "Pull request",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenPR,
			clearPR:         err == nil,
			err:             err,
		}
	})
}

func aiCmd(repoRoot string, status git.RepoStatus, cfg config.AIConfig, current screen, feedback string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		target := "commit_message"
		if current == screenPR {
			target = "pull_request"
		}
		resp, err := ai.Generate(ctx, repoRoot, status.Branch, status, cfg, target, feedback)
		return aiGeneratedMsg{target: target, resp: resp, err: err}
	})
}

func loadPRTemplateCmd(repoRoot string, customPath string) tea.Cmd {
	return withTimeout(20*time.Second, func(ctx context.Context) tea.Msg {
		_, body, err := templates.DiscoverPRTemplate(repoRoot, customPath)
		return prTemplateLoadedMsg{body: body, err: err}
	})
}

func runPreCommitCmd(repoRoot string, gitDir string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		result, err := git.RunPreCommitHook(ctx, repoRoot, gitDir)
		return operationResultMsg{
			title:           "Pre-commit",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			alwaysModal:     true,
			modalKind:       modalHooks,
			err:             err,
		}
	})
}

func runHookCmd(repoRoot string, gitDir string, hook hookOption) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		var (
			result execx.Result
			err    error
		)
		if hook.PreCommit {
			result, err = git.RunPreCommitHook(ctx, repoRoot, gitDir)
		} else {
			result, err = execx.Run(ctx, repoRoot, nil, hook.Command[0], hook.Command[1:]...)
			if err != nil {
				err = fmt.Errorf("%s: %w\n%s", hook.Name, err, result.Output())
			}
		}
		return operationResultMsg{
			title:           hook.Name,
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			alwaysModal:     true,
			modalKind:       modalHooks,
			err:             err,
		}
	})
}

func runShellCommandCmd(repoRoot string, command string) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		shell := strings.TrimSpace(os.Getenv("SHELL"))
		if shell == "" {
			shell = "/bin/sh"
		}
		result, err := execx.Run(ctx, repoRoot, nil, shell, "-lc", command)
		if err != nil {
			return shellCommandResultMsg{
				output: result.Output(),
				err:    fmt.Errorf("shell command: %w", err),
			}
		}
		return shellCommandResultMsg{output: result.Output()}
	})
}

func openEditorCmd(target screen, focus formFocus, initial string) tea.Cmd {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		return func() tea.Msg {
			return editorFinishedMsg{target: target, focus: focus, err: fmt.Errorf("$EDITOR is not set")}
		}
	}
	args := strings.Fields(editor)
	if len(args) == 0 {
		return func() tea.Msg {
			return editorFinishedMsg{target: target, focus: focus, err: fmt.Errorf("$EDITOR is not set")}
		}
	}
	tmpFile, err := os.CreateTemp("", "gitscribe-editor-*.md")
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{target: target, focus: focus, err: err}
		}
	}
	path := tmpFile.Name()
	if _, err := tmpFile.WriteString(initial); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(path)
		return func() tea.Msg {
			return editorFinishedMsg{target: target, focus: focus, err: err}
		}
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(path)
		return func() tea.Msg {
			return editorFinishedMsg{target: target, focus: focus, err: err}
		}
	}
	cmdArgs := append(args[1:len(args):len(args)], path)
	return tea.ExecProcess(exec.Command(args[0], cmdArgs...), func(err error) tea.Msg {
		defer func() {
			_ = os.Remove(path)
		}()
		if err != nil {
			return editorFinishedMsg{target: target, focus: focus, err: err}
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return editorFinishedMsg{target: target, focus: focus, err: readErr}
		}
		value := strings.TrimRight(string(content), "\n")
		if focus == focusTitle {
			value = strings.TrimSpace(strings.SplitN(value, "\n", 2)[0])
		}
		return editorFinishedMsg{target: target, focus: focus, value: value}
	})
}

func openUserShellCmd() tea.Cmd {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "/bin/sh"
	}
	return tea.ExecProcess(exec.Command(shell), func(err error) tea.Msg {
		return shellCommandResultMsg{err: err}
	})
}

func (m *Model) confirmActionCmd() tea.Cmd {
	switch m.confirm.Action {
	case confirmCommit:
		return commitCmd(m.repo.Root, m.commitForm.Title.Value(), m.commitForm.Body.Value(), m.commitForm.NoVerify)
	case confirmPR:
		return createPRCmd(m.repo.Root, m.cfg.PullRequest, m.prForm.Title.Value(), m.prForm.Body.Value())
	case confirmDiscard:
		change, ok := m.selectedFileChange()
		if !ok {
			return nil
		}
		return discardCmd(m.repo.Root, *change)
	case confirmReset:
		change, ok := m.selectedFileChange()
		if !ok {
			return nil
		}
		return resetCmd(m.repo.Root, *change)
	case confirmStash:
		return stashCmd(m.repo.Root, "")
	case confirmIgnore:
		row, ok := m.selectedTreeRow()
		if !ok {
			return nil
		}
		return ignoreCmd(m.repo.Root, row.Path)
	case confirmStashApply:
		if m.stashIndex < 0 || m.stashIndex >= len(m.stashEntries) {
			return nil
		}
		return applyStashCmd(m.repo.Root, m.stashEntries[m.stashIndex].RefName)
	case confirmStashPop:
		if m.stashIndex < 0 || m.stashIndex >= len(m.stashEntries) {
			return nil
		}
		return popStashCmd(m.repo.Root, m.stashEntries[m.stashIndex].RefName)
	case confirmStashDrop:
		if m.stashIndex < 0 || m.stashIndex >= len(m.stashEntries) {
			return nil
		}
		return dropStashCmd(m.repo.Root, m.stashEntries[m.stashIndex].RefName)
	default:
		return nil
	}
}

func loadRawFileCmd(repoRoot string, path string) tea.Cmd {
	return func() tea.Msg {
		content, err := git.LoadRawFile(repoRoot, path)
		if err != nil {
			logger.Error("load raw file", "path", path, "err", err)
		}
		return rawFileLoadedMsg{path: path, content: content, err: err}
	}
}

func loadPreviewCmd(repoRoot string, path string, syntaxTheme string) tea.Cmd {
	return func() tea.Msg {
		content, err := git.LoadRawFile(repoRoot, path)
		if err != nil {
			logger.Error("load preview file", "path", path, "err", err)
			return previewLoadedMsg{path: path, err: err}
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			rendered, err := glamour.Render(content, "dark")
			if err != nil {
				return previewLoadedMsg{path: path, content: content, err: err}
			}
			return previewLoadedMsg{path: path, content: rendered}
		}

		lexer := lexers.Match(path)
		if lexer == nil {
			lexer = lexers.Fallback
		}

		style := styles.Get(syntaxTheme)
		if style == nil {
			style = styles.Fallback
		}

		formatter := formatters.Get("terminal256")
		if formatter == nil {
			formatter = formatters.Fallback
		}

		iterator, err := lexer.Tokenise(nil, content)
		if err != nil {
			return previewLoadedMsg{path: path, content: content, err: err}
		}

		var buf strings.Builder
		if err := formatter.Format(&buf, style, iterator); err != nil {
			return previewLoadedMsg{path: path, content: content, err: err}
		}
		return previewLoadedMsg{path: path, content: buf.String()}
	}
}

func stashCmd(repoRoot string, name string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("stash", "name", name)
		result, err := git.Stash(ctx, repoRoot, name)
		return operationResultMsg{
			title:           "Stash",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func discardCmd(repoRoot string, change git.FileChange) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("discard", "path", change.Path)
		result, err := git.Discard(ctx, repoRoot, change)
		return operationResultMsg{
			title:           "Discard",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func resetCmd(repoRoot string, change git.FileChange) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("reset", "path", change.Path)
		result, err := git.Reset(ctx, repoRoot, change)
		return operationResultMsg{
			title:           "Reset",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
			err:             err,
		}
	})
}

func amendCmd(repoRoot string, title string, body string, noVerify bool) tea.Cmd {
	return withTimeout(90*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("amend commit", "title", title)
		result, err := git.Amend(ctx, repoRoot, strings.TrimSpace(title), strings.TrimRight(body, "\n"), noVerify)
		return operationResultMsg{
			title:           "Amend",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenCommit,
			refreshRepo:     true,
			clearCommit:     err == nil,
			err:             err,
		}
	})
}

func fetchCmd(repoRoot string) tea.Cmd {
	return withTimeout(60*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("fetch")
		result, err := git.Fetch(ctx, repoRoot)
		return operationResultMsg{
			title:           "Fetch",
			output:          result.Output(),
			success:         err == nil,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     err == nil,
			err:             err,
		}
	})
}

func stageAllCmd(repoRoot string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("stage all")
		_, err := git.StageAll(ctx, repoRoot)
		status, statusErr := git.LoadStatus(ctx, repoRoot)
		if statusErr != nil {
			err = statusErr
		}
		branches, _ := git.ListBranches(ctx, repoRoot)
		return repoLoadedMsg{status: status, branches: branches, err: err}
	})
}

func unstageAllCmd(repoRoot string) tea.Cmd {
	return withTimeout(30*time.Second, func(ctx context.Context) tea.Msg {
		logger.Info("unstage all")
		_, err := git.UnstageAll(ctx, repoRoot)
		status, statusErr := git.LoadStatus(ctx, repoRoot)
		if statusErr != nil {
			err = statusErr
		}
		branches, _ := git.ListBranches(ctx, repoRoot)
		return repoLoadedMsg{status: status, branches: branches, err: err}
	})
}

func copyPathCmd(path string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(path)
		if err != nil {
			logger.Error("copy path to clipboard", "path", path, "err", err)
		}
		return copyPathMsg{path: path, err: err}
	}
}

func ignoreCmd(repoRoot string, path string) tea.Cmd {
	return func() tea.Msg {
		logger.Info("ignore", "path", path)
		err := git.Ignore(repoRoot, path)
		if err != nil {
			return operationResultMsg{
				title:           "Ignore",
				output:          err.Error(),
				success:         false,
				successReturnTo: screenMain,
				failureReturnTo: screenMain,
				err:             err,
			}
		}
		return operationResultMsg{
			title:           "Ignore",
			output:          path + " added to .gitignore",
			success:         true,
			successReturnTo: screenMain,
			failureReturnTo: screenMain,
			refreshRepo:     true,
		}
	}
}

func openFileInEditorCmd(repoRoot string, filePath string, editorCommand string) tea.Cmd {
	editor := strings.TrimSpace(editorCommand)
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if editor == "" {
		return func() tea.Msg {
			return fileEditorFinishedMsg{err: fmt.Errorf("no editor configured ($EDITOR not set)")}
		}
	}
	args := strings.Fields(editor)
	if len(args) == 0 {
		return func() tea.Msg { return fileEditorFinishedMsg{err: fmt.Errorf("no editor configured")} }
	}
	cmdArgs := append(args[1:len(args):len(args)], filepath.Join(repoRoot, filePath))
	return tea.ExecProcess(exec.Command(args[0], cmdArgs...), func(err error) tea.Msg {
		return fileEditorFinishedMsg{err: err}
	})
}
