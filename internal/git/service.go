package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

var hexSHAPattern = regexp.MustCompile(`^[0-9a-f]{1,40}$`)
var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_.\-]{0,254}$`)

const (
	DiffModeAuto     = "auto"
	DiffModeStaged   = "staged"
	DiffModeUnstaged = "unstaged"
)

func LoadStatus(ctx context.Context, repoRoot string) (RepoStatus, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "status", "--porcelain=v2", "--branch")
	if err != nil {
		return RepoStatus{}, fmt.Errorf("git status: %w\n%s", err, result.Output())
	}
	status, err := ParseStatus(result.Stdout)
	if err != nil {
		return status, err
	}
	if status.HasStaged {
		stats, statsErr := LoadStagedDiffStats(ctx, repoRoot)
		if statsErr != nil {
			return status, statsErr
		}
		status.StagedStats = stats
	}
	if status.Ahead > 0 {
		commits, commitsErr := LoadBranchCommits(ctx, repoRoot, DefaultBranchCommitsLimit)
		if commitsErr == nil {
			status.BranchCommits = commits
		}
	}
	return status, nil
}

func ListBranches(ctx context.Context, repoRoot string) ([]string, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("git branch: %w\n%s", err, result.Output())
	}

	var branches []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

func SwitchBranch(ctx context.Context, repoRoot string, branch string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "switch", branch)
	if err == nil {
		return result, nil
	}

	fallback, fallbackErr := execx.Run(ctx, repoRoot, nil, "git", "checkout", branch)
	if fallbackErr != nil {
		if fallback.Output() != "" {
			return fallback, fmt.Errorf("git checkout: %w\n%s", fallbackErr, fallback.Output())
		}
		return fallback, fmt.Errorf("git checkout: %w", fallbackErr)
	}
	return fallback, nil
}

func ToggleStage(ctx context.Context, repoRoot string, change FileChange) (execx.Result, error) {
	if change.HasStaged() {
		result, err := execx.Run(ctx, repoRoot, nil, "git", "restore", "--staged", "--", change.Path)
		if err == nil {
			return result, nil
		}

		fallback, fallbackErr := execx.Run(ctx, repoRoot, nil, "git", "reset", "HEAD", "--", change.Path)
		if fallbackErr != nil {
			return fallback, fmt.Errorf("unstage file: %w\n%s", fallbackErr, fallback.Output())
		}
		return fallback, nil
	}

	result, err := execx.Run(ctx, repoRoot, nil, "git", "add", "--", change.Path)
	if err != nil {
		return result, fmt.Errorf("stage file: %w\n%s", err, result.Output())
	}
	return result, nil
}

func ToggleStagePaths(ctx context.Context, repoRoot string, paths []string, unstage bool) (execx.Result, error) {
	paths = uniqueNonEmptyPaths(paths)
	if len(paths) == 0 {
		return execx.Result{}, nil
	}

	if unstage {
		args := append([]string{"restore", "--staged", "--"}, paths...)
		result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
		if err == nil {
			return result, nil
		}

		fallbackArgs := append([]string{"reset", "HEAD", "--"}, paths...)
		fallback, fallbackErr := execx.Run(ctx, repoRoot, nil, "git", fallbackArgs...)
		if fallbackErr != nil {
			return fallback, fmt.Errorf("unstage paths: %w\n%s", fallbackErr, fallback.Output())
		}
		return fallback, nil
	}

	args := append([]string{"add", "--"}, paths...)
	result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
	if err != nil {
		return result, fmt.Errorf("stage paths: %w\n%s", err, result.Output())
	}
	return result, nil
}

func LoadDiff(ctx context.Context, repoRoot string, cfg config.GitConfig, change FileChange) (string, string, error) {
	mode := ChooseDiffMode(change, cfg.DiffMode)
	if mode == "" {
		return "", "", nil
	}

	template := cfg.UnstagedDiffCommand
	if mode == DiffModeStaged {
		template = cfg.StagedDiffCommand
	}

	expanded := execx.ExpandArgs(template, map[string]string{
		"{file}":      change.Path,
		"{repo_root}": repoRoot,
	})
	if len(expanded) == 0 {
		return "", mode, fmt.Errorf("diff command template is empty")
	}
	name, args := expanded[0], expanded[1:]

	result, err := execx.Run(ctx, repoRoot, nil, name, args...)
	if err != nil {
		return "", mode, fmt.Errorf("load diff: %w\n%s", err, result.Output())
	}
	return result.Stdout, mode, nil
}

func LoadDirectoryDiff(ctx context.Context, repoRoot string, dir string) (string, string, error) {
	pathspec := dir
	if strings.TrimSpace(pathspec) == "" {
		pathspec = "."
	}

	unstagedArgs := []string{"diff", "--no-ext-diff", "--", pathspec}
	stagedArgs := []string{"diff", "--cached", "--no-ext-diff", "--", pathspec}

	staged, err := execx.Run(ctx, repoRoot, nil, "git", stagedArgs...)
	if err != nil {
		return "", "", fmt.Errorf("load staged directory diff: %w\n%s", err, staged.Output())
	}
	unstaged, err := execx.Run(ctx, repoRoot, nil, "git", unstagedArgs...)
	if err != nil {
		return "", "", fmt.Errorf("load unstaged directory diff: %w\n%s", err, unstaged.Output())
	}

	sections := make([]string, 0, 2)
	mode := ""
	if strings.TrimSpace(staged.Stdout) != "" {
		sections = append(sections, "Staged changes\n\n"+staged.Stdout)
		mode = DiffModeStaged
	}
	if strings.TrimSpace(unstaged.Stdout) != "" {
		sections = append(sections, "Unstaged changes\n\n"+unstaged.Stdout)
		if mode == DiffModeStaged {
			mode = "combined"
		} else {
			mode = DiffModeUnstaged
		}
	}
	if len(sections) == 0 {
		return "", "directory", nil
	}
	if mode == "" {
		mode = "directory"
	}
	return strings.Join(sections, "\n\n"), mode, nil
}

func LoadStagedRepositoryDiff(ctx context.Context, repoRoot string) (string, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "diff", "--cached", "--no-ext-diff", "--", ".")
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w\n%s", err, result.Output())
	}
	return result.Stdout, nil
}

func LoadHistory(ctx context.Context, repoRoot string, maxCommits int) ([]CommitHistoryEntry, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "log", "--date-order", fmt.Sprintf("-%d", maxCommits), "--pretty=format:%h%x09%an%x09%s")
	if err != nil {
		return nil, fmt.Errorf("git log: %w\n%s", err, result.Output())
	}

	var entries []CommitHistoryEntry
	for _, line := range strings.Split(result.Stdout, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		entries = append(entries, CommitHistoryEntry{
			Hash:    parts[0],
			Author:  parts[1],
			Subject: parts[2],
		})
	}
	return entries, nil
}

func LoadCommitPatch(ctx context.Context, repoRoot string, hash string) (string, error) {
	if !hexSHAPattern.MatchString(hash) {
		return "", fmt.Errorf("invalid commit hash: %q", hash)
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "show", "--stat", "--patch", "--format=fuller", hash)
	if err != nil {
		return "", fmt.Errorf("git show %s: %w\n%s", hash, err, result.Output())
	}
	return result.Stdout, nil
}

func Push(ctx context.Context, repoRoot string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "push")
	if err != nil {
		return result, fmt.Errorf("git push: %w\n%s", err, result.Output())
	}
	return result, nil
}

func Pull(ctx context.Context, repoRoot string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "pull")
	if err != nil {
		return result, fmt.Errorf("git pull: %w\n%s", err, result.Output())
	}
	return result, nil
}

func BuildCommitArgs(title string, body string, noVerify bool) []string {
	args := []string{"commit", "-m", title}
	if strings.TrimSpace(body) != "" {
		args = append(args, "-m", body)
	}
	if noVerify {
		args = append(args, "--no-verify")
	}
	return args
}

func Commit(ctx context.Context, repoRoot string, title string, body string, noVerify bool) (execx.Result, error) {
	args := BuildCommitArgs(title, body, noVerify)
	result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
	if err != nil {
		return result, fmt.Errorf("git commit: %w\n%s", err, result.Output())
	}
	return result, nil
}

func ChooseDiffMode(change FileChange, configured string) string {
	switch configured {
	case DiffModeStaged:
		if change.HasStaged() {
			return DiffModeStaged
		}
		if change.HasUnstaged() {
			return DiffModeUnstaged
		}
	case DiffModeUnstaged:
		if change.HasUnstaged() {
			return DiffModeUnstaged
		}
		if change.HasStaged() {
			return DiffModeStaged
		}
	default:
		if change.HasStaged() {
			return DiffModeStaged
		}
		if change.HasUnstaged() {
			return DiffModeUnstaged
		}
	}
	return ""
}

func AheadBehindLabel(status RepoStatus) string {
	var segments []string
	if status.Ahead > 0 {
		segments = append(segments, "↑"+strconv.Itoa(status.Ahead))
	}
	if status.Behind > 0 {
		segments = append(segments, "↓"+strconv.Itoa(status.Behind))
	}
	return strings.Join(segments, " ")
}

func Stash(ctx context.Context, repoRoot string, name string) (execx.Result, error) {
	var result execx.Result
	var err error
	if name != "" {
		result, err = execx.Run(ctx, repoRoot, nil, "git", "stash", "push", "-m", name)
	} else {
		result, err = execx.Run(ctx, repoRoot, nil, "git", "stash", "push")
	}
	if err != nil {
		return result, fmt.Errorf("git stash: %w\n%s", err, result.Output())
	}
	return result, nil
}

func Discard(ctx context.Context, repoRoot string, change FileChange) (execx.Result, error) {
	var out execx.Result

	if change.HasStaged() {
		r, err := execx.Run(ctx, repoRoot, nil, "git", "restore", "--staged", "--", change.Path)
		if err != nil {
			// fallback
			r, err = execx.Run(ctx, repoRoot, nil, "git", "reset", "HEAD", "--", change.Path)
			if err != nil {
				return r, fmt.Errorf("unstage for discard: %w\n%s", err, r.Output())
			}
		}
		out = r
	}

	if change.HasUnstaged() || change.IsUntracked {
		if change.IsUntracked {
			r, err := execx.Run(ctx, repoRoot, nil, "git", "clean", "-f", "--", change.Path)
			if err != nil {
				return r, fmt.Errorf("clean untracked: %w\n%s", err, r.Output())
			}
			out = r
		} else {
			r, err := execx.Run(ctx, repoRoot, nil, "git", "restore", "--", change.Path)
			if err != nil {
				return r, fmt.Errorf("restore file: %w\n%s", err, r.Output())
			}
			out = r
		}
	}

	return out, nil
}

func Reset(ctx context.Context, repoRoot string, change FileChange) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "restore", "--staged", "--", change.Path)
	if err != nil {
		result, err = execx.Run(ctx, repoRoot, nil, "git", "reset", "HEAD", "--", change.Path)
		if err != nil {
			return result, fmt.Errorf("git reset: %w\n%s", err, result.Output())
		}
	}
	return result, nil
}

func Amend(ctx context.Context, repoRoot string, title string, body string, noVerify bool) (execx.Result, error) {
	args := []string{"commit", "--amend", "-m", title}
	if strings.TrimSpace(body) != "" {
		args = append(args, "-m", body)
	}
	if noVerify {
		args = append(args, "--no-verify")
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
	if err != nil {
		return result, fmt.Errorf("git commit --amend: %w\n%s", err, result.Output())
	}
	return result, nil
}

func Fetch(ctx context.Context, repoRoot string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "fetch")
	if err != nil {
		return result, fmt.Errorf("git fetch: %w\n%s", err, result.Output())
	}
	return result, nil
}

func StageAll(ctx context.Context, repoRoot string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "add", ".")
	if err != nil {
		return result, fmt.Errorf("git add .: %w\n%s", err, result.Output())
	}
	return result, nil
}

func UnstageAll(ctx context.Context, repoRoot string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "restore", "--staged", ".")
	if err != nil {
		result, err = execx.Run(ctx, repoRoot, nil, "git", "reset", "HEAD")
		if err != nil {
			return result, fmt.Errorf("git restore --staged: %w\n%s", err, result.Output())
		}
	}
	return result, nil
}

func Ignore(repoRoot string, path string) error {
	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, path)
	return err
}

func LoadRawFile(repoRoot string, path string) (string, error) {
	fullPath := filepath.Join(repoRoot, path)
	cleanRoot := filepath.Clean(repoRoot)
	if !strings.HasPrefix(filepath.Clean(fullPath), cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside repository root", path)
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

func CreateBranch(ctx context.Context, repoRoot, name, base string) (execx.Result, error) {
	if !validBranchName.MatchString(name) {
		return execx.Result{}, fmt.Errorf("invalid branch name: %q", name)
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "checkout", "-b", name, base)
	if err != nil {
		return result, fmt.Errorf("git checkout -b: %w\n%s", err, result.Output())
	}
	return result, nil
}

func uniqueNonEmptyPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		filtered = append(filtered, path)
	}
	return filtered
}
