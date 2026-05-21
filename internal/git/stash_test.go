package git

import (
	"context"
	"strings"
	"testing"
)

func TestParseStashesEmpty(t *testing.T) {
	entries := parseStashes("")
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %d", len(entries))
	}
}

func TestParseStashesBasic(t *testing.T) {
	input := "stash@{0}\tOn main: my-feature\nstash@{1}\tWIP on feat: quick save\n"
	entries := parseStashes(input)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].RefName != "stash@{0}" {
		t.Fatalf("entries[0].RefName = %q, want stash@{0}", entries[0].RefName)
	}
	if entries[0].Index != 0 {
		t.Fatalf("entries[0].Index = %d, want 0", entries[0].Index)
	}
	if entries[0].Message != "On main: my-feature" {
		t.Fatalf("entries[0].Message = %q, want 'On main: my-feature'", entries[0].Message)
	}
	if entries[1].Index != 1 {
		t.Fatalf("entries[1].Index = %d, want 1", entries[1].Index)
	}
}

func TestParseStashesPreservesMessageWithTabs(t *testing.T) {
	input := "stash@{0}\tWIP: msg with\ttabs\n"
	entries := parseStashes(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "WIP: msg with\ttabs" {
		t.Fatalf("message = %q", entries[0].Message)
	}
}

func TestShowStashRejectsInvalidRef(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	_, err := ShowStash(context.Background(), dir, "--index")
	if err == nil {
		t.Fatalf("expected error for invalid stash ref --index, got nil")
	}
	if !strings.Contains(err.Error(), "invalid stash ref") {
		t.Fatalf("expected 'invalid stash ref' error, got %v", err)
	}
}

func TestApplyStashRejectsInvalidRef(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	_, err := ApplyStash(context.Background(), dir, "--index")
	if err == nil {
		t.Fatalf("expected error for invalid stash ref --index, got nil")
	}
	if !strings.Contains(err.Error(), "invalid stash ref") {
		t.Fatalf("expected 'invalid stash ref' error, got %v", err)
	}
}

func TestPopStashRejectsInvalidRef(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	_, err := PopStash(context.Background(), dir, "--index")
	if err == nil {
		t.Fatalf("expected error for invalid stash ref --index, got nil")
	}
	if !strings.Contains(err.Error(), "invalid stash ref") {
		t.Fatalf("expected 'invalid stash ref' error, got %v", err)
	}
}

func TestDropStashRejectsInvalidRef(t *testing.T) {
	dir := mustGitRepo(t)
	mustCommit(t, dir, "file.txt", "init")
	_, err := DropStash(context.Background(), dir, "--index")
	if err == nil {
		t.Fatalf("expected error for invalid stash ref --index, got nil")
	}
	if !strings.Contains(err.Error(), "invalid stash ref") {
		t.Fatalf("expected 'invalid stash ref' error, got %v", err)
	}
}
