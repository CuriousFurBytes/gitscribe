package repo

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

var ErrNotGitRepo = errors.New("current directory is not inside a Git repository")

type Info struct {
	Root   string
	GitDir string
	Branch string
}

func Detect(ctx context.Context, dir string) (Info, error) {
	rootResult, err := execx.Run(ctx, dir, nil, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return Info{}, ErrNotGitRepo
	}

	root := strings.TrimSpace(rootResult.Stdout)
	if root == "" {
		return Info{}, ErrNotGitRepo
	}

	gitDirResult, err := execx.Run(ctx, dir, nil, "git", "rev-parse", "--git-dir")
	if err != nil {
		return Info{}, fmt.Errorf("resolve git dir: %w", err)
	}

	branchResult, err := execx.Run(ctx, dir, nil, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return Info{}, fmt.Errorf("resolve branch: %w", err)
	}

	gitDir := strings.TrimSpace(gitDirResult.Stdout)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}

	return Info{
		Root:   root,
		GitDir: gitDir,
		Branch: strings.TrimSpace(branchResult.Stdout),
	}, nil
}
