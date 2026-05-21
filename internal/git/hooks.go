package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

func RunPreCommitHook(ctx context.Context, repoRoot string, gitDir string) (execx.Result, error) {
	if gitDir == "" {
		gitDir = filepath.Join(repoRoot, ".git")
	}
	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		return execx.Result{}, fmt.Errorf("pre-commit hook not found at %s", hookPath)
	}
	if info.Mode()&0o111 == 0 {
		return execx.Result{}, fmt.Errorf("pre-commit hook is not executable: %s", hookPath)
	}

	result, err := execx.Run(ctx, repoRoot, nil, hookPath)
	if err != nil {
		return result, fmt.Errorf("pre-commit hook: %w\n%s", err, result.Output())
	}
	return result, nil
}
