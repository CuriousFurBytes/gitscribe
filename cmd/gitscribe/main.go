package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/app"
	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/logger"
	"github.com/CuriousFurBytes/gitscribe/internal/repo"
)

const version = "0.1.0"

func main() {
	var (
		configPath   string
		noAnimation  bool
		showVersion  bool
		directCommit bool
		directPR     bool
	)

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&noAnimation, "no-animation", false, "disable intro animation")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&directCommit, "commit", false, "open commit screen directly (exit on cancel or success)")
	flag.BoolVar(&directPR, "pr", false, "open PR screen directly (exit on cancel or success)")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return
	}

	if directCommit && directPR {
		fmt.Fprintln(os.Stderr, "gitscribe: --commit and --pr cannot be used together")
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "gitscribe: warning: could not init logger:", err)
	}

	ctx := context.Background()
	repoInfo, err := repo.Detect(ctx, ".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitscribe: not inside a Git repository")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg, err := config.Load(config.LoadOptions{
		ExplicitPath: configPath,
		RepoRoot:     repoInfo.Root,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitscribe: invalid configuration")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if noAnimation {
		cfg.UI.ShowIntroAnimation = false
	}

	opts := app.Options{
		DirectCommit: directCommit,
		DirectPR:     directPR,
		LogFilePath:  logger.LogFile,
	}

	logger.Info("gitscribe starting", "version", version, "repo", repoInfo.Root)

	model := app.New(cfg, repoInfo, opts)
	model.RestoreDraft()
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
