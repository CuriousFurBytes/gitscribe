package git

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

// DiffStats holds aggregate counts from a `git diff --shortstat` output.
type DiffStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

// IsZero reports whether the stats have no recorded changes.
func (s DiffStats) IsZero() bool {
	return s.FilesChanged == 0 && s.Insertions == 0 && s.Deletions == 0
}

// Summary renders a compact, human-friendly representation such as
// "3 files, +42 -7" suitable for display in the commit screen header.
func (s DiffStats) Summary() string {
	if s.IsZero() {
		return ""
	}
	noun := "files"
	if s.FilesChanged == 1 {
		noun = "file"
	}
	return fmt.Sprintf("%d %s, +%d -%d", s.FilesChanged, noun, s.Insertions, s.Deletions)
}

var (
	shortstatFilesPattern      = regexp.MustCompile(`(\d+)\s+files?\s+changed`)
	shortstatInsertionsPattern = regexp.MustCompile(`(\d+)\s+insertions?\(\+\)`)
	shortstatDeletionsPattern  = regexp.MustCompile(`(\d+)\s+deletions?\(-\)`)
)

// ParseDiffStats parses the output of `git diff --shortstat` (or
// `git diff --cached --shortstat`) into a DiffStats struct. Empty input
// returns a zero-value DiffStats with no error.
func ParseDiffStats(output string) (DiffStats, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return DiffStats{}, nil
	}

	stats := DiffStats{}
	if m := shortstatFilesPattern.FindStringSubmatch(trimmed); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return DiffStats{}, fmt.Errorf("parse files changed: %w", err)
		}
		stats.FilesChanged = n
	}
	if m := shortstatInsertionsPattern.FindStringSubmatch(trimmed); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return DiffStats{}, fmt.Errorf("parse insertions: %w", err)
		}
		stats.Insertions = n
	}
	if m := shortstatDeletionsPattern.FindStringSubmatch(trimmed); len(m) == 2 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return DiffStats{}, fmt.Errorf("parse deletions: %w", err)
		}
		stats.Deletions = n
	}
	return stats, nil
}

// LoadStagedDiffStats runs `git diff --cached --shortstat` against the
// given repository and parses the result into a DiffStats value.
func LoadStagedDiffStats(ctx context.Context, repoRoot string) (DiffStats, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "diff", "--cached", "--shortstat")
	if err != nil {
		return DiffStats{}, fmt.Errorf("git diff --cached --shortstat: %w\n%s", err, result.Output())
	}
	return ParseDiffStats(result.Stdout)
}
