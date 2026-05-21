package app

import (
	"strings"
	"testing"
)

func TestRenderBranchPillContainsBranchName(t *testing.T) {
	m := newReadyTestModel()
	m.repo.Branch = "feature/x"
	if !strings.Contains(m.renderBranchPill(), "feature/x") {
		t.Fatalf("expected branch name in pill")
	}
}
