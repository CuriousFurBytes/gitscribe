package git

import (
	"strings"
	"testing"
)

func TestParseWorktreesPorcelain(t *testing.T) {
	input := `worktree /home/user/repo
HEAD abc12345def67890
branch refs/heads/main

worktree /home/user/repo-feat
HEAD 1234567890abcdef
branch refs/heads/feature/new
`
	entries := parseWorktrees(input)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Path != "/home/user/repo" {
		t.Fatalf("entries[0].Path = %q, want /home/user/repo", entries[0].Path)
	}
	if entries[0].Branch != "main" {
		t.Fatalf("entries[0].Branch = %q, want main", entries[0].Branch)
	}
	if entries[0].Head != "abc12345" {
		t.Fatalf("entries[0].Head = %q, want abc12345", entries[0].Head)
	}
	if !entries[0].IsCurrent {
		t.Fatalf("entries[0].IsCurrent should be true (first worktree = main)")
	}

	if entries[1].Branch != "feature/new" {
		t.Fatalf("entries[1].Branch = %q, want feature/new", entries[1].Branch)
	}
	if entries[1].IsCurrent {
		t.Fatalf("entries[1].IsCurrent should be false")
	}
}

func TestParseWorktreesLockedFlag(t *testing.T) {
	input := `worktree /home/user/repo
HEAD abc12345
branch refs/heads/main
locked

`
	entries := parseWorktrees(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].IsLocked {
		t.Fatalf("expected IsLocked = true")
	}
}

func TestParseWorktreesEmpty(t *testing.T) {
	entries := parseWorktrees("")
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %d", len(entries))
	}
}

func TestParseWorktreesStripsRefsHeads(t *testing.T) {
	input := "worktree /tmp/x\nHEAD aabbccdd\nbranch refs/heads/my-branch\n\n"
	entries := parseWorktrees(input)
	if len(entries) == 0 {
		t.Fatalf("expected 1 entry")
	}
	if strings.Contains(entries[0].Branch, "refs/heads/") {
		t.Fatalf("branch should strip refs/heads/ prefix, got %q", entries[0].Branch)
	}
}
