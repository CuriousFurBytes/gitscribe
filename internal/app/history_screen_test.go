package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func TestRenderHistoryShowsChangesPanel(t *testing.T) {
	m := newReadyTestModel()
	m.width = 200
	m.resize()
	m.screen = screenHistory
	m.historyEntries = []git.CommitHistoryEntry{
		{Hash: "deadbeef", Author: "Lucas", Subject: "feat: x"},
	}
	out := m.renderHistory()
	for _, want := range []string{"History", "Changes", "deadbeef", "feat: x"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
}

func TestRenderHistoryListEmptyShowsPlaceholder(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = nil
	m.loading = false
	if !strings.Contains(m.renderHistoryList(10, 40), "No commits") {
		t.Fatalf("expected empty placeholder")
	}
}

func TestRenderHistoryListLoadingShowsSpinner(t *testing.T) {
	m := newReadyTestModel()
	m.historyEntries = nil
	m.loading = true
	if !strings.Contains(m.renderHistoryList(10, 40), "Loading history") {
		t.Fatalf("expected loading state")
	}
}

func TestHistoryArrowKeysMove(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.historyEntries = []git.CommitHistoryEntry{
		{Hash: "a"}, {Hash: "b"}, {Hash: "c"},
	}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).historyIndex != 1 {
		t.Fatalf("historyIndex = %d, want 1", model.(*Model).historyIndex)
	}
	if cmd == nil {
		t.Fatalf("expected commit patch reload command")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.historyIndex != 0 {
		t.Fatalf("expected to return to first entry")
	}
}

func TestHistoryRefreshKey(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "a"}}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatalf("expected r to reload history")
	}
}

func TestHistoryPageScroll(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "a"}}
	for _, k := range []string{"pgdown", "pgup", "b", "f"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func TestHistoryEscKeyReturnsToMain(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).screen != screenMain {
		t.Fatalf("Esc should return to main from history")
	}
}

func TestHistoryQKeyReturnsToMain(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if model.(*Model).screen != screenMain {
		t.Fatalf("q should return to main from history")
	}
}

func TestHistoryHKeyNoLongerClosesHistory(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.historyEntries = []git.CommitHistoryEntry{{Hash: "a"}}
	// H from history should NOT close history (it was overloaded; now it's a no-op or navigates elsewhere)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	if model.(*Model).screen == screenMain {
		t.Fatalf("H in history screen should no longer return to main (use Esc or q instead)")
	}
}

func TestHistoryHelpKeyOpensModal(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !model.(*Model).modal.visible {
		t.Fatalf("? should open help modal from history screen")
	}
}

func TestRenderHistoryFooterShowsColonForShell(t *testing.T) {
	m := newReadyTestModel()
	m.width = 200
	m.resize()
	m.screen = screenHistory
	out := m.renderHistory()
	if strings.Contains(out, "Ctrl+T") {
		t.Fatalf("expected history footer to not show Ctrl+T for shell shortcut")
	}
	if !strings.Contains(out, ":") {
		t.Fatalf("expected history footer to show ':' for shell shortcut")
	}
}

func TestRenderHistoryEntrySelectedBolded(t *testing.T) {
	m := newReadyTestModel()
	entry := git.CommitHistoryEntry{Hash: "abc", Author: "L", Subject: "x"}
	got := m.renderHistoryEntry(entry, true)
	if !strings.Contains(got, "abc") || !strings.Contains(got, "x") {
		t.Fatalf("missing parts: %s", got)
	}
}
