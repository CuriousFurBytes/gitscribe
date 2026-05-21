package app

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	"github.com/CuriousFurBytes/gitscribe/internal/ai"
	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
	ghcli "github.com/CuriousFurBytes/gitscribe/internal/github"
	"github.com/CuriousFurBytes/gitscribe/internal/repo"
	"github.com/CuriousFurBytes/gitscribe/internal/theme"
)

const appTitle = "GitScribe 󰊢"
const spinnerDot = "⣾"

type screen string

const (
	screenMain    screen = "main"
	screenHistory screen = "history"
	screenCommit  screen = "commit"
	screenPR      screen = "pr"
	screenLogs    screen = "logs"
	screenStash   screen = "stash"
	screenPRList  screen = "pr_list"
)

type modalKind string

const (
	modalNone     modalKind = ""
	modalHelp     modalKind = "help"
	modalLogs     modalKind = "logs"
	modalHooks    modalKind = "hooks"
	modalShell    modalKind = "shell"
	modalStash    modalKind = "stash"
	modalWorktree modalKind = "worktree"
	modalPRURL    modalKind = "pr_url"
)

type confirmAction string

const (
	confirmCommit     confirmAction = "commit"
	confirmPR         confirmAction = "pr"
	confirmDiscard    confirmAction = "discard"
	confirmReset      confirmAction = "reset"
	confirmStash      confirmAction = "stash"
	confirmIgnore     confirmAction = "ignore"
	confirmStashApply confirmAction = "stash_apply"
	confirmStashPop   confirmAction = "stash_pop"
	confirmStashDrop  confirmAction = "stash_drop"
)

type diffTab int

const (
	tabDiff diffTab = iota
	tabRaw
	tabPreview
)

type treeRow struct {
	Path   string
	Label  string
	IsDir  bool
	Level  int
	Change *git.FileChange
}

type treeState struct {
	Rows  []treeRow
	Index int
}

type formFocus int

const (
	focusTitle formFocus = iota
	focusBody
)

type formState struct {
	Title                textinput.Model
	Body                 textarea.Model
	Focus                formFocus
	Error                string
	Loading              bool
	UserFeedback         string
	NoVerify             bool
	FeedbackInputVisible bool
	FeedbackInput        textinput.Model
}

type confirmState struct {
	Action  confirmAction
	Title   string
	Message string
}

type shortcutHint struct {
	Key  string
	Text string
}

type modalState struct {
	visible     bool
	kind        modalKind
	title       string
	body        string
	success     bool
	loading     bool
	refreshRepo bool
	returnTo    screen
	viewport    viewport.Model
	input       textinput.Model
	hooks       []hookOption
	hookIndex   int
	selecting   bool
}

type hookOption struct {
	Name      string
	Command   []string
	PreCommit bool
}

type operationStatus struct {
	label string
}

type toastKind string

const (
	toastSuccess toastKind = "success"
	toastInfo    toastKind = "info"
)

type toastState struct {
	text string
	kind toastKind
}

func (t toastState) visible() bool {
	return t.text != ""
}

type treeStatus struct {
	RepoStatus git.RepoStatus
}

type Model struct {
	cfg     config.Config
	repo    repo.Info
	styles  theme.Styles
	screen  screen
	width   int
	height  int
	ready   bool
	loading bool
	notice  string
	spinner spinner.Model

	titleAnimationFrame int
	titleAnimationUntil time.Time
	operationStatus     operationStatus

	status treeStatus

	historyEntries []git.CommitHistoryEntry
	historyIndex   int

	branches    []string
	branchIndex int

	diffViewport    viewport.Model
	diffContent     string
	diffModeLabel   string
	diffTab         diffTab
	rawContent      string
	previewContent  string
	historyViewport viewport.Model

	tree treeState

	commitForm formState
	prForm     formState
	modal      modalState

	branchSelector    bool
	branchCreating    bool
	branchCreateInput textinput.Model
	branchFilter      string

	worktreeEntries    []git.WorktreeEntry
	worktreeIndex      int
	worktreeCreating   bool
	worktreeCreateStep int
	worktreeCreatePath string

	stashEntries  []git.StashEntry
	stashIndex    int
	stashViewport viewport.Model

	prList      []ghcli.PullRequestSummary
	prListIndex int

	confirm *confirmState

	logFilePath  string
	directMode   bool
	directScreen screen
	amendMode    bool

	toast toastState
}

type repoLoadedMsg struct {
	status   git.RepoStatus
	branches []string
	err      error
}

type diffLoadedMsg struct {
	path    string
	content string
	mode    string
	err     error
}

type historyLoadedMsg struct {
	entries []git.CommitHistoryEntry
	err     error
}

type commitPatchLoadedMsg struct {
	hash    string
	content string
	err     error
}

type operationResultMsg struct {
	title           string
	output          string
	stderr          string
	success         bool
	successReturnTo screen
	failureReturnTo screen
	refreshRepo     bool
	clearCommit     bool
	clearPR         bool
	alwaysModal     bool
	modalKind       modalKind
	prURL           string
	err             error
}

type aiGeneratedMsg struct {
	target string
	resp   ai.Response
	err    error
}

type prTemplateLoadedMsg struct {
	body string
	err  error
}

type titleTickMsg struct{}

type toastMsg struct {
	text string
	kind toastKind
}

type toastExpireMsg struct{}

const toastDuration = 3 * time.Second

type shellCommandResultMsg struct {
	output string
	err    error
}

type fileEditorFinishedMsg struct {
	err error
}

type editorFinishedMsg struct {
	target screen
	focus  formFocus
	value  string
	err    error
}

type rawFileLoadedMsg struct {
	path    string
	content string
	err     error
}

type previewLoadedMsg struct {
	path    string
	content string
	err     error
}

type copyPathMsg struct {
	path string
	err  error
}

type worktreeListLoadedMsg struct {
	entries []git.WorktreeEntry
	err     error
}

type stashListLoadedMsg struct {
	entries []git.StashEntry
	err     error
}

type stashDiffLoadedMsg struct {
	ref     string
	content string
	err     error
}

type prListLoadedMsg struct {
	entries []ghcli.PullRequestSummary
	err     error
}
