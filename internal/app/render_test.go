package app

import (
	"strings"
	"testing"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestRenderBaseScreenSwitchesPerScreen(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "x"}}
	for _, sc := range []screen{screenHistory, screenCommit, screenPR, screenMain} {
		m.screen = sc
		if m.renderBaseScreen() == "" {
			t.Fatalf("expected non-empty render for screen %q", sc)
		}
	}
}

func TestRenderModalSwitchesPerKind(t *testing.T) {
	m := newReadyTestModel()
	m.openHelpModal()
	if m.renderModal() == "" {
		t.Fatalf("help modal empty")
	}
	m.openShellModal()
	if m.renderModal() == "" {
		t.Fatalf("shell modal empty")
	}
	m.openHooksModal()
	if m.renderModal() == "" {
		t.Fatalf("hooks modal empty")
	}
	m.openLogsModal(operationResultMsg{title: "Run", success: true, output: "log"})
	if m.renderModal() == "" {
		t.Fatalf("logs modal empty")
	}
}

func TestStatusBadgesCoversAllCombinations(t *testing.T) {
	m := newReadyTestModel()
	cases := []git.FileChange{
		{IsUntracked: true, UnstagedStatus: git.StatusUntracked},
		{StagedStatus: git.StatusModified, UnstagedStatus: git.StatusModified, IsChangedAfterStage: true},
		{StagedStatus: git.StatusConflicted},
		{UnstagedStatus: git.StatusConflicted},
		{UnstagedStatus: git.StatusModified},
	}
	for _, c := range cases {
		out := m.statusBadges(c)
		if out == "" {
			t.Fatalf("expected badges for %+v", c)
		}
	}
}

func TestRenderMainDiffModeLabel(t *testing.T) {
	m := newReadyTestModel()
	m.diffModeLabel = "Working"
	out := m.renderMain()
	if !strings.Contains(out, "Working") {
		t.Fatalf("expected diff mode label in header")
	}
}

func TestColorizeLogDispatchToDiff(t *testing.T) {
	m := newReadyTestModel()
	out := m.colorizeLog("@@ hunk @@\nplain")
	if !strings.Contains(out, "hunk") {
		t.Fatalf("expected hunk text preserved: %s", out)
	}
}
