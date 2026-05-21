package app

import (
	"testing"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestChangesUnderPathRoot(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "a.go"}, {Path: "pkg/b.go"},
	})
	got := m.changesUnderPath(".")
	if len(got) != 2 {
		t.Fatalf("expected all files, got %d", len(got))
	}
}

func TestChangesUnderPathPrefix(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "pkg/a.go"}, {Path: "pkg/b.go"}, {Path: "other/c.go"},
	})
	got := m.changesUnderPath("pkg")
	if len(got) != 2 {
		t.Fatalf("expected 2 in pkg, got %d", len(got))
	}
}

func TestSetCurrentBranchIndexFallsBackToZero(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.repo.Branch = "missing"
	m.setCurrentBranchIndex()
	if m.branchIndex != 0 {
		t.Fatalf("expected fallback to 0")
	}
}

func TestSetCurrentBranchIndexFindsMatch(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.repo.Branch = "dev"
	m.setCurrentBranchIndex()
	if m.branchIndex != 1 {
		t.Fatalf("expected idx 1, got %d", m.branchIndex)
	}
}

func TestSelectedFileChangeWithoutChange(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = []treeRow{{Path: "x", IsDir: true}}
	m.tree.Index = 0
	if _, ok := m.selectedFileChange(); ok {
		t.Fatalf("expected directory row to return no file")
	}
}

func TestSelectedTreeRowOutOfRange(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Index = 999
	if _, ok := m.selectedTreeRow(); ok {
		t.Fatalf("expected out-of-range to return false")
	}
}

func TestLoadSelectionDiffCmdDirectory(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{{Path: "pkg/a.go", UnstagedStatus: git.StatusModified}})
	for i, row := range m.tree.Rows {
		if row.IsDir && row.Path != "." {
			m.tree.Index = i
			break
		}
	}
	if m.loadSelectionDiffCmd() == nil {
		t.Fatalf("expected directory diff cmd")
	}
}

func TestLoadSelectionDiffCmdNoSelection(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = nil
	if m.loadSelectionDiffCmd() != nil {
		t.Fatalf("expected nil cmd with empty rows")
	}
}

func TestSetDiffEmptyShowsPlaceholder(t *testing.T) {
	m := newReadyTestModel()
	m.setDiff("", "")
	if m.diffViewport.View() == "" {
		t.Fatalf("expected placeholder content")
	}
}

func TestRebuildTreePreservesSelection(t *testing.T) {
	m := newReadyTestModelWithFiles([]git.FileChange{
		{Path: "a.go", UnstagedStatus: git.StatusModified},
		{Path: "b.go", UnstagedStatus: git.StatusModified},
	})
	for i, row := range m.tree.Rows {
		if row.Path == "b.go" {
			m.tree.Index = i
			break
		}
	}
	m.rebuildTree()
	if m.tree.Rows[m.tree.Index].Path != "b.go" {
		t.Fatalf("expected selection preserved on rebuild")
	}
}
