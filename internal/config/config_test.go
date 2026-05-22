package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMergesDefaultsUserAndRepoConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	userConfigPath := filepath.Join(tmp, "xdg", "gitscribe", "config.toml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("mkdir user config dir: %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("[git]\ndiff_mode = \"unstaged\"\n[logs]\nmax_lines = 42\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	repoRoot := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".gitscribe.toml"), []byte("[git]\ndiff_mode = \"staged\"\n"), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	cfg, err := Load(LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Git.DiffMode != "staged" {
		t.Fatalf("expected repo config to override user config, got %q", cfg.Git.DiffMode)
	}
	if cfg.Logs.MaxLines != 42 {
		t.Fatalf("expected user config value to be preserved, got %d", cfg.Logs.MaxLines)
	}
	if cfg.UI.MinWidth != Defaults().UI.MinWidth {
		t.Fatalf("expected defaults to remain, got %d", cfg.UI.MinWidth)
	}
}

func TestValidateRejectsInvalidEnum(t *testing.T) {
	cfg := Defaults()
	cfg.Git.DiffMode = "broken"
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected invalid enum to fail validation")
	}
}

func TestValidateAcceptsTerminalColorCodes(t *testing.T) {
	cfg := Defaults()
	cfg.Theme.AccentFG = "39"
	cfg.Theme.BadgeBG = "236"
	cfg.Theme.ShortcutKeyFG = "15"

	if err := Validate(cfg); err != nil {
		t.Fatalf("expected ANSI color codes to validate, got %v", err)
	}
}

func TestLoadParsesHooksAndKeybindings(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	repoRoot := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	configBody := `
[[hooks]]
name = "Lint"
command = ["make", "lint"]

[keybindings]
open_commit = ["x"]
quit = ["q", "esc"]
`
	if err := os.WriteFile(filepath.Join(repoRoot, ".gitscribe.toml"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	cfg, err := Load(LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := len(cfg.Hooks), 1; got != want {
		t.Fatalf("hook count = %d, want %d", got, want)
	}
	if cfg.Hooks[0].Name != "Lint" {
		t.Fatalf("hook name = %q, want %q", cfg.Hooks[0].Name, "Lint")
	}
	if got := cfg.Keybindings[ActionOpenCommit]; len(got) != 1 || got[0] != "x" {
		t.Fatalf("open_commit binding = %v, want [x]", got)
	}
	if got := cfg.Keybindings[ActionQuit]; len(got) != 2 || got[1] != "esc" {
		t.Fatalf("quit binding = %v, want [q esc]", got)
	}
}

func TestValidateRejectsUnknownKeybindingAction(t *testing.T) {
	cfg := Defaults()
	cfg.Keybindings = map[string][]string{
		"not_real": {"x"},
	}

	if err := Validate(cfg); err == nil {
		t.Fatalf("expected invalid keybinding action to fail validation")
	}
}

func TestValidateRejectsHookWithoutCommand(t *testing.T) {
	cfg := Defaults()
	cfg.Hooks = []HookConfig{{Name: "Lint"}}

	if err := Validate(cfg); err == nil {
		t.Fatalf("expected hook without command to fail validation")
	}
}

func TestDefaultsHavePreviewSyntaxTheme(t *testing.T) {
	cfg := Defaults()
	if cfg.Git.PreviewSyntaxTheme == "" {
		t.Fatalf("expected non-empty default PreviewSyntaxTheme")
	}
}

func TestValidateRejectsDisallowedGHArgs(t *testing.T) {
	cfg := Defaults()
	cfg.PullRequest.GHArgs = []string{"--repo", "attacker/repo"}
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected disallowed gh_args --repo to fail validation")
	}
}

func TestValidateAcceptsAllowedGHArgs(t *testing.T) {
	cfg := Defaults()
	cfg.PullRequest.GHArgs = []string{"--assignee", "@me", "--draft"}
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected allowed gh_args to pass validation, got %v", err)
	}
}

func TestValidateRejectsInvalidAIModel(t *testing.T) {
	cfg := Defaults()
	cfg.AI.Model = "bad;cmd"
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected invalid ai.model to fail validation")
	}
}

func TestValidateRejectsEmptyPreviewSyntaxTheme(t *testing.T) {
	cfg := Defaults()
	cfg.Git.PreviewSyntaxTheme = ""
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected empty preview_syntax_theme to fail validation")
	}
}

