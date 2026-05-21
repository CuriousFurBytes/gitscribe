package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderBranchSelectorListsBranches(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	out := m.renderBranchSelector()
	for _, want := range []string{"main", "dev", "Switch", "Cancel"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
}

func TestBranchSelectorArrowsAndEnter(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev", "feat"}
	m.branchIndex = 0
	m.branchSelector = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).branchIndex != 1 {
		t.Fatalf("branchIndex = %d, want 1", model.(*Model).branchIndex)
	}
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.(*Model).branchIndex != 0 {
		t.Fatalf("expected branchIndex back to 0")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected switch command")
	}
	if m.branchSelector {
		t.Fatalf("expected branch selector closed after enter")
	}
}

func TestBranchSelectorEsc(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	m.branchSelector = true
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).branchSelector {
		t.Fatalf("expected esc to close branch selector")
	}
}

func TestBranchSelectorCtrlNEntersCreateMode(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	m.branchSelector = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := model.(*Model)
	if !got.branchCreating {
		t.Fatalf("Ctrl+N should set branchCreating = true")
	}
	if !got.branchSelector {
		t.Fatalf("branch selector should remain open in create mode")
	}
}

func TestBranchSelectorTypingFiltersList(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "develop", "feature/login", "fix/typo"}
	m.branchIndex = 0
	m.branchSelector = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	got := model.(*Model)
	if got.branchFilter != "f" {
		t.Fatalf("branchFilter = %q, want %q", got.branchFilter, "f")
	}
	out := got.renderBranchSelector()
	if !strings.Contains(out, "feature/login") || !strings.Contains(out, "fix/typo") {
		t.Fatalf("expected filtered list to include feature/login and fix/typo, got: %s", out)
	}
	if strings.Contains(out, "main") || strings.Contains(out, "develop") {
		t.Fatalf("expected non-matching branches hidden, got: %s", out)
	}
	if !strings.Contains(out, "filter") {
		t.Fatalf("expected filter input line in render, got: %s", out)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got = model.(*Model)
	if got.branchFilter != "fe" {
		t.Fatalf("branchFilter after second char = %q, want %q", got.branchFilter, "fe")
	}
	out = got.renderBranchSelector()
	if !strings.Contains(out, "feature/login") {
		t.Fatalf("expected feature/login to remain visible, got: %s", out)
	}
	if strings.Contains(out, "fix/typo") {
		t.Fatalf("expected fix/typo to be filtered out, got: %s", out)
	}
}

func TestBranchSelectorBackspaceShrinksFilter(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchSelector = true
	m.branchFilter = "de"

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	got := model.(*Model)
	if got.branchFilter != "d" {
		t.Fatalf("backspace should shrink filter to %q, got %q", "d", got.branchFilter)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	got = model.(*Model)
	if got.branchFilter != "" {
		t.Fatalf("backspace from one-char should clear filter, got %q", got.branchFilter)
	}
}

func TestBranchSelectorEscClearsFilterBeforeClosing(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchSelector = true
	m.branchFilter = "ma"

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.branchFilter != "" {
		t.Fatalf("Esc with non-empty filter should clear filter, got %q", got.branchFilter)
	}
	if !got.branchSelector {
		t.Fatalf("Esc with non-empty filter should not close selector")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got = model.(*Model)
	if got.branchSelector {
		t.Fatalf("second Esc with empty filter should close selector")
	}
}

func TestBranchSelectorArrowsNavigateFilteredList(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "develop", "feature/login", "feature/logout"}
	m.branchSelector = true
	m.branchIndex = 0

	// Type "feat" to narrow to two feature branches
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	if m.branchFilter != "feat" {
		t.Fatalf("expected filter 'feat', got %q", m.branchFilter)
	}
	// Should clamp branchIndex to within filtered list (length 2)
	if m.branchIndex < 0 || m.branchIndex > 1 {
		t.Fatalf("branchIndex %d outside filtered range", m.branchIndex)
	}

	// Move down within filtered list
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.branchIndex != 1 {
		t.Fatalf("down should move to index 1 in filtered list, got %d", m.branchIndex)
	}

	// Cannot move past last filtered item
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.branchIndex != 1 {
		t.Fatalf("down at end of filtered list should clamp, got %d", m.branchIndex)
	}

	// Enter switches to the branch at filtered index 1 (feature/logout)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter should issue switch command")
	}
}

