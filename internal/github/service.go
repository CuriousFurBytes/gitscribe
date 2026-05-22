package github

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

var prURLPattern = regexp.MustCompile(`https?://[^\s]+/pull/\d+`)

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

// ExtractPRURL parses gh's output for the pull request URL it printed on
// success. Returns an empty string if no URL is found.
func ExtractPRURL(output string) string {
	match := prURLPattern.FindString(output)
	// Strip a single trailing punctuation char like '.' or ',' that
	// might be picked up if the URL is embedded in prose.
	for len(match) > 0 {
		last := match[len(match)-1]
		if last == '.' || last == ',' || last == ')' || last == ']' {
			match = match[:len(match)-1]
			continue
		}
		break
	}
	return match
}

// OpenInBrowser opens the given URL in the user's default browser.
func OpenInBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("no url to open")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