func TestDefaultsHaveNewKeybindingActions(t *testing.T) {
	kb := DefaultKeybindings()
	for _, action := range []string{
		ActionStash, ActionDiscard, ActionReset, ActionAmend,
		ActionFetch, ActionStageAll, ActionCopyPath, ActionIgnore,
		ActionCommitNoVerify, ActionCommitAI, ActionOpenLogs,
		ActionTabDiff, ActionTabRaw, ActionTabPreview,
	} {
		if _, ok := kb[action]; !ok {
			t.Fatalf("missing default keybinding for action %q", action)
		}
	}
}

func TestDefaultOpenShellIsColon(t *testing.T) {
	kb := DefaultKeybindings()
	keys := kb[ActionOpenShell]
	if len(keys) == 0 || keys[0] != ":" {
		t.Fatalf("expected ActionOpenShell default to be ':', got %v", keys)
	}
}

func TestDefaultKeybindingsHasOpenStash(t *testing.T) {
	kb := DefaultKeybindings()
	if keys, ok := kb[ActionOpenStash]; !ok || len(keys) == 0 {
		t.Fatalf("expected ActionOpenStash keybinding, got %v", kb[ActionOpenStash])
	}
}

func TestDefaultKeybindingsHasOpenWorktrees(t *testing.T) {
	kb := DefaultKeybindings()
	if keys, ok := kb[ActionOpenWorktrees]; !ok || len(keys) == 0 {
		t.Fatalf("expected ActionOpenWorktrees keybinding, got %v", kb[ActionOpenWorktrees])
	}
}

func TestDefaultKeybindingsHasCtrlBForBranches(t *testing.T) {
	kb := DefaultKeybindings()
	keys := kb[ActionOpenBranches]
	for _, k := range keys {
		if k == "ctrl+b" {
			return
		}
	}
	t.Fatalf("expected ctrl+b in ActionOpenBranches keybindings, got %v", keys)
}

func TestValidateAcceptsNewKeybindingActions(t *testing.T) {
	cfg := Defaults()
	cfg.Keybindings = map[string][]string{
		ActionStash:    {"S"},
		ActionDiscard:  {"x"},
		ActionCommitAI: {"g"},
		ActionOpenLogs: {"l"},
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestDefaultsCommitTitleMaxLength(t *testing.T) {
	cfg := Defaults()
	if cfg.Commit.TitleMaxLength != 72 {
		t.Fatalf("Commit.TitleMaxLength = %d, want 72", cfg.Commit.TitleMaxLength)
	}
}

func TestDefaultsPullRequestTitleMaxLength(t *testing.T) {
	cfg := Defaults()
	if cfg.PullRequest.TitleMaxLength != 72 {
		t.Fatalf("PullRequest.TitleMaxLength = %d, want 72", cfg.PullRequest.TitleMaxLength)
	}
}

func TestValidateRejectsTooSmallCommitTitleMaxLength(t *testing.T) {
	cfg := Defaults()
	cfg.Commit.TitleMaxLength = 5
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected too-small commit.title_max_length to fail validation")
	}
}

func TestValidateRejectsTooSmallPullRequestTitleMaxLength(t *testing.T) {
	cfg := Defaults()
	cfg.PullRequest.TitleMaxLength = 5
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected too-small pull_request.title_max_length to fail validation")
	}
}

func TestDefaultsPullRequestOpenURLInBrowserIsFalse(t *testing.T) {
	cfg := Defaults()
	if cfg.PullRequest.OpenURLInBrowser {
		t.Fatalf("expected PullRequest.OpenURLInBrowser default false, got true")
	}
}

func TestLoadParsesPullRequestOpenURLInBrowser(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	repoRoot := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	body := "[pull_request]\nopen_url_in_browser = true\n"
	if err := os.WriteFile(filepath.Join(repoRoot, ".gitscribe.toml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	cfg, err := Load(LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.PullRequest.OpenURLInBrowser {
		t.Fatalf("expected PullRequest.OpenURLInBrowser = true after override")
	}
}

func TestLoadOverridesTitleMaxLength(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	repoRoot := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	configBody := `
[commit]
title_max_length = 50

[pull_request]
title_max_length = 100
`
	if err := os.WriteFile(filepath.Join(repoRoot, ".gitscribe.toml"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	cfg, err := Load(LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Commit.TitleMaxLength != 50 {
		t.Fatalf("Commit.TitleMaxLength = %d, want 50", cfg.Commit.TitleMaxLength)
	}
	if cfg.PullRequest.TitleMaxLength != 100 {
		t.Fatalf("PullRequest.TitleMaxLength = %d, want 100", cfg.PullRequest.TitleMaxLength)
	}
}
