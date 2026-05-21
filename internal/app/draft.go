package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type draftData struct {
	CommitTitle string `json:"commit_title"`
	CommitBody  string `json:"commit_body"`
	PRTitle     string `json:"pr_title"`
	PRBody      string `json:"pr_body"`
}

func (d draftData) empty() bool {
	return d.CommitTitle == "" && d.CommitBody == "" && d.PRTitle == "" && d.PRBody == ""
}

func draftPath(repoRoot string) string {
	h := sha256.Sum256([]byte(repoRoot))
	return filepath.Join(os.TempDir(), fmt.Sprintf("gitscribe-draft-%x.json", h[:8]))
}

func saveDraft(repoRoot string, data draftData) {
	if data.empty() {
		_ = os.Remove(draftPath(repoRoot))
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	_ = os.WriteFile(draftPath(repoRoot), b, 0600)
}

func loadAndDeleteDraft(repoRoot string) (draftData, bool) {
	path := draftPath(repoRoot)
	b, err := os.ReadFile(path)
	if err != nil {
		return draftData{}, false
	}
	_ = os.Remove(path)
	var data draftData
	if err := json.Unmarshal(b, &data); err != nil {
		return draftData{}, false
	}
	return data, !data.empty()
}

func (m *Model) RestoreDraft() {
	draft, ok := loadAndDeleteDraft(m.repo.Root)
	if !ok {
		return
	}
	m.commitForm.Title.SetValue(draft.CommitTitle)
	m.commitForm.Body.SetValue(draft.CommitBody)
	m.prForm.Title.SetValue(draft.PRTitle)
	m.prForm.Body.SetValue(draft.PRBody)
}

func (m *Model) persistDraft() {
	saveDraft(m.repo.Root, draftData{
		CommitTitle: m.commitForm.Title.Value(),
		CommitBody:  m.commitForm.Body.Value(),
		PRTitle:     m.prForm.Title.Value(),
		PRBody:      m.prForm.Body.Value(),
	})
}
