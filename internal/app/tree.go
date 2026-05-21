package app

import (
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

func (m *Model) setDiff(content string, mode string) {
	m.diffContent = content
	m.diffModeLabel = mode
	if strings.TrimSpace(content) == "" {
		content = "Select a file to view its diff."
	}
	m.diffViewport.SetContent(m.colorizeDiff(content))
	m.diffViewport.GotoTop()
}

func (m *Model) resize() {
	_, rightWidth := mainPanelWidths(m.width)
	panelHeight := max(10, m.height-5)
	innerHeight := max(5, panelHeight-2)

	m.diffViewport.Width = max(20, rightWidth-4)
	m.diffViewport.Height = innerHeight
	m.historyViewport.Width = max(20, rightWidth-4)
	m.historyViewport.Height = innerHeight
	m.stashViewport.Width = max(20, rightWidth-4)
	m.stashViewport.Height = innerHeight
	m.modal.viewport.Width = max(24, min(102, m.width-14))
	m.modal.viewport.Height = max(8, min(m.height-14, m.height/2))

	formWidth := max(56, int(float64(m.width)*0.54))
	m.commitForm.Title.Width = formWidth - 2
	m.prForm.Title.Width = formWidth - 2
	m.commitForm.Body.SetWidth(formWidth - 2)
	m.prForm.Body.SetWidth(formWidth - 2)
	m.commitForm.Body.SetHeight(max(6, m.height/4-3))
	m.prForm.Body.SetHeight(max(6, m.height/4-3))
}

func (m *Model) rebuildTree() {
	selectedPath := ""
	if change, ok := m.selectedFileChange(); ok {
		selectedPath = change.Path
	} else if m.tree.Index >= 0 && m.tree.Index < len(m.tree.Rows) {
		selectedPath = m.tree.Rows[m.tree.Index].Path
	}

	rows := buildTreeRows(filepath.Base(m.repo.Root), m.status.RepoStatus.Files)
	m.tree.Rows = rows
	m.tree.Index = 0
	for i, row := range rows {
		if row.Path == selectedPath {
			m.tree.Index = i
			break
		}
	}
}

func (m *Model) selectedFileChange() (*git.FileChange, bool) {
	if m.tree.Index < 0 || m.tree.Index >= len(m.tree.Rows) {
		return nil, false
	}
	row := m.tree.Rows[m.tree.Index]
	if row.Change == nil {
		return nil, false
	}
	return row.Change, true
}

func (m *Model) selectedTreeRow() (treeRow, bool) {
	if m.tree.Index < 0 || m.tree.Index >= len(m.tree.Rows) {
		return treeRow{}, false
	}
	return m.tree.Rows[m.tree.Index], true
}

func (m *Model) loadSelectionDiffCmd() tea.Cmd {
	row, ok := m.selectedTreeRow()
	if !ok {
		m.setDiff("", "")
		return nil
	}
	if row.IsDir {
		return loadDirectoryDiffCmd(m.repo.Root, row.Path)
	}
	if row.Change == nil {
		m.setDiff("", "")
		return nil
	}
	return loadDiffCmd(m.repo.Root, m.cfg.Git, row.Path, *row.Change)
}

func (m *Model) setCurrentBranchIndex() {
	for i, branch := range m.branches {
		if branch == m.repo.Branch {
			m.branchIndex = i
			return
		}
	}
	m.branchIndex = 0
}

func (m *Model) changesUnderPath(path string) []git.FileChange {
	if path == "." {
		return append([]git.FileChange(nil), m.status.RepoStatus.Files...)
	}
	prefix := path + "/"
	changes := make([]git.FileChange, 0)
	for _, change := range m.status.RepoStatus.Files {
		if change.Path == path || strings.HasPrefix(change.Path, prefix) {
			changes = append(changes, change)
		}
	}
	return changes
}

func buildTreeRows(rootName string, files []git.FileChange) []treeRow {
	root := &treeNode{Name: rootName, Path: ".", IsDir: true, Children: map[string]*treeNode{}}
	for _, file := range files {
		parts := strings.Split(file.Path, "/")
		current := root
		accum := ""
		for i, part := range parts {
			if accum == "" {
				accum = part
			} else {
				accum = accum + "/" + part
			}
			child, ok := current.Children[part]
			if !ok {
				child = &treeNode{Name: part, Path: accum, IsDir: i < len(parts)-1, Children: map[string]*treeNode{}}
				current.Children[part] = child
			}
			current = child
		}
		copy := file
		if strings.TrimSpace(copy.DisplayPath) == "" {
			copy.DisplayPath = copy.Path
		}
		current.Change = &copy
	}

	rows := []treeRow{{Path: ".", Label: root.Name + "/", IsDir: true, Level: 0}}
	rows = append(rows, flattenTree(root, 0)...)
	return rows
}

type treeNode struct {
	Name     string
	Path     string
	IsDir    bool
	Change   *git.FileChange
	Children map[string]*treeNode
}

func flattenTree(node *treeNode, level int) []treeRow {
	keys := make([]string, 0, len(node.Children))
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i int, j int) bool {
		left := node.Children[keys[i]]
		right := node.Children[keys[j]]
		if left.IsDir != right.IsDir {
			return left.IsDir
		}
		return left.Name < right.Name
	})

	rows := []treeRow{}
	for _, key := range keys {
		child := node.Children[key]
		label := child.Name
		if child.IsDir {
			label = label + "/"
		}
		row := treeRow{
			Path:   child.Path,
			Label:  label,
			IsDir:  child.IsDir,
			Level:  level + 1,
			Change: child.Change,
		}
		if child.Change != nil && strings.TrimSpace(child.Change.DisplayPath) != "" {
			row.Label = displayName(child.Change.DisplayPath)
		}
		rows = append(rows, row)
		if child.IsDir {
			rows = append(rows, flattenTree(child, level+1)...)
		}
	}
	return rows
}

func indent(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat("  ", level)
}

func shouldUnstageDirectory(changes []git.FileChange) bool {
	hasStagedOnly := false
	for _, change := range changes {
		if change.IsUntracked || change.HasUnstaged() {
			return false
		}
		if change.HasStaged() {
			hasStagedOnly = true
		}
	}
	return hasStagedOnly
}

func changePaths(changes []git.FileChange, unstage bool) []string {
	paths := make([]string, 0, len(changes))
	for _, change := range changes {
		if unstage {
			if !change.HasStaged() {
				continue
			}
		} else if !change.IsUntracked && !change.HasUnstaged() {
			continue
		}
		paths = append(paths, change.Path)
	}
	return paths
}
