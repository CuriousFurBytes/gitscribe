package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPRTemplatePriority(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".github"), 0o755); err != nil {
		t.Fatalf("mkdir .github: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, ".github", "pull_request_template.md"), []byte("primary"), 0o644); err != nil {
		t.Fatalf("write primary template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "docs", "pull_request_template.md"), []byte("secondary"), 0o644); err != nil {
		t.Fatalf("write secondary template: %v", err)
	}

	path, body, err := DiscoverPRTemplate(repoRoot, "")
	if err != nil {
		t.Fatalf("DiscoverPRTemplate() error = %v", err)
	}
	if filepath.Base(path) != "pull_request_template.md" || body != "primary" {
		t.Fatalf("expected .github template to win, got path=%q body=%q", path, body)
	}
}

func TestDiscoverPRTemplateUsesCustomFallback(t *testing.T) {
	repoRoot := t.TempDir()
	custom := filepath.Join(repoRoot, "custom-pr.md")
	if err := os.WriteFile(custom, []byte("custom"), 0o644); err != nil {
		t.Fatalf("write custom template: %v", err)
	}

	path, body, err := DiscoverPRTemplate(repoRoot, custom)
	if err != nil {
		t.Fatalf("DiscoverPRTemplate() error = %v", err)
	}
	if path != custom || body != "custom" {
		t.Fatalf("expected custom template fallback, got path=%q body=%q", path, body)
	}
}

func TestDiscoverPRTemplateRejectsPathTraversal(t *testing.T) {
	repoRoot := t.TempDir()
	_, _, err := DiscoverPRTemplate(repoRoot, "../../../etc/passwd")
	if err == nil {
		t.Fatalf("expected error for path traversal outside repo root, got nil")
	}
}