func TestBranchSelectorOpenResetsFilter(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchSelector = false
	m.branchFilter = "stale"

	// Simulate opening via main screen 'b' key
	m.screen = screenMain
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if !m.branchSelector {
		t.Fatalf("'b' should open branch selector")
	}
	if m.branchFilter != "" {
		t.Fatalf("opening selector should reset filter, got %q", m.branchFilter)
	}
}

func TestBranchSelectorCreateEscReturnsToList(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	m.branchSelector = true
	m.branchCreating = true

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(*Model)
	if got.branchCreating {
		t.Fatalf("Esc should exit create mode")
	}
	if !got.branchSelector {
		t.Fatalf("Esc in create mode should return to branch list, not close selector")
	}
}

func TestBranchSelectorCreateEnterTriggersCreate(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main", "dev"}
	m.branchIndex = 0
	m.branchSelector = true
	m.branchCreating = true
	m.branchCreateInput.SetValue("feature/new")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter with name should trigger create branch command")
	}
	if m.branchSelector {
		t.Fatalf("branch selector should close after creating branch")
	}
}

func TestRenderBranchSelectorShowsCreateHint(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"main"}
	out := m.renderBranchSelector()
	if !strings.Contains(out, "New") && !strings.Contains(out, "n") {
		t.Fatalf("branch selector should show hint for creating new branch")
	}
}

func TestHistoryCtrlBOpensBranchSelector(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenHistory
	m.branches = []string{"main", "dev"}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	if !model.(*Model).branchSelector {
		t.Fatalf("Ctrl+B in history screen should open branch selector")
	}
}

func TestBranchSelectorDownClamps(t *testing.T) {
	m := newReadyTestModel()
	m.branches = []string{"only"}
	m.branchSelector = true
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.branchIndex != 0 {
		t.Fatalf("expected index clamped at 0")
	}
}

func TestFilterBranches(t *testing.T) {
	branches := []string{
		"main",
		"develop",
		"feature/login",
		"feature/Logout",
		"fix/typo",
		"release/1.0",
	}
	cases := []struct {
		name   string
		query  string
		input  []string
		want   []string
	}{
		{
			name:  "empty query returns all branches",
			query: "",
			input: branches,
			want:  branches,
		},
		{
			name:  "whitespace-only query returns all branches",
			query: "   ",
			input: branches,
			want:  branches,
		},
		{
			name:  "substring match is case-insensitive",
			query: "LOGIN",
			input: branches,
			want:  []string{"feature/login"},
		},
		{
			name:  "matches multiple branches",
			query: "feature",
			input: branches,
			want:  []string{"feature/login", "feature/Logout"},
		},
		{
			name:  "matches across slashes",
			query: "ure/log",
			input: branches,
			want:  []string{"feature/login", "feature/Logout"},
		},
		{
			name:  "no matches returns empty slice",
			query: "nope",
			input: branches,
			want:  []string{},
		},
		{
			name:  "trims surrounding whitespace from query",
			query: "  main  ",
			input: branches,
			want:  []string{"main"},
		},
		{
			name:  "nil input returns nil",
			query: "main",
			input: nil,
			want:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterBranches(tc.input, tc.query)
			if len(got) != len(tc.want) {
				t.Fatalf("filterBranches(%v, %q) = %v, want %v", tc.input, tc.query, got, tc.want)
			}
			for i, b := range got {
				if b != tc.want[i] {
					t.Fatalf("filterBranches(%v, %q)[%d] = %q, want %q", tc.input, tc.query, i, b, tc.want[i])
				}
			}
		})
	}
}
