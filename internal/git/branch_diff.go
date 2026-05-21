package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

// DefaultBranchSeriesDiffBytes caps the size of the diff loaded for the
// current branch series view. Output larger than this is truncated with a
// notice appended at the end.
const DefaultBranchSeriesDiffBytes = 1 << 20 // 1 MiB

const diffTruncatedNotice = "\n\n... diff truncated (output exceeded cap) ..."

// BuildBranchSeriesDiffArgs returns the argv (excluding the leading "git")
// used to load the diff of the current branch series against origin/<base>.
// It uses the three-dot range (diff from the merge base) so the output
// reflects the changes introduced by the branch itself rather than diffs
// caused by base-branch updates. An empty or whitespace-only base falls
// back to "main".
func BuildBranchSeriesDiffArgs(base string) []string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "main"
	}
	return []string{"diff", "--no-ext-diff", fmt.Sprintf("origin/%s...HEAD", base)}
}

// CapDiffOutput returns content unchanged when it fits within maxBytes; if
// it exceeds the cap, it returns the first maxBytes bytes followed by a
// short truncation notice. A non-positive maxBytes disables the cap.
func CapDiffOutput(content string, maxBytes int) string {
	if maxBytes <= 0 || len(content) <= maxBytes {
		return content
	}
	return content[:maxBytes] + diffTruncatedNotice
}

// LoadBranchSeriesDiff returns the diff of the commits on the current
// branch against the origin/<base> branch (three-dot range), using the
// base detected by DetectDefaultBaseBranch. Output is capped at
// DefaultBranchSeriesDiffBytes.
func LoadBranchSeriesDiff(ctx context.Context, repoRoot string) (string, error) {
	base := DetectDefaultBaseBranch(ctx, repoRoot)
	args := BuildBranchSeriesDiffArgs(base)
	result, err := execx.Run(ctx, repoRoot, nil, "git", args...)
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, result.Output())
	}
	return CapDiffOutput(result.Stdout, DefaultBranchSeriesDiffBytes), nil
}
