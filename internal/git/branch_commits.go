package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

// DefaultBranchCommitsLimit caps the number of commits listed when
// summarising the commits that will be included in a pull request.
const DefaultBranchCommitsLimit = 20

// CommitSummary holds the abbreviated SHA and subject for a single commit
// in the list of commits between a branch and its base.
type CommitSummary struct {
	ShortSHA string
	Subject  string
}

// ParseBranchCommits parses tab-separated `<short-sha>\t<subject>` lines
// (one per commit, as emitted by
// `git log --pretty=format:%h%x09%s <base>..HEAD`) into a slice of
// CommitSummary values. Empty or whitespace-only input returns nil.
// Malformed lines (no tab separator) are skipped.
func ParseBranchCommits(output string) []CommitSummary {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}
	var entries []CommitSummary
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		sha := strings.TrimSpace(parts[0])
		if sha == "" {
			continue
		}
		entries = append(entries, CommitSummary{ShortSHA: sha, Subject: parts[1]})
	}
	return entries
}

// DetectDefaultBaseBranch attempts to identify the remote default branch
// by inspecting `refs/remotes/origin/HEAD`. It returns the short branch
// name (e.g. "main"). If detection fails it falls back to "main".
func DetectDefaultBaseBranch(ctx context.Context, repoRoot string) string {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(result.Stdout)
	// Expected output: "origin/main"; strip the remote prefix if present.
	if idx := strings.Index(ref, "/"); idx >= 0 {
		ref = ref[idx+1:]
	}
	if ref == "" {
		return "main"
	}
	return ref
}

// LoadBranchCommits returns the commits reachable from HEAD that are not
// reachable from origin/<base> — i.e. the commits a pull request would
// contain. The list is capped at limit entries (use
// DefaultBranchCommitsLimit for the standard cap).
func LoadBranchCommits(ctx context.Context, repoRoot string, limit int) ([]CommitSummary, error) {
	if limit <= 0 {
		limit = DefaultBranchCommitsLimit
	}
	base := DetectDefaultBaseBranch(ctx, repoRoot)
	rev := fmt.Sprintf("origin/%s..HEAD", base)
	args := []string{
		"log",
		fmt.Sprintf("-%d", limit),
		"--pretty=format:%h%x09%s",
		rev,
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w\n%s", rev, err, result.Output())
	}
	return ParseBranchCommits(result.Stdout), nil
}
