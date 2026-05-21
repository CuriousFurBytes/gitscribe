package app

import (
	"strings"
	"testing"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestRenderBottomStatusLineEmpty(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus = git.RepoStatus{}
	m.operationStatus = operationStatus{}
	m.notice = ""
	if m.renderBottomStatusLine(80) != "" {
		t.Fatalf("expected empty status line")
	}
}

func TestRenderBottomStatusLineShowsOperationAndNotice(t *testing.T) {
	m := newReadyTestModel()
	m.operationStatus = operationStatus{label: "Pulling"}
	m.notice = "oops"
	out := m.renderBottomStatusLine(80)
	if !strings.Contains(out, "Pulling") || !strings.Contains(out, "oops") {
		t.Fatalf("missing parts in: %s", out)
	}
}

func TestRenderBottomStatusLineWithAheadBehind(t *testing.T) {
	m := newReadyTestModel()
	m.status.RepoStatus = git.RepoStatus{Ahead: 2, Behind: 1}
	out := m.renderBottomStatusLine(80)
	if out == "" {
		t.Fatalf("expected ahead/behind label rendered")
	}
}

func TestRenderShortcutHintsJoins(t *testing.T) {
	m := newReadyTestModel()
	out := m.renderShortcutHints(shortcutHint{Key: "a", Text: "Alpha"}, shortcutHint{Key: "b", Text: "Beta"})
	if !strings.Contains(out, "Alpha") || !strings.Contains(out, "Beta") {
		t.Fatalf("missing hint text: %s", out)
	}
}

func TestLimitShortcutHintsKeepsPinnedLast(t *testing.T) {
	hints := []shortcutHint{
		{Key: "a", Text: "A"},
		{Key: "b", Text: "B"},
		{Key: "c", Text: "C"},
		{Key: "d", Text: "D"},
		{Key: "e", Text: "E"},
		{Key: "f", Text: "F"},
		{Key: "g", Text: "Help"},
	}
	out := limitShortcutHints(hints)
	if len(out) != maxBottomHints {
		t.Fatalf("len = %d, want %d", len(out), maxBottomHints)
	}
	if out[len(out)-1].Text != "Help" {
		t.Fatalf("expected Help to be pinned last, got %v", out[len(out)-1])
	}
}

func TestLimitShortcutHintsShortListPassthrough(t *testing.T) {
	hints := []shortcutHint{{Key: "a", Text: "A"}}
	out := limitShortcutHints(hints)
	if len(out) != 1 {
		t.Fatalf("expected passthrough")
	}
}

func TestSingleLineCollapsesWhitespace(t *testing.T) {
	got := singleLine("\n  hello\nworld\t!  \n")
	if got != "hello world !" {
		t.Fatalf("singleLine = %q", got)
	}
}
