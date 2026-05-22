package github

import (
	"testing"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestBuildCreatePRArgs(t *testing.T) {
	cfg := config.PullRequestConfig{
		DefaultBase: "main",
		GHArgs:      []string{"--assignee", "@me"},
	}

	got := BuildCreatePRArgs(cfg, "feat: add history", "body")
	want := []string{"pr", "create", "--base", "main", "--title", "feat: add history", "--body", "body", "--assignee", "@me"}
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExtractPRURLFindsURLInGhOutput(t *testing.T) {
	output := "Creating pull request for feature into main in owner/repo\n\nhttps://github.com/owner/repo/pull/42\n"
	got := ExtractPRURL(output)
	want := "https://github.com/owner/repo/pull/42"
	if got != want {
		t.Fatalf("ExtractPRURL = %q, want %q", got, want)
	}
}

func TestExtractPRURLFindsURLWithTrailingPunctuation(t *testing.T) {
	output := "Created: https://github.com/owner/repo/pull/123."
	got := ExtractPRURL(output)
	want := "https://github.com/owner/repo/pull/123"
	if got != want {
		t.Fatalf("ExtractPRURL = %q, want %q", got, want)
	}
}

func TestExtractPRURLReturnsEmptyWhenAbsent(t *testing.T) {
	if got := ExtractPRURL("no url here"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestExtractPRURLIgnoresOtherGithubURLs(t *testing.T) {
	output := "See https://github.com/owner/repo for repo. https://github.com/owner/repo/pull/7 is the PR."
	got := ExtractPRURL(output)
	want := "https://github.com/owner/repo/pull/7"
	if got != want {
		t.Fatalf("ExtractPRURL = %q, want %q", got, want)
	}
}
