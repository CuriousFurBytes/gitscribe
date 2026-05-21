package git

import (
	"strings"
	"testing"
)

func TestBuildBranchSeriesDiffArgs(t *testing.T) {
	t.Run("uses three-dot range against origin base", func(t *testing.T) {
		got := BuildBranchSeriesDiffArgs("main")
		want := []string{"diff", "--no-ext-diff", "origin/main...HEAD"}
		if len(got) != len(want) {
			t.Fatalf("expected %d args, got %d (%v)", len(want), len(got), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("arg %d: expected %q, got %q", i, want[i], got[i])
			}
		}
	})

	t.Run("trims whitespace on base name", func(t *testing.T) {
		got := BuildBranchSeriesDiffArgs("  develop  ")
		if got[2] != "origin/develop...HEAD" {
			t.Fatalf("expected origin/develop...HEAD, got %q", got[2])
		}
	})

	t.Run("falls back to main on empty base", func(t *testing.T) {
		got := BuildBranchSeriesDiffArgs("")
		if got[2] != "origin/main...HEAD" {
			t.Fatalf("expected fallback to origin/main...HEAD, got %q", got[2])
		}
	})
}

func TestCapDiffOutput(t *testing.T) {
	t.Run("returns input unchanged when under cap", func(t *testing.T) {
		in := "small diff content"
		got := CapDiffOutput(in, 1024)
		if got != in {
			t.Fatalf("expected unchanged output, got %q", got)
		}
	})

	t.Run("truncates and appends notice when over cap", func(t *testing.T) {
		in := strings.Repeat("a", 200)
		got := CapDiffOutput(in, 50)
		if len(got) > 50+len(diffTruncatedNotice)+1 {
			t.Fatalf("expected truncated output near cap, got len=%d", len(got))
		}
		if !strings.Contains(got, diffTruncatedNotice) {
			t.Fatalf("expected truncation notice in output: %q", got)
		}
		if !strings.HasPrefix(got, strings.Repeat("a", 50)) {
			t.Fatalf("expected prefix to be first 50 chars of input, got %q", got[:min(60, len(got))])
		}
	})

	t.Run("non-positive cap returns input unchanged", func(t *testing.T) {
		in := "anything"
		if got := CapDiffOutput(in, 0); got != in {
			t.Fatalf("expected passthrough on zero cap, got %q", got)
		}
		if got := CapDiffOutput(in, -1); got != in {
			t.Fatalf("expected passthrough on negative cap, got %q", got)
		}
	})
}
