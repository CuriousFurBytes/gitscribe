package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ghcli "github.com/CuriousFurBytes/gitscribe/internal/github"
)

func samplePRs() []ghcli.PullRequestSummary {
	return []ghcli.PullRequestSummary{
		{
			Number:      42,
			Title:       "feat: add foo",
			Author:      "alice",
			HeadRefName: "feature-x",
			BaseRefName: "main",
			URL:         "https://github.com/owner/repo/pull/42",
			IsDraft:     false,
		},
		{
			Number:      7,
			Title:       "wip: bar",
			Author:      "bob",
			HeadRefName: "wip-bar",
			BaseRefName: "main",
			URL:         "https://github.com/owner/repo/pull/7",
			IsDraft:     true,
		},
	}
}

func newReadyPRListModel() *Model {
	m := newReadyTestModel()
	m.width = 200
	m.resize()
	m.screen = screenPRList
	m.prList = samplePRs()
	m.prListIndex = 0
	return m
}

func TestMainOKeyOpensPRListScreen(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenMain
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	got := model.(*Model)
	if got.screen != screenPRList {
		t.Fatalf("O key should open PR list screen, got %q", got.screen)
	}
}

func TestPRListLoadedMsgPopulatesEntries(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPRList
	prs := samplePRs()
	model, _ := m.Update(prListLoadedMsg{entries: prs})
	got := model.(*Model)
	if len(got.prList) != 2 {
		t.Fatalf("expected 2 PR entries, got %d", len(got.prList))
	}
	if got.prList[0].Number != 42 {
		t.Fatalf("first PR number = %d, want 42", got.prList[0].Number)
	}
}

func TestPRListLoadedMsgErrorSetsNotice(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPRList
	model, _ := m.Update(prListLoadedMsg{err: assertErr("gh: command not found")})
	got := model.(*Model)
	if !strings.Contains(got.notice, "gh: command not found") {
		t.Fatalf("expected notice to mention gh error, got %q", got.notice)
	}
}

func TestPRListScreenArrowKeysNavigate(t *testing.T) {
	m := newReadyPRListModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.(*Model).prListIndex != 1 {
		t.Fatalf("down should advance pr list index, got %d", model.(*Model).prListIndex)
	}
	model, _ = model.(*Model).Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.(*Model).prListIndex != 0 {
		t.Fatalf("up should decrement pr list index, got %d", model.(*Model).prListIndex)
	}
}

func TestPRListScreenEscReturnsToMain(t *testing.T) {
	m := newReadyPRListModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model.(*Model).screen != screenMain {
		t.Fatalf("Esc should return to main from PR list screen")
	}
}

func TestPRListScreenQReturnsToMain(t *testing.T) {
	m := newReadyPRListModel()
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if model.(*Model).screen != screenMain {
		t.Fatalf("q should return to main from PR list screen")
	}
}

func TestPRListScreenRKeyRefreshes(t *testing.T) {
	m := newReadyPRListModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatalf("r key should trigger pr list reload")
	}
}

func TestRenderPRListScreenShowsEntries(t *testing.T) {
	m := newReadyPRListModel()
	out := m.renderPRList()
	for _, want := range []string{"#42", "feat: add foo", "alice", "feature-x", "main", "draft"} {
		if !strings.Contains(out, want) {
			t.Fatalf("PR list screen missing %q in output: %s", want, out)
		}
	}
}

func TestRenderPRListScreenEmptyState(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPRList
	m.prList = nil
	m.loading = false
	out := m.renderPRList()
	if !strings.Contains(out, "No open PRs") {
		t.Fatalf("empty PR list should show no-open-PRs message, got: %s", out)
	}
}

func TestRenderPRListScreenLoadingState(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPRList
	m.prList = nil
	m.loading = true
	out := m.renderPRList()
	if !strings.Contains(strings.ToLower(out), "loading") {
		t.Fatalf("loading PR list should show loading hint, got: %s", out)
	}
}

func TestRenderPRListScreenHasHeader(t *testing.T) {
	m := newReadyPRListModel()
	out := m.renderPRList()
	if !strings.Contains(out, "Open Pull Requests") {
		t.Fatalf("PR list screen should have header, got: %s", out)
	}
}

func TestPRListScreenEnterOpensSelectedPRInBrowser(t *testing.T) {
	m := newReadyPRListModel()
	m.prListIndex = 1 // wip: bar / pull/7

	var openedURL string
	prev := openURLInBrowser
	openURLInBrowser = func(url string) error {
		openedURL = url
		return nil
	}
	t.Cleanup(func() { openURLInBrowser = prev })

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*Model)
	if openedURL != "https://github.com/owner/repo/pull/7" {
		t.Fatalf("expected opener to receive selected PR URL, got %q", openedURL)
	}
	if got.screen != screenPRList {
		t.Fatalf("Enter should keep PR list screen visible, got %q", got.screen)
	}
}

func TestPRListScreenEnterReportsBrowserError(t *testing.T) {
	m := newReadyPRListModel()

	prev := openURLInBrowser
	openURLInBrowser = func(url string) error {
		return assertErr("no browser available")
	}
	t.Cleanup(func() { openURLInBrowser = prev })

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*Model)
	if !strings.Contains(got.notice, "no browser available") {
		t.Fatalf("expected notice to mention browser error, got %q", got.notice)
	}
}

func TestPRListScreenEnterNoopOnEmpty(t *testing.T) {
	m := newReadyTestModel()
	m.screen = screenPRList
	m.prList = nil

	called := false
	prev := openURLInBrowser
	openURLInBrowser = func(url string) error {
		called = true
		return nil
	}
	t.Cleanup(func() { openURLInBrowser = prev })

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if called {
		t.Fatalf("Enter on empty PR list must not invoke browser opener")
	}
}
