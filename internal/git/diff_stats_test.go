package git

import "testing"

func TestParseDiffStats(t *testing.T) {
	t.Run("empty output", func(t *testing.T) {
		got, err := ParseDiffStats("")
		if err != nil {
			t.Fatalf("ParseDiffStats() error = %v", err)
		}
		if got != (DiffStats{}) {
			t.Fatalf("expected zero value, got %+v", got)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		got, err := ParseDiffStats("   \n  \n")
		if err != nil {
			t.Fatalf("ParseDiffStats() error = %v", err)
		}
		if got != (DiffStats{}) {
			t.Fatalf("expected zero value, got %+v", got)
		}
	})

	t.Run("insertions and deletions", func(t *testing.T) {
		out := " 3 files changed, 42 insertions(+), 7 deletions(-)\n"
		got, err := ParseDiffStats(out)
		if err != nil {
			t.Fatalf("ParseDiffStats() error = %v", err)
		}
		want := DiffStats{FilesChanged: 3, Insertions: 42, Deletions: 7}
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("single insertion singular form", func(t *testing.T) {
		out := " 1 file changed, 1 insertion(+)\n"
		got, err := ParseDiffStats(out)
		if err != nil {
			t.Fatalf("ParseDiffStats() error = %v", err)
		}
		want := DiffStats{FilesChanged: 1, Insertions: 1, Deletions: 0}
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("only deletions", func(t *testing.T) {
		out := " 2 files changed, 5 deletions(-)\n"
		got, err := ParseDiffStats(out)
		if err != nil {
			t.Fatalf("ParseDiffStats() error = %v", err)
		}
		want := DiffStats{FilesChanged: 2, Insertions: 0, Deletions: 5}
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})
}

func TestDiffStatsIsZero(t *testing.T) {
	if !(DiffStats{}).IsZero() {
		t.Fatalf("expected zero DiffStats to report IsZero()")
	}
	if (DiffStats{FilesChanged: 1}).IsZero() {
		t.Fatalf("expected non-zero DiffStats to report not IsZero()")
	}
}

func TestDiffStatsSummary(t *testing.T) {
	t.Run("zero is empty", func(t *testing.T) {
		if got := (DiffStats{}).Summary(); got != "" {
			t.Fatalf("expected empty summary, got %q", got)
		}
	})

	t.Run("plural files", func(t *testing.T) {
		s := DiffStats{FilesChanged: 3, Insertions: 42, Deletions: 7}
		want := "3 files, +42 -7"
		if got := s.Summary(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("single file", func(t *testing.T) {
		s := DiffStats{FilesChanged: 1, Insertions: 1, Deletions: 0}
		want := "1 file, +1 -0"
		if got := s.Summary(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
