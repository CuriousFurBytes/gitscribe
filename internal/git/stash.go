package git

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

var stashRefPattern = regexp.MustCompile(`^stash@\{\d+\}$`)

type StashEntry struct {
	Index   int
	RefName string
	Message string
}

func ListStashes(ctx context.Context, repoRoot string) ([]StashEntry, error) {
	result, err := execx.Run(ctx, repoRoot, nil, "git", "stash", "list", "--pretty=format:%gd\t%s")
	if err != nil {
		return nil, fmt.Errorf("git stash list: %w\n%s", err, result.Output())
	}
	return parseStashes(result.Stdout), nil
}

func parseStashes(output string) []StashEntry {
	var entries []StashEntry
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		refName := parts[0]
		message := parts[1]

		idx := 0
		if n := strings.TrimPrefix(refName, "stash@{"); n != refName {
			n = strings.TrimSuffix(n, "}")
			if i, err := strconv.Atoi(n); err == nil {
				idx = i
			}
		}

		entries = append(entries, StashEntry{
			Index:   idx,
			RefName: refName,
			Message: message,
		})
	}
	return entries
}

func validateStashRef(ref string) error {
	if !stashRefPattern.MatchString(ref) {
		return fmt.Errorf("invalid stash ref: %q", ref)
	}
	return nil
}

func ShowStash(ctx context.Context, repoRoot, ref string) (string, error) {
	if err := validateStashRef(ref); err != nil {
		return "", err
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "stash", "show", "-p", ref)
	if err != nil {
		return "", fmt.Errorf("git stash show: %w\n%s", err, result.Output())
	}
	return result.Stdout, nil
}

func ApplyStash(ctx context.Context, repoRoot, ref string) (execx.Result, error) {
	if err := validateStashRef(ref); err != nil {
		return execx.Result{}, err
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "stash", "apply", ref)
	if err != nil {
		return result, fmt.Errorf("git stash apply: %w\n%s", err, result.Output())
	}
	return result, nil
}

func PopStash(ctx context.Context, repoRoot, ref string) (execx.Result, error) {
	if err := validateStashRef(ref); err != nil {
		return execx.Result{}, err
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "stash", "pop", ref)
	if err != nil {
		return result, fmt.Errorf("git stash pop: %w\n%s", err, result.Output())
	}
	return result, nil
}

func DropStash(ctx context.Context, repoRoot, ref string) (execx.Result, error) {
	if err := validateStashRef(ref); err != nil {
		return execx.Result{}, err
	}
	result, err := execx.Run(ctx, repoRoot, nil, "git", "stash", "drop", ref)
	if err != nil {
		return result, fmt.Errorf("git stash drop: %w\n%s", err, result.Output())
	}
	return result, nil
}
