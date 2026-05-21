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
