package github

import (
	"testing"
)

func TestParsePullRequestListEmpty(t *testing.T) {
	got, err := ParsePullRequestList([]byte("[]"))
	if err != nil {
		t.Fatalf("ParsePullRequestList([]) error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(got))
	}
}

func TestParsePullRequestListSingle(t *testing.T) {
	in := []byte(`[
		{
			"number": 42,
			"title": "feat: add foo",
			"author": {"login": "alice"},
			"headRefName": "feature-x",
			"baseRefName": "main",
			"url": "https://github.com/owner/repo/pull/42",
			"isDraft": false
		}
	]`)

	got, err := ParsePullRequestList(in)
	if err != nil {
		t.Fatalf("ParsePullRequestList error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(got))
	}
	pr := got[0]
	if pr.Number != 42 {
		t.Fatalf("Number = %d, want 42", pr.Number)
	}
	if pr.Title != "feat: add foo" {
		t.Fatalf("Title = %q, want %q", pr.Title, "feat: add foo")
	}
	if pr.Author != "alice" {
		t.Fatalf("Author = %q, want %q", pr.Author, "alice")
	}
	if pr.HeadRefName != "feature-x" {
		t.Fatalf("HeadRefName = %q, want %q", pr.HeadRefName, "feature-x")
	}
	if pr.BaseRefName != "main" {
		t.Fatalf("BaseRefName = %q, want %q", pr.BaseRefName, "main")
	}
	if pr.URL != "https://github.com/owner/repo/pull/42" {
		t.Fatalf("URL = %q, want %q", pr.URL, "https://github.com/owner/repo/pull/42")
	}
	if pr.IsDraft {
		t.Fatalf("IsDraft = true, want false")
	}
}

func TestParsePullRequestListMultipleAndDraft(t *testing.T) {
	in := []byte(`[
		{
			"number": 1,
			"title": "first",
			"author": {"login": "bob"},
			"headRefName": "a",
			"baseRefName": "main",
			"url": "https://github.com/o/r/pull/1",
			"isDraft": true
		},
		{
			"number": 2,
			"title": "second",
			"author": {"login": "carol"},
			"headRefName": "b",
			"baseRefName": "main",
			"url": "https://github.com/o/r/pull/2",
			"isDraft": false
		}
	]`)

	got, err := ParsePullRequestList(in)
	if err != nil {
		t.Fatalf("ParsePullRequestList error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(got))
	}
	if !got[0].IsDraft {
		t.Fatalf("expected first PR to be draft")
	}
	if got[1].IsDraft {
		t.Fatalf("expected second PR not to be draft")
	}
	if got[0].Author != "bob" {
		t.Fatalf("Author[0] = %q, want bob", got[0].Author)
	}
	if got[1].Number != 2 {
		t.Fatalf("Number[1] = %d, want 2", got[1].Number)
	}
}

func TestParsePullRequestListMissingAuthor(t *testing.T) {
	in := []byte(`[
		{
			"number": 7,
			"title": "no author",
			"headRefName": "x",
			"baseRefName": "main",
			"url": "https://github.com/o/r/pull/7",
			"isDraft": false
		}
	]`)

	got, err := ParsePullRequestList(in)
	if err != nil {
		t.Fatalf("ParsePullRequestList error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(got))
	}
	if got[0].Author != "" {
		t.Fatalf("Author = %q, want empty", got[0].Author)
	}
}

func TestParsePullRequestListInvalidJSON(t *testing.T) {
	_, err := ParsePullRequestList([]byte("not json"))
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestParsePullRequestListEmptyInput(t *testing.T) {
	got, err := ParsePullRequestList([]byte(""))
	if err != nil {
		t.Fatalf("ParsePullRequestList(empty) error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice for empty input, got %d", len(got))
	}
}

func TestBuildListOpenPRArgs(t *testing.T) {
	args := BuildListOpenPRArgs(50)
	want := []string{"pr", "list", "--state", "open", "--limit", "50", "--json", "number,title,author,headRefName,baseRefName,url,isDraft"}
	if len(args) != len(want) {
		t.Fatalf("args length = %d, want %d (%v)", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildListOpenPRArgsDefaultLimitWhenZero(t *testing.T) {
	args := BuildListOpenPRArgs(0)
	// expect default 30 (matches gh default behavior conceptually).
	foundLimit := false
	for i, a := range args {
		if a == "--limit" && i+1 < len(args) && args[i+1] == "30" {
			foundLimit = true
			break
		}
	}
	if !foundLimit {
		t.Fatalf("expected default limit of 30 in args: %v", args)
	}
}
