package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	colorPattern     = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	ansiColorPattern = regexp.MustCompile(`^(?:[0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	errBadEnum       = errors.New("invalid configuration enum")
)

type LoadOptions struct {
	ExplicitPath string
	RepoRoot     string
}

type Config struct {
	UI          UIConfig            `toml:"ui"`
	Theme       ThemeConfig         `toml:"theme"`
	Git         GitConfig           `toml:"git"`
	Commit      CommitConfig        `toml:"commit"`
	AI          AIConfig            `toml:"ai"`
	PullRequest PullRequestConfig   `toml:"pull_request"`
	History     HistoryConfig       `toml:"history"`
	Logs        LogsConfig          `toml:"logs"`
	Hooks       []HookConfig        `toml:"hooks"`
	Keybindings map[string][]string `toml:"keybindings"`
}

type HookConfig struct {
	Name    string   `toml:"name"`
	Command []string `toml:"command"`
}

type UIConfig struct {
	ShowIntroAnimation bool   `toml:"show_intro_animation"`
	BorderStyle        string `toml:"border_style"`
	MinWidth           int    `toml:"min_width"`
	MinHeight          int    `toml:"min_height"`
}

type ThemeConfig struct {
	StatusFG             string `toml:"status_fg"`
	TextFG               string `toml:"text_fg"`
	AccentFG             string `toml:"accent_fg"`
	AccentAltFG          string `toml:"accent_alt_fg"`
	SuccessFG            string `toml:"success_fg"`
	ErrorFG              string `toml:"error_fg"`
	WarningFG            string `toml:"warning_fg"`
	BadgeFG              string `toml:"badge_fg"`
	BadgeBG              string `toml:"badge_bg"`
	SelectedArrowFG      string `toml:"selected_arrow_fg"`
	ShortcutKeyFG        string `toml:"shortcut_key_fg"`
	ShortcutTextFG       string `toml:"shortcut_text_fg"`
	BranchBG             string `toml:"branch_bg"`
	BranchFG             string `toml:"branch_fg"`
	MainFilesBorder      string `toml:"main_files_border"`
	MainFilesTitle       string `toml:"main_files_title"`
	MainDiffBorder       string `toml:"main_diff_border"`
	MainDiffTitle        string `toml:"main_diff_title"`
	HistoryListBorder    string `toml:"history_list_border"`
	HistoryListTitle     string `toml:"history_list_title"`
	HistoryChangesBorder string `toml:"history_changes_border"`
	HistoryChangesTitle  string `toml:"history_changes_title"`
	CommitTitleBorder    string `toml:"commit_title_border"`
	CommitTitleTitle     string `toml:"commit_title_title"`
	CommitBodyBorder     string `toml:"commit_body_border"`
	CommitBodyTitle      string `toml:"commit_body_title"`
	PRTitleBorder        string `toml:"pr_title_border"`
	PRTitleTitle         string `toml:"pr_title_title"`
	PRBodyBorder         string `toml:"pr_body_border"`
	PRBodyTitle          string `toml:"pr_body_title"`
	LogsBorder           string `toml:"logs_border"`
	LogsTitle            string `toml:"logs_title"`
	HelpBorder           string `toml:"help_border"`
	HelpTitle            string `toml:"help_title"`
	ConfirmBorder        string `toml:"confirm_border"`
	ConfirmTitle         string `toml:"confirm_title"`
	StagedFG             string `toml:"staged_fg"`
	StagedBG             string `toml:"staged_bg"`
	UnstagedFG           string `toml:"unstaged_fg"`
	UnstagedBG           string `toml:"unstaged_bg"`
	ChangedFG            string `toml:"changed_fg"`
	ChangedBG            string `toml:"changed_bg"`
	UntrackedFG          string `toml:"untracked_fg"`
	UntrackedBG          string `toml:"untracked_bg"`
}

type GitConfig struct {
	DiffMode                string   `toml:"diff_mode"`
	UnstagedDiffCommand     []string `toml:"unstaged_diff_command"`
	StagedDiffCommand       []string `toml:"staged_diff_command"`
	DiffCommandPlaceholders []string `toml:"diff_command_placeholders"`
	DiffRenderer            string   `toml:"diff_renderer"`
	BranchSwitchCommand     string   `toml:"branch_switch_command"`
	PreviewSyntaxTheme      string   `toml:"preview_syntax_theme"`
	EditorCommand           string   `toml:"editor_command"`
}

type CommitConfig struct {
	RequireConfirmation bool `toml:"require_confirmation"`
	BodyWrapWidth       int  `toml:"body_wrap_width"`
}

type AIConfig struct {
	Enabled           bool     `toml:"enabled"`
	CommandTemplate   []string `toml:"command_template"`
	Model             string   `toml:"model"`
	Mode              string   `toml:"mode"`
	Style             string   `toml:"style"`
	MessageFormat     string   `toml:"message_format"`
	PromptComplement  string   `toml:"prompt_complement"`
	MaxDiffBytes      int      `toml:"max_diff_bytes"`
	PerFileChunkBytes int      `toml:"per_file_chunk_bytes"`
	SummaryPromptFile string   `toml:"summary_prompt_file"`
	FinalPromptFile   string   `toml:"final_prompt_file"`
	OutputFormat      string   `toml:"output_format"`
}

type PullRequestConfig struct {
	Enabled      bool     `toml:"enabled"`
	DefaultBase  string   `toml:"default_base"`
	TemplatePath string   `toml:"template_path"`
	GHArgs       []string `toml:"gh_args"`
}

type HistoryConfig struct {
	MaxCommits int  `toml:"max_commits"`
	ShowAuthor bool `toml:"show_author"`
}

type LogsConfig struct {
	MaxLines           int  `toml:"max_lines"`
	AutoCloseOnSuccess bool `toml:"auto_close_on_success"`
}

const (
	ActionQuit            = "quit"
	ActionOpenHelp        = "open_help"
	ActionOpenHooks       = "open_hooks"
	ActionOpenShell       = "open_shell"
	ActionRefresh         = "refresh"
	ActionOpenHistory     = "open_history"
	ActionOpenCommit      = "open_commit"
	ActionOpenPullRequest = "open_pull_request"
	ActionPull            = "pull"
	ActionPush            = "push"
	ActionOpenBranches    = "open_branches"
	ActionSubmit          = "submit"
	ActionCancel          = "cancel"
	ActionNextField       = "next_field"
	ActionPrevField       = "prev_field"
	ActionFocusTitle      = "focus_title"
	ActionFocusBody       = "focus_body"
	ActionOpenEditor      = "open_editor"
	ActionClose           = "close"

	ActionStash          = "stash"
	ActionDiscard        = "discard"
	ActionReset          = "reset"
	ActionAmend          = "amend"
	ActionFetch          = "fetch"
	ActionStageAll       = "stage_all"
	ActionCopyPath       = "copy_path"
	ActionIgnore         = "ignore"
	ActionCommitNoVerify = "commit_no_verify"
	ActionCommitAI       = "commit_ai"
	ActionOpenLogs       = "open_logs"
	ActionTabDiff        = "tab_diff"
	ActionTabRaw         = "tab_raw"
	ActionTabPreview     = "tab_preview"
	ActionOpenStash      = "open_stash"
	ActionOpenWorktrees  = "open_worktrees"
)

func DefaultKeybindings() map[string][]string {
	return map[string][]string{
		ActionQuit:            {"q", "esc"},
		ActionOpenHelp:        {"?"},
		ActionOpenHooks:       {"ctrl+g"},
		ActionOpenShell:       {":"},
		ActionRefresh:         {"r"},
		ActionOpenHistory:     {"H"},
		ActionOpenCommit:      {"c"},
		ActionOpenPullRequest: {"ctrl+p"},
		ActionPull:            {"p"},
		ActionPush:            {"P"},
		ActionOpenBranches:    {"b", "ctrl+b"},
		ActionSubmit:          {"ctrl+s"},
		ActionCancel:          {"esc"},
		ActionNextField:       {"tab"},
		ActionPrevField:       {"shift+tab"},
		ActionFocusTitle:      {"shift+left", "shift+up"},
		ActionFocusBody:       {"shift+right", "shift+down"},
		ActionOpenEditor:      {"e"},
		ActionClose:           {"esc", "q", "enter"},

		ActionStash:          {"s"},
		ActionDiscard:        {"d"},
		ActionReset:          {"D"},
		ActionAmend:          {"A"},
		ActionFetch:          {"f"},
		ActionStageAll:       {"ctrl+a"},
		ActionCopyPath:       {"ctrl+o"},
		ActionIgnore:         {"i"},
		ActionCommitNoVerify: {"w"},
		ActionCommitAI:       {"a"},
		ActionOpenLogs:       {"L"},
		ActionTabDiff:        {"1"},
		ActionTabRaw:         {"2"},
		ActionTabPreview:     {"3"},
		ActionOpenStash:      {"S"},
		ActionOpenWorktrees:  {"W"},
	}
}

func SupportedKeybindingActions() map[string]struct{} {
	actions := DefaultKeybindings()
	supported := make(map[string]struct{}, len(actions))
	for action := range actions {
		supported[action] = struct{}{}
	}
	return supported
}

func Defaults() Config {
	return Config{
		UI: UIConfig{
			ShowIntroAnimation: true,
			BorderStyle:        "rounded",
			MinWidth:           100,
			MinHeight:          28,
		},
		Theme: ThemeConfig{
			StatusFG:             "#C0C0C0",
			TextFG:               "#D1D5DB",
			AccentFG:             "#60A5FA",
			AccentAltFG:          "#A78BFA",
			SuccessFG:            "#4ADE80",
			ErrorFG:              "#F87171",
			WarningFG:            "#FDE68A",
			BadgeFG:              "#111827",
			BadgeBG:              "#E5E7EB",
			SelectedArrowFG:      "#60A5FA",
			ShortcutKeyFG:        "#FFFFFF",
			ShortcutTextFG:       "#9CA3AF",
			BranchBG:             "#FFFFFF",
			BranchFG:             "#4B5563",
			MainFilesBorder:      "#666666",
			MainFilesTitle:       "#FFFFFF",
			MainDiffBorder:       "#666666",
			MainDiffTitle:        "#FFFFFF",
			HistoryListBorder:    "#666666",
			HistoryListTitle:     "#FFFFFF",
			HistoryChangesBorder: "#666666",
			HistoryChangesTitle:  "#FFFFFF",
			CommitTitleBorder:    "#666666",
			CommitTitleTitle:     "#FFFFFF",
			CommitBodyBorder:     "#666666",
			CommitBodyTitle:      "#FFFFFF",
			PRTitleBorder:        "#666666",
			PRTitleTitle:         "#FFFFFF",
			PRBodyBorder:         "#666666",
			PRBodyTitle:          "#FFFFFF",
			LogsBorder:           "#666666",
			LogsTitle:            "#FFFFFF",
			HelpBorder:           "#666666",
			HelpTitle:            "#FFFFFF",
			ConfirmBorder:        "#666666",
			ConfirmTitle:         "#FFFFFF",
			StagedFG:             "#A7F3D0",
			StagedBG:             "#064E3B",
			UnstagedFG:           "#FDE68A",
			UnstagedBG:           "#78350F",
			ChangedFG:            "#FCA5A5",
			ChangedBG:            "#7F1D1D",
			UntrackedFG:          "#BFDBFE",
			UntrackedBG:          "#1E3A8A",
		},
		Git: GitConfig{
			DiffMode:                "auto",
			UnstagedDiffCommand:     []string{"git", "diff", "--no-ext-diff", "--", "{file}"},
			StagedDiffCommand:       []string{"git", "diff", "--cached", "--no-ext-diff", "--", "{file}"},
			DiffCommandPlaceholders: []string{"{file}", "{repo_root}"},
			DiffRenderer:            "plain",
			BranchSwitchCommand:     "git switch",
			PreviewSyntaxTheme:      "dracula",
			EditorCommand:           "",
		},
		Commit: CommitConfig{
			RequireConfirmation: true,
			BodyWrapWidth:       72,
		},
		AI: AIConfig{
			Enabled:           true,
			CommandTemplate:   []string{"claude", "--model", "{model}", "--output-format", "{output_format}", "-p"},
			Model:             "sonnet",
			Mode:              "title_and_body",
			Style:             "formal",
			MessageFormat:     "conventional",
			PromptComplement:  "",
			MaxDiffBytes:      120000,
			PerFileChunkBytes: 30000,
			OutputFormat:      "json",
		},
		PullRequest: PullRequestConfig{
			Enabled:      true,
			DefaultBase:  "",
			TemplatePath: "",
			GHArgs:       []string{"--assignee", "@me"},
		},
		History: HistoryConfig{
			MaxCommits: 200,
			ShowAuthor: true,
		},
		Logs: LogsConfig{
			MaxLines:           1000,
			AutoCloseOnSuccess: true,
		},
	}
}

func Load(opts LoadOptions) (Config, error) {
	cfg := Defaults()
	paths, err := configPaths(opts)
	if err != nil {
		return Config{}, err
	}

	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return Config{}, fmt.Errorf("stat %s: %w", path, err)
		}

		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode %s: %w", path, err)
		}
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func UserConfigPath() (string, error) {
	configHome, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); runtime.GOOS != "windows" && xdgHome != "" {
		configHome = xdgHome
	}

	return filepath.Join(configHome, "gitscribe", "config.toml"), nil
}

func Validate(cfg Config) error {
	if err := validateOneOf("ui.border_style", cfg.UI.BorderStyle, "rounded", "normal", "thick", "double"); err != nil {
		return err
	}
	if cfg.UI.MinWidth < 40 || cfg.UI.MinHeight < 12 {
		return fmt.Errorf("ui minimum size is too small")
	}
	if err := validateOneOf("git.diff_mode", cfg.Git.DiffMode, "auto", "staged", "unstaged"); err != nil {
		return err
	}
	if err := validateOneOf("git.diff_renderer", cfg.Git.DiffRenderer, "plain", "ansi"); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Git.PreviewSyntaxTheme) == "" {
		return fmt.Errorf("git.preview_syntax_theme must not be empty")
	}
	if len(cfg.Git.UnstagedDiffCommand) == 0 || len(cfg.Git.StagedDiffCommand) == 0 {
		return fmt.Errorf("git diff commands must not be empty")
	}
	if cfg.Commit.BodyWrapWidth < 20 {
		return fmt.Errorf("commit.body_wrap_width must be at least 20")
	}
	if cfg.AI.Enabled {
		if len(cfg.AI.CommandTemplate) == 0 {
			return fmt.Errorf("ai.command_template must not be empty when AI is enabled")
		}
		if err := validateOneOf("ai.mode", cfg.AI.Mode, "title_and_body", "title_only"); err != nil {
			return err
		}
		if err := validateOneOf("ai.style", cfg.AI.Style, "formal", "neutral", "fun"); err != nil {
			return err
		}
		if err := validateOneOf("ai.message_format", cfg.AI.MessageFormat, "conventional", "emoji", "normal"); err != nil {
			return err
		}
		if cfg.AI.OutputFormat != "json" {
			return fmt.Errorf("ai.output_format must be json")
		}
		if cfg.AI.MaxDiffBytes <= 0 || cfg.AI.PerFileChunkBytes <= 0 {
			return fmt.Errorf("ai diff byte limits must be positive")
		}
	}
	allowedGHFlags := map[string]bool{
		"--assignee": true, "--reviewer": true, "--label": true,
		"--milestone": true, "--project": true, "--draft": true,
	}
	for _, arg := range cfg.PullRequest.GHArgs {
		if strings.HasPrefix(arg, "--") {
			if !allowedGHFlags[arg] {
				return fmt.Errorf("pull_request.gh_args contains disallowed flag: %q", arg)
			}
		}
	}
	if cfg.AI.Enabled {
		safeIdentifier := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
		if cfg.AI.Model != "" && !safeIdentifier.MatchString(cfg.AI.Model) {
			return fmt.Errorf("ai.model contains invalid characters: %q", cfg.AI.Model)
		}
	}
	if cfg.History.MaxCommits <= 0 {
		return fmt.Errorf("history.max_commits must be positive")
	}
	if cfg.Logs.MaxLines <= 0 {
		return fmt.Errorf("logs.max_lines must be positive")
	}
	supportedActions := SupportedKeybindingActions()
	for action, keys := range cfg.Keybindings {
		if _, ok := supportedActions[action]; !ok {
			return fmt.Errorf("keybindings.%s is not a supported action", action)
		}
		if len(keys) == 0 {
			return fmt.Errorf("keybindings.%s must have at least one key", action)
		}
		for _, key := range keys {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("keybindings.%s contains an empty key", action)
			}
		}
	}
	for _, hook := range cfg.Hooks {
		if strings.TrimSpace(hook.Name) == "" {
			return fmt.Errorf("hooks.name must not be empty")
		}
		if len(hook.Command) == 0 {
			return fmt.Errorf("hook %q must have a command", hook.Name)
		}
		for _, part := range hook.Command {
			if strings.TrimSpace(part) == "" {
				return fmt.Errorf("hook %q contains an empty command segment", hook.Name)
			}
		}
	}

	for field, color := range map[string]string{
		"theme.status_fg":              cfg.Theme.StatusFG,
		"theme.text_fg":                cfg.Theme.TextFG,
		"theme.accent_fg":              cfg.Theme.AccentFG,
		"theme.accent_alt_fg":          cfg.Theme.AccentAltFG,
		"theme.success_fg":             cfg.Theme.SuccessFG,
		"theme.error_fg":               cfg.Theme.ErrorFG,
		"theme.warning_fg":             cfg.Theme.WarningFG,
		"theme.badge_fg":               cfg.Theme.BadgeFG,
		"theme.badge_bg":               cfg.Theme.BadgeBG,
		"theme.selected_arrow_fg":      cfg.Theme.SelectedArrowFG,
		"theme.shortcut_key_fg":        cfg.Theme.ShortcutKeyFG,
		"theme.shortcut_text_fg":       cfg.Theme.ShortcutTextFG,
		"theme.branch_bg":              cfg.Theme.BranchBG,
		"theme.branch_fg":              cfg.Theme.BranchFG,
		"theme.main_files_border":      cfg.Theme.MainFilesBorder,
		"theme.main_files_title":       cfg.Theme.MainFilesTitle,
		"theme.main_diff_border":       cfg.Theme.MainDiffBorder,
		"theme.main_diff_title":        cfg.Theme.MainDiffTitle,
		"theme.history_list_border":    cfg.Theme.HistoryListBorder,
		"theme.history_list_title":     cfg.Theme.HistoryListTitle,
		"theme.history_changes_border": cfg.Theme.HistoryChangesBorder,
		"theme.history_changes_title":  cfg.Theme.HistoryChangesTitle,
		"theme.commit_title_border":    cfg.Theme.CommitTitleBorder,
		"theme.commit_title_title":     cfg.Theme.CommitTitleTitle,
		"theme.commit_body_border":     cfg.Theme.CommitBodyBorder,
		"theme.commit_body_title":      cfg.Theme.CommitBodyTitle,
		"theme.pr_title_border":        cfg.Theme.PRTitleBorder,
		"theme.pr_title_title":         cfg.Theme.PRTitleTitle,
		"theme.pr_body_border":         cfg.Theme.PRBodyBorder,
		"theme.pr_body_title":          cfg.Theme.PRBodyTitle,
		"theme.logs_border":            cfg.Theme.LogsBorder,
		"theme.logs_title":             cfg.Theme.LogsTitle,
		"theme.help_border":            cfg.Theme.HelpBorder,
		"theme.help_title":             cfg.Theme.HelpTitle,
		"theme.confirm_border":         cfg.Theme.ConfirmBorder,
		"theme.confirm_title":          cfg.Theme.ConfirmTitle,
		"theme.staged_fg":              cfg.Theme.StagedFG,
		"theme.staged_bg":              cfg.Theme.StagedBG,
		"theme.unstaged_fg":            cfg.Theme.UnstagedFG,
		"theme.unstaged_bg":            cfg.Theme.UnstagedBG,
		"theme.changed_fg":             cfg.Theme.ChangedFG,
		"theme.changed_bg":             cfg.Theme.ChangedBG,
		"theme.untracked_fg":           cfg.Theme.UntrackedFG,
		"theme.untracked_bg":           cfg.Theme.UntrackedBG,
	} {
		if !isValidColorCode(color) {
			return fmt.Errorf("%s must be a #RRGGBB color or terminal color code (0-255)", field)
		}
	}

	return nil
}

func configPaths(opts LoadOptions) ([]string, error) {
	userConfig, err := UserConfigPath()
	if err != nil {
		return nil, err
	}

	paths := []string{userConfig}
	if opts.RepoRoot != "" {
		paths = append(paths, filepath.Join(opts.RepoRoot, ".gitscribe.toml"))
	}
	if opts.ExplicitPath != "" {
		paths = append(paths, opts.ExplicitPath)
	}

	return paths, nil
}

func validateOneOf(field string, value string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}

	return fmt.Errorf("%w: %s=%q (allowed: %s)", errBadEnum, field, value, strings.Join(allowed, ", "))
}

func isValidColorCode(value string) bool {
	return colorPattern.MatchString(value) || ansiColorPattern.MatchString(value)
}
