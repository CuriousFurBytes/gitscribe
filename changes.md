# v0.0.4

## Stash action modal

- `s` key opens a name-input modal instead of a plain confirm dialog
- Enter a name (optional) then press `Enter` to stash; `Esc` cancels
- Named stash uses `git stash push -m <name>`; unnamed uses `git stash push`

## Help menu

- Added `S` hint for stash screen navigation
- Added `W` hint for worktree modal
- `H` hint already present for history screen
- Added `Ctrl+B` hint for branch modal
- All `Ctrl+*` hints now capitalized (`Ctrl` not `ctrl`) throughout the UI

## Main screen

- Raw and Preview tabs are hidden when no file is selected or the selected file is deleted
- Deleted files display with red strikethrough styling in the file tree
- Left panel width reduced from 30% to 20% (more space for the diff/preview)
- Scroll position indicator shown in diff panel header when content overflows
- `S` key opens the new stash screen
- `W` key opens the new worktrees modal

## Branch modal

- `n` key in the branch selector opens a name input to create a new branch
- New branch is based on the currently selected branch in the list
- Enter a name then press `Enter` to create and switch; `Esc` returns to the list
- Branch modal accessible via `b` or `Ctrl+B` from main, history, and stash screens

## Stash screen

- `S` key (from main screen or history screen) navigates to the stash screen
- Two-panel layout mirroring the history screen: left = stash list, right = stash diff
- `Enter` opens confirm dialog to apply the selected stash
- `p` opens confirm dialog to pop (apply + drop) the selected stash
- `d` opens confirm dialog to drop the selected stash
- `r` refreshes the stash list
- `Esc` or `q` returns to the main screen
- Branch modal (`b`/`Ctrl+B`) and worktree modal (`W`) accessible from stash screen

## History screen

- `H` key no longer closes the history screen (avoids accidental close)
- Use `Esc` or `q` to return to the main screen
- Global shortcuts (`?`, `Ctrl+G`, `:`) continue to work from history
- `b` / `Ctrl+B` opens branch modal from history screen
- `W` opens worktree modal from history screen

## Worktree screen (modal)

- `W` key opens a worktrees overlay from main, history, and stash screens
- Lists all worktrees with branch name, short HEAD hash, and current-worktree indicator
- `n` enters a two-step create flow: path input → branch name input → `git worktree add`
- `Esc` closes the modal (or cancels create mode without closing)

## Commit and PR screen

- Status bar now shows `Ctrl+?` instead of `?` for the help hint
- `Ctrl+R` regenerates the AI commit/PR message (clears feedback)
- `Ctrl+E` opens an inline feedback input; `Enter` regenerates with that feedback; `Esc` cancels
- `Ctrl+L` clears both title and body fields and focuses the title

## New keybindings

| Action | Default | Description |
|---|---|---|
| `open_stash` | `S` | Open stash screen |
| `open_worktrees` | `W` | Open worktrees modal |
| `open_branches` | `b`, `Ctrl+B` | Open branch selector (extended) |
