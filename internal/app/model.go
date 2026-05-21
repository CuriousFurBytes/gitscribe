package app

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/repo"
	"github.com/CuriousFurBytes/gitscribe/internal/theme"
)

type Options struct {
	DirectCommit bool
	DirectPR     bool
	LogFilePath  string
}

func New(cfg config.Config, repoInfo repo.Info, opts Options) *Model {
	spin := spinner.New()
	spin.Spinner = spinner.Spinner{
		Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
		FPS:    time.Second / 10,
	}

	commitForm := newForm("Commit title", "Commit body", 72)
	prForm := newForm("PR title", "PR body", 120)
	shellInput := textinput.New()
	shellInput.Prompt = "$ "
	shellInput.Placeholder = "git status"
	shellInput.Focus()

	startScreen := screenMain
	var directScreen screen
	directMode := false
	if opts.DirectCommit {
		startScreen = screenCommit
		directScreen = screenCommit
		directMode = true
		commitForm.Title.Focus()
		commitForm.Body.Blur()
	} else if opts.DirectPR {
		startScreen = screenPR
		directScreen = screenPR
		directMode = true
		prForm.Title.Focus()
		prForm.Body.Blur()
	}

	m := &Model{
		cfg:          cfg,
		repo:         repoInfo,
		styles:       theme.New(cfg.Theme),
		screen:       startScreen,
		loading:      true,
		spinner:      spin,
		status:       treeStatus{},
		tree:         treeState{},
		commitForm:   commitForm,
		prForm:       prForm,
		logFilePath:  opts.LogFilePath,
		directMode:   directMode,
		directScreen: directScreen,
	}
	if cfg.UI.ShowIntroAnimation && !directMode {
		m.titleAnimationUntil = time.Now().Add(5 * time.Second)
	}
	branchCreateInput := textinput.New()
	branchCreateInput.Prompt = "> "
	branchCreateInput.Placeholder = "Branch name"

	m.modal.viewport = viewport.New(10, 10)
	m.modal.input = shellInput
	m.branchCreateInput = branchCreateInput
	m.diffViewport = viewport.New(10, 10)
	m.historyViewport = viewport.New(10, 10)
	m.stashViewport = viewport.New(10, 10)

	return m
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		loadRepoCmd(m.repo.Root),
	}
	if m.shouldAnimateTitle() {
		cmds = append(cmds, titleTickCmd())
	}
	return tea.Batch(cmds...)
}

func (m *Model) activeForm() *formState {
	if m.screen == screenPR {
		return &m.prForm
	}
	return &m.commitForm
}
