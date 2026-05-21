package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

type WorktreeEntry struct {
	Path      string
	Branch    string
	Head      string
	IsCurrent bool
	IsLocked  bool
}

func ListWorktrees(ctx context.Context, repoRoot string) ([]WorktreeEntry, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w\n%s", err, result.Output())
	}
	return parseWorktrees(result.Stdout), nil
}

func parseWorktrees(output string) []WorktreeEntry {
	var entries []WorktreeEntry
	var current WorktreeEntry
	first := true

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				if first {
					current.IsCurrent = true
					first = false
				}
				entries = append(entries, current)
				current = WorktreeEntry{}
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Head = strings.TrimPrefix(line, "HEAD ")
			if len(current.Head) > 8 {
				current.Head = current.Head[:8]
			}
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		} else if line == "locked" {
			current.IsLocked = true
		}
	}
	if current.Path != "" {
		if first {
			current.IsCurrent = true
		}
		entries = append(entries, current)
	}
	return entries
}

func CreateWorktree(ctx context.Context, repoRoot, path, branch string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "worktree", "add", path, "-b", branch)
	if err != nil {
		return result, fmt.Errorf("git worktree add: %w\n%s", err, result.Output())
	}
	return result, nil
}

func RemoveWorktree(ctx context.Context, repoRoot, path string) (execx.Result, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "worktree", "remove", path)
	if err != nil {
		return result, fmt.Errorf("git worktree remove: %w\n%s", err, result.Output())
	}
	return result, nil
}
