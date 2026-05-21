package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunPreCommitHookErrorsWhenMissing(t *testing.T) {
	repoRoot := t.TempDir()

	_, err := RunPreCommitHook(context.Background(), repoRoot, filepath.Join(repoRoot, ".git"))
	if err == nil {
		t.Fatalf("expected missing hook to fail")
	}
}

func TestRunPreCommitHookErrorsWhenNotExecutable(t *testing.T) {
	repoRoot := t.TempDir()
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
		t.Fatalf("mkdir hook dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho hi\n"), 0o644); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	_, err := RunPreCommitHook(context.Background(), repoRoot, filepath.Join(repoRoot, ".git"))
	if err == nil {
		t.Fatalf("expected non-executable hook to fail")
	}
}

func TestRunPreCommitHookExecutesHook(t *testing.T) {
	repoRoot := t.TempDir()
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
		t.Fatalf("mkdir hook dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho hook-ok\n"), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	result, err := RunPreCommitHook(context.Background(), repoRoot, filepath.Join(repoRoot, ".git"))
	if err != nil {
		t.Fatalf("RunPreCommitHook() error = %v", err)
	}
	if result.Stdout != "hook-ok\n" {
		t.Fatalf("stdout = %q, want %q", result.Stdout, "hook-ok\n")
	}
}
