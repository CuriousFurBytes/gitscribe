package github

import (
	"context"
	"fmt"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

func BuildCreatePRArgs(cfg config.PullRequestConfig, title string, body string) []string {
	args := []string{"pr", "create"}
	if cfg.DefaultBase != "" {
		args = append(args, "--base", cfg.DefaultBase)
	}
	args = append(args, "--title", title, "--body", body)
	args = append(args, cfg.GHArgs...)
	return args
}

func CreatePR(ctx context.Context, repoRoot string, cfg config.PullRequestConfig, title string, body string) (execx.Result, error) {
	args := BuildCreatePRArgs(cfg, title, body)
	result, err := execx.Run(ctx, repoRoot, nil, "gh", args...)
	if err != nil {
		return result, fmt.Errorf("gh pr create: %w\n%s", err, result.Output())
	}
	return result, nil
}
