# GitScribe

GitScribe is a keyboard-first Git terminal UI written in Go with Bubble Tea. It keeps status, diffs, history, commit authoring, pull request creation, logs, and optional AI-assisted message generation in one local workflow.

## Requirements

- Go 1.26+
- `git`
- `gh` for pull request creation
- an external AI command that reads JSON from stdin and returns JSON for AI generation

## Installation

### `go install`

```bash
go install github.com/CuriousFurBytes/gitscribe@latest
```

This drops a `gitscribe` binary into `$(go env GOBIN)` (or `$(go env GOPATH)/bin`). Make sure that directory is on your `PATH`.

### Prebuilt binaries

Download the archive for your OS/arch from the [latest GitHub Release](https://github.com/CuriousFurBytes/gitscribe/releases/latest), extract it, and move the `gitscribe` binary onto your `PATH`.

### Build from source

```bash
git clone https://github.com/CuriousFurBytes/gitscribe.git
cd gitscribe
go build .
```

Run it from inside a Git repository:

```bash
./gitscribe
```

## CLI flags

```bash
gitscribe              # open main TUI
gitscribe --commit     # open commit screen directly (exits on success or cancel)
gitscribe --pr         # open PR screen directly (exits on success or cancel)
gitscribe --version    # print version
gitscribe --help       # show flags
gitscribe --config <path>       # use a specific config file
gitscribe --no-animation        # disable startup animation
```

## Keyboard shortcuts

### Main screen

| Key | Action |
|---|---|
| `↑/↓` `j/k` | Navigate file tree |
| `Space` | Stage / unstage file or directory |
| `Ctrl+A` | Stage all / unstage all |
| `1` `2` `3` | Switch diff tab: Diff / Raw / Preview |
| `e` | Open file in `$EDITOR` |
| `Ctrl+O` | Copy file path to clipboard |
| `r` | Refresh |
| `f` | Fetch |
| `p` / `P` | Pull / Push |
| `s` | Stash changes (opens name input) |
| `S` | Open stash screen |
| `d` | Discard unstaged changes (with confirm) |
| `D` | Reset staged changes (with confirm) |
| `i` | Add file to `.gitignore` |
| `c` | Open commit screen |
| `a` | Open commit screen with AI pre-fill |
| `A` | Amend last commit |
| `w` | Open commit screen with `--no-verify` |
| `Ctrl+P` | Open pull request screen |
| `H` | Open commit history |
| `b` / `Ctrl+B` | Switch branch (also opens create-branch flow) |
| `W` | Open worktrees modal |
| `Ctrl+G` | Run hooks (including pre-commit) |
| `:` | Open shell modal |
| `L` | View application logs |
| `?` | Toggle help |
| `q` `Esc` | Quit |

### Commit / PR screen

| Key | Action |
|---|---|
| `Enter` (title field) | Submit |
| `Shift+Enter` (body field) | Submit |
| `Ctrl+A` | Generate AI message |
| `Ctrl+R` | Regenerate AI message |
| `Ctrl+E` | Regenerate AI with feedback (opens inline input) |
| `Ctrl+L` | Clear title and body |
| `Ctrl+W` | Toggle `--no-verify` (commit only) |
| `Tab` | Switch focus between title and body |
| `Esc` | Cancel and return to main |
| `Ctrl+?` | Help |

### Stash screen

| Key | Action |
|---|---|
| `↑/↓` `j/k` | Navigate stash list |
| `Enter` | Apply selected stash |
| `p` | Pop selected stash (apply + drop) |
| `d` | Drop selected stash |
| `r` | Refresh stash list |
| `Esc` / `q` | Return to main |

### Worktrees modal

| Key | Action |
|---|---|
| `↑/↓` `j/k` | Navigate worktree list |
| `n` | Create new worktree (two-step: path, then branch) |
| `Esc` | Close |

### Diff tabs

| Tab | Key | Description |
|---|---|---|
| Diff | `1` | Git diff of the selected file (default) |
| Raw | `2` | Raw file content — editable with configured `editor_command` |
| Preview | `3` | Syntax-highlighted preview; Markdown files render with glamour |

### Commit and PR workflows

- Commit and PR forms show a confirmation modal before running `git commit` or `gh pr create`.
- `Enter` in the title field submits. `Shift+Enter` in the body field submits (requires kitty or compatible terminal).
- `Tab` moves between title and body fields.
- `Ctrl+A` triggers AI generation at any time; `Ctrl+R` regenerates without feedback.
- `Ctrl+E` opens an inline feedback input — type what to change, press `Enter` to regenerate.
- `Ctrl+L` clears both title and body fields.
- `Ctrl+W` toggles `--no-verify` (shown as a badge in the title label).
- `--no-verify` badge is also pre-enabled when launching from `w` in the main screen.
- The hooks modal always includes `.git/hooks/pre-commit` and any custom configured hooks.

### History and diffs

- The main diff panel prefers the staged diff when a file has both staged and unstaged changes.
- The history screen loads recent commits and shows the selected commit patch in the right panel.
- The file tree opens directories expanded by default and keeps nested untracked files visible.
- Switch between Diff / Raw / Preview with `1` / `2` / `3`. Raw and Preview are hidden when no file is selected or the file is deleted.
- Deleted files are shown with red strikethrough in the file tree.
- The file tree uses 20% of the terminal width; the diff panel uses the remaining 80%.
- A scroll indicator appears in the diff panel header when content extends past the visible area.

### Branch management

- Press `b` or `Ctrl+B` to open the branch selector.
- In the branch selector, press `n` to create a new branch based on the currently selected branch.
- The branch selector is available from the main screen, history screen, and stash screen.

### Stash management

- Press `s` to stash changes: a name-input modal appears; press `Enter` to stash (name is optional).
- Press `S` to open the stash screen and browse all stashes.
- In the stash screen, `Enter` applies a stash, `p` pops it, `d` drops it.

### Worktree management

- Press `W` from the main screen, history screen, or stash screen to open the worktree modal.
- Press `n` inside the modal to create a new worktree (path, then branch name).

### Logging

- Application logs are written to a temp file at startup.
- Press `L` in the main screen to view logs in a scrollable modal.
- Logs include all screen transitions, git operations, AI calls, and errors.

### Title and modal behavior

- The GitScribe title animates during the first few seconds after startup without blocking input.
- Pull, push, commit, and PR creation show a spinner and status message in the bottom line while they run.
- The shell runs inside a modal (`:`) so you can execute commands without leaving the TUI flow.
- Help and logs use modal overlays so you keep the current screen context visible.

## Configuration

GitScribe loads TOML config from:

1. `$XDG_CONFIG_HOME/gitscribe/config.toml`
2. `<repo>/.gitscribe.toml`
3. `--config <path>`

See [`examples/config.toml`](examples/config.toml) for the full schema.

Theme colors accept either hex values like `#60A5FA` or terminal palette codes like `39`.

Custom hooks are configured with repeated `[[hooks]]` tables, and keybinding overrides live under `[keybindings]`. Unspecified keybindings keep their built-in defaults.

### AI command contract

GitScribe sends JSON to the configured external command on stdin:

```json
{
  "type": "commit_message",
  "repository_root": "/path/to/repo",
  "branch": "feature/example",
  "diff": "diff --git ...",
  "files": ["a.go", "b.go"],
  "style": "formal",
  "message_format": "conventional",
  "prompt_complement": "",
  "user_feedback": ""
}
```

The command must return JSON on stdout:

```json
{
  "title": "feat: add history screen",
  "body": "Render commit history and the selected patch."
}
```

## Known limitations

- The UI is keyboard-first and does not yet implement advanced mouse workflows.
- AI regeneration-with-feedback is scaffolded but not exposed through a dedicated inline feedback field yet.
- Successful operations auto-close their logs screen by default.
