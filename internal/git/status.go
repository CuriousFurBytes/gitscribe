package git

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParseStatus(output string) (RepoStatus, error) {
	status := RepoStatus{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# branch.head "):
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			parts := strings.Fields(strings.TrimPrefix(line, "# branch.ab "))
			if len(parts) == 2 {
				ahead, err := strconv.Atoi(strings.TrimPrefix(parts[0], "+"))
				if err != nil {
					return RepoStatus{}, fmt.Errorf("parse branch ahead count: %w", err)
				}
				behind, err := strconv.Atoi(strings.TrimPrefix(parts[1], "-"))
				if err != nil {
					return RepoStatus{}, fmt.Errorf("parse branch behind count: %w", err)
				}
				status.Ahead = ahead
				status.Behind = behind
			}
		case strings.HasPrefix(line, "? "):
			path := strings.TrimPrefix(line, "? ")
			change := FileChange{
				Path:           path,
				DisplayPath:    path,
				StagedStatus:   StatusUnmodified,
				UnstagedStatus: StatusUntracked,
				IsUntracked:    true,
			}
			status.HasUntracked = true
			status.HasUncommitted = true
			status.HasUnstaged = true
			status.Files = append(status.Files, change)
		case strings.HasPrefix(line, "1 "):
			change, err := parseOrdinaryLine(line)
			if err != nil {
				return RepoStatus{}, err
			}
			applyChangeFlags(&status, &change)
			status.Files = append(status.Files, change)
		case strings.HasPrefix(line, "2 "):
			change, err := parseRenameLine(line)
			if err != nil {
				return RepoStatus{}, err
			}
			applyChangeFlags(&status, &change)
			status.Files = append(status.Files, change)
		case strings.HasPrefix(line, "u "):
			change, err := parseUnmergedLine(line)
			if err != nil {
				return RepoStatus{}, err
			}
			applyChangeFlags(&status, &change)
			status.Files = append(status.Files, change)
		}
	}

	sort.Slice(status.Files, func(i int, j int) bool {
		return status.Files[i].Path < status.Files[j].Path
	})

	if status.Branch == "" {
		status.Branch = "HEAD"
	}

	return status, nil
}

func parseOrdinaryLine(line string) (FileChange, error) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return FileChange{}, fmt.Errorf("unexpected status entry: %s", line)
	}

	path := strings.Join(fields[8:], " ")
	staged, unstaged := parseXY(fields[1])
	change := FileChange{
		Path:                path,
		DisplayPath:         path,
		StagedStatus:        staged,
		UnstagedStatus:      unstaged,
		IsChangedAfterStage: staged != StatusUnmodified && unstaged != StatusUnmodified,
		IsDeleted:           staged == StatusDeleted || unstaged == StatusDeleted,
	}
	return change, nil
}

func parseRenameLine(line string) (FileChange, error) {
	parts := strings.SplitN(line, "\t", 2)
	fields := strings.Fields(parts[0])
	if len(fields) < 10 {
		return FileChange{}, fmt.Errorf("unexpected rename status entry: %s", line)
	}

	originalPath := ""
	if len(parts) == 2 {
		originalPath = parts[1]
	}

	path := fields[9]
	staged, unstaged := parseXY(fields[1])
	change := FileChange{
		Path:                path,
		DisplayPath:         path,
		OriginalPath:        originalPath,
		StagedStatus:        staged,
		UnstagedStatus:      unstaged,
		IsChangedAfterStage: staged != StatusUnmodified && unstaged != StatusUnmodified,
		IsDeleted:           staged == StatusDeleted || unstaged == StatusDeleted,
		IsRenamed:           staged == StatusRenamed || unstaged == StatusRenamed,
	}
	return change, nil
}

func parseUnmergedLine(line string) (FileChange, error) {
	fields := strings.Fields(line)
	if len(fields) < 11 {
		return FileChange{}, fmt.Errorf("unexpected unmerged status entry: %s", line)
	}
	path := strings.Join(fields[10:], " ")
	return FileChange{
		Path:           path,
		DisplayPath:    path,
		StagedStatus:   StatusConflicted,
		UnstagedStatus: StatusConflicted,
	}, nil
}

func parseXY(xy string) (FileStatus, FileStatus) {
	if len(xy) < 2 {
		return StatusUnmodified, StatusUnmodified
	}

	return statusFromCode(xy[0]), statusFromCode(xy[1])
}

func statusFromCode(code byte) FileStatus {
	switch code {
	case '.', ' ':
		return StatusUnmodified
	case 'M':
		return StatusModified
	case 'A':
		return StatusAdded
	case 'D':
		return StatusDeleted
	case 'R':
		return StatusRenamed
	case 'C':
		return StatusCopied
	case 'U':
		return StatusConflicted
	case '?':
		return StatusUntracked
	default:
		return StatusUnmodified
	}
}

func applyChangeFlags(status *RepoStatus, change *FileChange) {
	status.HasUncommitted = true
	if change.HasStaged() {
		status.HasStaged = true
	}
	if change.HasUnstaged() {
		status.HasUnstaged = true
	}
	if change.IsUntracked {
		status.HasUntracked = true
	}
}
