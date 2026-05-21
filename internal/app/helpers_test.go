package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestClampBounds(t *testing.T) {
	if clamp(-1, 0, 10) != 0 {
		t.Fatalf("low bound")
	}
	if clamp(20, 0, 10) != 10 {
		t.Fatalf("high bound")
	}
	if clamp(5, 0, 10) != 5 {
		t.Fatalf("middle pass-through")
	}
}

func TestMinMax(t *testing.T) {
	if min(1, 2) != 1 || min(3, 2) != 2 {
		t.Fatalf("min broken")
	}
	if max(1, 2) != 2 || max(3, 2) != 3 {
		t.Fatalf("max broken")
	}
}

func TestTruncateANSIEdgeCases(t *testing.T) {
	if truncateANSI("abc", 0) != "" {
		t.Fatalf("zero width should be empty")
	}
	if truncateANSI("a", 5) != "a" {
		t.Fatalf("short string should pass through")
	}
	if got := truncateANSI("abcdef", 1); got == "" {
		t.Fatalf("width=1 should return one char, got %q", got)
	}
	if got := truncateANSI("abcdef", 3); !strings.HasSuffix(got, "…") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}

func TestPadANSITrimsOversize(t *testing.T) {
	if got := padANSI("abcdef", 3); len(got) != 3 {
		t.Fatalf("padANSI did not trim oversize, got %q", got)
	}
}

func TestNormalizeLinesPadsAndTruncates(t *testing.T) {
	lines := normalizeLines("a\nb", 4, 5)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	lines = normalizeLines("a\nb\nc\nd", 2, 5)
	if len(lines) != 2 {
		t.Fatalf("expected truncation to 2 lines, got %d", len(lines))
	}
}

func TestOverlayCenteredPlacesOverlay(t *testing.T) {
	out := overlayCentered("base\nbase\nbase", "X", 4, 3)
	if !strings.Contains(out, "X") {
		t.Fatalf("expected overlay rendered")
	}
}

func TestColorizeDiffMatchesLines(t *testing.T) {
	m := newReadyTestModel()
	diff := strings.Join([]string{
		"diff --git a/x b/x",
		"index 123..456",
		"--- a/x",
		"+++ b/x",
		"@@ -1 +1 @@",
		"-old",
		"+new",
		"Staged changes",
		"Unstaged changes",
	}, "\n")
	out := m.colorizeDiff(diff)
	if !strings.Contains(out, "@@") {
		t.Fatalf("missing hunk marker")
	}
}

func TestColorizeLogFallsThroughForPlain(t *testing.T) {
	m := newReadyTestModel()
	if m.colorizeLog("plain text") != "plain text" {
		t.Fatalf("expected passthrough")
	}
}

func TestNormalizedKeybindingsSkipsBlanks(t *testing.T) {
	// Single chars preserve case; modifier combos are lowercased.
	got := normalizedKeybindings([]string{"  ", "Q", "esc"})
	if len(got) != 2 || got[0] != "Q" || got[1] != "esc" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestTrimLogShortPassthrough(t *testing.T) {
	if trimLog("one\ntwo", 5) != "one\ntwo" {
		t.Fatalf("expected passthrough")
	}
}

func TestModelInitReturnsCmd(t *testing.T) {
	m := newReadyTestModel()
	if m.Init() == nil {
		t.Fatalf("expected non-nil init cmd")
	}
}

func TestViewLoadingReturnsPlaceholder(t *testing.T) {
	m := newReadyTestModel()
	m.ready = false
	if !strings.Contains(m.View(), "Loading") {
		t.Fatalf("expected Loading text")
	}
}

func TestViewTooSmall(t *testing.T) {
	m := newReadyTestModel()
	m.width = 10
	m.height = 5
	if !strings.Contains(m.View(), "too small") {
		t.Fatalf("expected too-small message")
	}
}

func TestViewWithOverlays(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	m.branchSelector = true
	m.confirm = &confirmState{Action: confirmCommit, Title: "Confirm", Message: "Sure?"}
	m.openHelpModal()
	if !strings.Contains(m.View(), "Help") {
		t.Fatalf("expected topmost overlay to show Help")
	}
}

func TestRenderTreeEmptyWhenLoading(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = nil
	m.loading = true
	if !strings.Contains(m.renderTree(10, 80), "Loading") {
		t.Fatalf("expected loading state")
	}
}

func TestRenderTreeWorkingTreeClean(t *testing.T) {
	m := newReadyTestModel()
	m.tree.Rows = nil
	m.loading = false
	if !strings.Contains(m.renderTree(10, 80), "clean") {
		t.Fatalf("expected clean message")
	}
}

func TestRenderTreeTruncatesWideRows(t *testing.T) {
	m := newReadyTestModelWithFiles(nil)
	m.tree.Rows = []treeRow{{Path: "a", Label: strings.Repeat("x", 200)}}
	out := m.renderTree(10, 10)
	if !strings.Contains(out, "…") {
		t.Fatalf("expected truncation: %s", out)
	}
}

func TestDisplayNameHandlesBlank(t *testing.T) {
	if displayName("   ") != "" {
		t.Fatalf("expected empty for blank path")
	}
	if displayName("pkg/file.go") != "file.go" {
		t.Fatalf("expected base name")
	}
}

func TestVisibleAppTitleAnimates(t *testing.T) {
	m := newReadyTestModel()
	if !strings.Contains(m.visibleAppTitle(), appTitle) {
		t.Fatalf("expected app title contained")
	}
}

func TestHandleGlobalKeyMatchesShellAction(t *testing.T) {
	m := newReadyTestModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	if model.(*Model).modal.kind != modalShell {
		t.Fatalf("expected shell modal")
	}
}
