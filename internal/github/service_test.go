package github

import (
	"testing"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
)

func TestParsePRURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain url on single line",
			in:   "https://github.com/owner/repo/pull/42\n",
			want: "https://github.com/owner/repo/pull/42",
		},
		{
			name: "url is last non-empty line after notices",
			in:   "Creating pull request for branch\n\nhttps://github.com/owner/repo/pull/7\n",
			want: "https://github.com/owner/repo/pull/7",
		},
		{
			name: "trailing whitespace ignored",
			in:   "https://github.com/owner/repo/pull/9   \n\n",
			want: "https://github.com/owner/repo/pull/9",
		},
		{
			name: "url embedded among other text picks the URL line",
			in:   "Warning: foo\nhttps://github.com/o/r/pull/3\nDone.\n",
			want: "https://github.com/o/r/pull/3",
		},
		{
			name: "enterprise host",
			in:   "https://gh.example.com/owner/repo/pull/1\n",
			want: "https://gh.example.com/owner/repo/pull/1",
		},
		{
			name: "no url returns empty",
			in:   "nothing here\n",
			want: "",
		},
		{
			name: "empty output",
			in:   "",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParsePRURL(tc.in)
			if got != tc.want {
				t.Fatalf("ParsePRURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

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
