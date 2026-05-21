package git

import "testing"

func TestParseBranchCommits(t *testing.T) {
	t.Run("empty output", func(t *testing.T) {
		if got := ParseBranchCommits(""); got != nil {
			t.Fatalf("expected nil for empty input, got %+v", got)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		if got := ParseBranchCommits("\n  \n"); got != nil {
			t.Fatalf("expected nil for whitespace input, got %+v", got)
		}
	})

	t.Run("single commit", func(t *testing.T) {
		out := "abc1234\tfeat: add thing\n"
		got := ParseBranchCommits(out)
		want := []CommitSummary{{ShortSHA: "abc1234", Subject: "feat: add thing"}}
		if len(got) != len(want) || got[0] != want[0] {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("multiple commits preserve order", func(t *testing.T) {
		out := "abc1234\tfeat: add thing\ndef5678\tfix: a bug\n0011aa2\tdocs: tweak\n"
		got := ParseBranchCommits(out)
		want := []CommitSummary{
			{ShortSHA: "abc1234", Subject: "feat: add thing"},
			{ShortSHA: "def5678", Subject: "fix: a bug"},
			{ShortSHA: "0011aa2", Subject: "docs: tweak"},
		}
		if len(got) != len(want) {
			t.Fatalf("expected %d entries, got %d (%+v)", len(want), len(got), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("entry %d: expected %+v, got %+v", i, want[i], got[i])
			}
		}
	})

	t.Run("skips malformed lines", func(t *testing.T) {
		out := "abc1234\tfeat: add thing\nno-tab-here\ndef5678\tfix: a bug\n"
		got := ParseBranchCommits(out)
		if len(got) != 2 {
			t.Fatalf("expected 2 entries, got %d (%+v)", len(got), got)
		}
		if got[0].ShortSHA != "abc1234" || got[1].ShortSHA != "def5678" {
			t.Fatalf("unexpected entries: %+v", got)
		}
	})

	t.Run("subject may contain tabs", func(t *testing.T) {
		// SplitN with N=2 means anything after the first tab stays in subject.
		out := "abc1234\tfeat: add\tnested\tstuff\n"
		got := ParseBranchCommits(out)
		if len(got) != 1 {
			t.Fatalf("expected 1 entry, got %d (%+v)", len(got), got)
		}
		if got[0].Subject != "feat: add\tnested\tstuff" {
			t.Fatalf("unexpected subject: %q", got[0].Subject)
		}
	})
}
