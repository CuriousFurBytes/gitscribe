package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func newReadyStashModel() *Model {
	m := newReadyTestModel()
	m.width = 200
	m.resize()
	m.screen = screenStash
	m.stashEntries = []git.StashEntry{
		{Index: 0, RefName: "stash@{0}", Message: "On main: my-feature"},
		{Index: 1, RefName: "stash@{1}", Message: "WIP on dev: quick save"},
	}
	m.stashIndex = 0
	return m
}

func TestMainSKeyOpensStashScreen(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenMain
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	got := model.(*Model)
	if got.screen != screenStash {
		t.Fatalf("S key should open stash screen, got %q", got.screen)
	}
}

func TestStashListLoadedMsgSetsEntries(t *testing.T) {
	m := newReadyTestModel()
	entries := []git.StashEntry{{Index: 0, RefName: "stash@{0}", Message: "test"}}
	model, _ := m.Update(stashListLoadedMsg{entries: entries})
	got := model.(*Model)
	if len(got.stashEntries) != 1 {
		t.Fatalf("expected 1 stash entry, got %d", len(got.stashEntries))
	}
}

func TestStashScreenArrowKeysNavigate(t *testing.T) {
	m := newReadyStashModel()
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).stashIndex != 1 {
		t.Fatalf("down should advance stash index")
	}
	if cmd == nil {
		t.Fatalf("expected diff load command after navigation")
	}
}

func TestStashScreenEscReturnsToMain(t *testing.T) {
	m := newReadyStashModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).screen != screenMain {
		t.Fatalf("Esc should return to main from stash screen")
	}
}

func TestStashScreenQReturnsToMain(t *testing.T) {
	m := newReadyStashModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if model.(*Model).screen != screenMain {
		t.Fatalf("q should return to main from stash screen")
	}
}

func TestStashScreenEnterOpensApplyConfirm(t *testing.T) {
	m := newReadyStashModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*Model)
	if got.confirm == nil {
		t.Fatalf("Enter should open confirm dialog for apply")
	}
	if got.confirm.Action != confirmStashApply {
		t.Fatalf("confirm.Action = %q, want confirmStashApply", got.confirm.Action)
	}
}

func TestStashScreenPKeyOpensPopConfirm(t *testing.T) {
	m := newReadyStashModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	got := model.(*Model)
	if got.confirm == nil {
		t.Fatalf("p should open confirm dialog for pop")
	}
	if got.confirm.Action != confirmStashPop {
		t.Fatalf("confirm.Action = %q, want confirmStashPop", got.confirm.Action)
	}
}

func TestStashScreenDKeyOpensDropConfirm(t *testing.T) {
	m := newReadyStashModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := model.(*Model)
	if got.confirm == nil {
		t.Fatalf("d should open confirm dialog for drop")
	}
	if got.confirm.Action != confirmStashDrop {
		t.Fatalf("confirm.Action = %q, want confirmStashDrop", got.confirm.Action)
	}
}

func TestRenderStashScreenShowsTwoPanels(t *testing.T) {
	m := newReadyStashModel()
	out := m.renderStash()
	for _, want := range []string{"Stash", "Changes"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stash screen missing %q: %s", want, out)
		}
	}
}

func TestRenderStashScreenShowsEntries(t *testing.T) {
	m := newReadyStashModel()
	out := m.renderStash()
	if !strings.Contains(out, "my-feature") {
		t.Fatalf("stash screen should show stash messages")
	}
}

func TestRenderStashScreenEmptyState(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenStash
	m.stashEntries = nil
	m.loading = false
	out := m.renderStash()
	if !strings.Contains(out, "No stashes") {
		t.Fatalf("empty stash screen should show no-stashes message")
	}
}

func TestStashScreenRKeyRefreshes(t *testing.T) {
	m := newReadyStashModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatalf("r key should trigger stash list reload")
	}
}

func TestStashScreenPageScrollKeys(t *testing.T) {
	m := newReadyStashModel()
	for _, k := range []string{"pgdown", "pgup", "b", "f"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func TestRenderStashModalVisible(t *testing.T) {
	m := newReadyTestModel()
	m.openStashModal()
	out := m.renderStashModal()
	if !strings.Contains(strings.ToLower(out), "stash") {
		t.Fatalf("stash modal should contain stash text")
	}
}
