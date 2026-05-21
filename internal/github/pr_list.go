package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/CuriousFurBytes/gitscribe/internal/execx"
)

// defaultListLimit matches the default page size of `gh pr list` and is used
// when callers do not specify their own limit.
const defaultListLimit = 30

// PullRequestSummary is a single open pull request as returned by
// `gh pr list --json number,title,author,headRefName,baseRefName,url,isDraft`.
type PullRequestSummary struct {
	Number      int
	Title       string
	Author      string
	HeadRefName string
	BaseRefName string
	URL         string
	IsDraft     bool
}

// rawPullRequest mirrors the JSON shape emitted by `gh`. Using an intermediate
// type lets us flatten the nested `author.login` field into a plain string on
// PullRequestSummary while still tolerating missing fields.
type rawPullRequest struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Author      *struct {
		Login string `json:"login"`
	} `json:"author"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
	URL         string `json:"url"`
	IsDraft     bool   `json:"isDraft"`
}

// ParsePullRequestList decodes the output of `gh pr list --json ...` into a
// slice of PullRequestSummary. Empty input is treated as an empty list.
func ParsePullRequestList(data []byte) ([]PullRequestSummary, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return []PullRequestSummary{}, nil
	}
	var raw []rawPullRequest
	if err := json.Unmarshal(trimmed, &raw); err != nil {
		return nil, fmt.Errorf("decode pr list: %w", err)
	}
	out := make([]PullRequestSummary, 0, len(raw))
	for _, r := range raw {
		summary := PullRequestSummary{
			Number:      r.Number,
			Title:       r.Title,
			HeadRefName: r.HeadRefName,
			BaseRefName: r.BaseRefName,
			URL:         r.URL,
			IsDraft:     r.IsDraft,
		}
		if r.Author != nil {
			summary.Author = r.Author.Login
		}
		out = append(out, summary)
	}
	return out, nil
}

// BuildListOpenPRArgs returns the argv (excluding the leading "gh") used to
// list open pull requests for the current repository. A limit of zero falls
// back to defaultListLimit.
func BuildListOpenPRArgs(limit int) []string {
	if limit <= 0 {
		limit = defaultListLimit
	}
	return []string{
		"pr", "list",
		"--state", "open",
		"--limit", strconv.Itoa(limit),
		"--json", "number,title,author,headRefName,baseRefName,url,isDraft",
	}
}

// ListOpenPullRequests shells out to `gh pr list` for the repository rooted at
// repoRoot and returns the parsed pull requests. Errors from `gh` (e.g. not
// installed, not authenticated) are wrapped with the command output to aid
// debugging in the UI.
func ListOpenPullRequests(ctx context.Context, repoRoot string, limit int) ([]PullRequestSummary, error) {
	args := BuildListOpenPRArgs(limit)
	result, err := execx.Run(ctx, repoRoot, nil, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w\n%s", err, result.Output())
	}
	return ParsePullRequestList([]byte(result.Stdout))
}
