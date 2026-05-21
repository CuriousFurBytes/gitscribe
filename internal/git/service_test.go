package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mustGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	env := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func mustCommit(t *testing.T, dir, file, msg string) {
	t.Helper()
	env := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if err := os.WriteFile(filepath.Join(dir, file), []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", file},
		{"commit", "-m", msg},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func TestStash(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Stash(context.Background(), dir, "")
	if err != nil {
		t.Fatalf("Stash failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "file.txt"))
	if string(data) != "content\n" {
		t.Fatalf("expected file restored to stashed state, got %q", string(data))
	}
}

func TestStashWithName(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Stash(context.Background(), dir, "my-feature")
	if err != nil {
		t.Fatalf("Stash with name failed: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "stash", "list").Output()
	if err != nil {
		t.Fatalf("git stash list: %v", err)
	}
	if !strings.Contains(string(out), "my-feature") {
		t.Fatalf("expected stash to have name 'my-feature', got: %s", out)
	}
}

func TestDiscard(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	change := FileChange{Path: "file.txt", UnstagedStatus: StatusModified}
	_, err := Discard(context.Background(), dir, change)
	if err != nil {
		t.Fatalf("Discard failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "file.txt"))
	if string(data) != "content\n" {
		t.Fatalf("expected original content, got %q", string(data))
	}
}

func TestReset(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "add", "file.txt").Run()

	change := FileChange{Path: "file.txt", StagedStatus: StatusModified}
	_, err := Reset(context.Background(), dir, change)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestAmend(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "original")

	_, err := Amend(context.Background(), dir, "amended title", "", false)
	if err != nil {
		t.Fatalf("Amend failed: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "log", "--oneline", "-1").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "amended title") {
		t.Fatalf("expected amended title in log, got %q", string(out))
	}
}

func TestFetch(t *testing.T) {
	dir := mustGitRepo(t)
	// fetch with no remote should fail gracefully — returns error but not panic
	_, err := Fetch(context.Background(), dir)
	// err expected (no remote), just check it doesn't panic
	_ = err
}

func TestStageAll(t *testing.T) {
	dir := mustGitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := StageAll(context.Background(), dir)
	if err != nil {
		t.Fatalf("StageAll failed: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "A  new.txt") {
		t.Fatalf("expected new.txt staged, got %q", string(out))
	}
}

func TestUnstageAll(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	cmd.Env = env
	_ = cmd.Run()

	_, err := UnstageAll(context.Background(), dir)
	if err != nil {
		t.Fatalf("UnstageAll failed: %v", err)
	}
}

func TestIgnore(t *testing.T) {
	dir := mustGitRepo(t)

	if err := Ignore(dir, "secret.txt"); err != nil {
		t.Fatalf("Ignore failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to exist: %v", err)
	}
	if !strings.Contains(string(data), "secret.txt") {
		t.Fatalf("expected secret.txt in .gitignore, got %q", string(data))
	}
}

func TestIgnoreAppendsToExisting(t *testing.T) {
	dir := mustGitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Ignore(dir, "secret.txt"); err != nil {
		t.Fatalf("Ignore failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !strings.Contains(string(data), "*.log") || !strings.Contains(string(data), "secret.txt") {
		t.Fatalf("expected both entries, got %q", string(data))
	}
}

func TestLoadRawFile(t *testing.T) {
	dir := mustGitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := LoadRawFile(dir, "hello.txt")
	if err != nil {
		t.Fatalf("LoadRawFile failed: %v", err)
	}
	if content != "hello\n" {
		t.Fatalf("content = %q, want %q", content, "hello\n")
	}
}

func TestLoadRawFileMissing(t *testing.T) {
	dir := mustGitRepo(t)
	_, err := LoadRawFile(dir, "does-not-exist.txt")
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestCreateBranchRejectsInvalidName(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	_, err := CreateBranch(context.Background(), dir, "-f", "main")
	if err == nil {
		t.Fatalf("expected error for invalid branch name -f, got nil")
	}
	if !strings.Contains(err.Error(), "invalid branch name") {
		t.Fatalf("expected 'invalid branch name' error, got %v", err)
	}
}

func TestLoadCommitPatchRejectsInvalidHash(t *testing.T) {
	dir := mustGitRepo(t)
	_, err := LoadCommitPatch(context.Background(), dir, "--exec=bad")
	if err == nil {
		t.Fatalf("expected error for invalid commit hash, got nil")
	}
	if !strings.Contains(err.Error(), "invalid commit hash") {
		t.Fatalf("expected 'invalid commit hash' error, got %v", err)
	}
}

func TestLoadRawFileRejectsPathTraversal(t *testing.T) {
	dir := mustGitRepo(t)
	outside := filepath.Join(dir, "../outside.txt")
	_ = os.WriteFile(filepath.Clean(outside), []byte("secret"), 0o644)
	_, err := LoadRawFile(dir, "../outside.txt")
	if err == nil {
		t.Fatalf("expected error for path traversal outside repo root, got nil")
	}
	if !strings.Contains(err.Error(), "outside") {
		t.Fatalf("expected 'outside repository root' error, got %v", err)
	}
}
