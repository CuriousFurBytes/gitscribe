package git

import "testing"

func TestParseStatusScenarios(t *testing.T) {
	t.Run("clean repo", func(t *testing.T) {
		status, err := ParseStatus("# branch.head main\n# branch.ab +0 -0\n")
		if err != nil {
			t.Fatalf("ParseStatus() error = %v", err)
		}
		if status.Branch != "main" {
			t.Fatalf("expected branch main, got %q", status.Branch)
		}
		if len(status.Files) != 0 {
			t.Fatalf("expected no files, got %d", len(status.Files))
		}
	})

	t.Run("mixed working tree", func(t *testing.T) {
		output := `# branch.head main
# branch.ab +2 -1
1 .M N... 100644 100644 100644 abc1234 abc1234 modified.txt
1 M. N... 100644 100644 100644 abc1234 abc1234 staged.txt
1 MM N... 100644 100644 100644 abc1234 abc1234 mixed.txt
? new.txt
1 .D N... 100644 100644 000000 abc1234 0000000 deleted.txt
2 R. N... 100644 100644 100644 abc1234 abc1234 R100 renamed.txt	old-name.txt
`
		status, err := ParseStatus(output)
		if err != nil {
			t.Fatalf("ParseStatus() error = %v", err)
		}

		if !status.HasUncommitted || !status.HasStaged || !status.HasUnstaged || !status.HasUntracked {
			t.Fatalf("expected all working tree flags to be true: %+v", status)
		}
		if status.Ahead != 2 || status.Behind != 1 {
			t.Fatalf("expected ahead/behind 2/1, got %d/%d", status.Ahead, status.Behind)
		}
		if len(status.Files) != 6 {
			t.Fatalf("expected 6 files, got %d", len(status.Files))
		}

		mixed := findChange(t, status.Files, "mixed.txt")
		if !mixed.IsChangedAfterStage {
			t.Fatalf("expected mixed.txt to be changed after staging, got %+v", mixed)
		}

		renamed := findChange(t, status.Files, "renamed.txt")
		if !renamed.IsRenamed || renamed.OriginalPath != "old-name.txt" {
			t.Fatalf("expected renamed file metadata, got %+v", renamed)
		}

		untracked := findChange(t, status.Files, "new.txt")
		if !untracked.IsUntracked || untracked.UnstagedStatus != StatusUntracked {
			t.Fatalf("expected untracked file metadata, got %+v", untracked)
		}
	})
}

func TestBuildCommitArgs(t *testing.T) {
	got := BuildCommitArgs("feat: add tree", "body", false)
	want := []string{"commit", "-m", "feat: add tree", "-m", "body"}
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestChooseDiffMode(t *testing.T) {
	change := FileChange{StagedStatus: StatusModified, UnstagedStatus: StatusModified}
	if got := ChooseDiffMode(change, DiffModeAuto); got != DiffModeStaged {
		t.Fatalf("auto mode should prefer staged diff, got %q", got)
	}
	if got := ChooseDiffMode(change, DiffModeUnstaged); got != DiffModeUnstaged {
		t.Fatalf("explicit unstaged mode should prefer unstaged diff, got %q", got)
	}
}

func findChange(t *testing.T, files []FileChange, path string) FileChange {
	t.Helper()
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("path %q not found in %+v", path, files)
	return FileChange{}
}
