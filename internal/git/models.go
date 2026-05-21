package git

type FileStatus string

const (
	StatusUnmodified FileStatus = "unmodified"
	StatusModified   FileStatus = "modified"
	StatusAdded      FileStatus = "added"
	StatusDeleted    FileStatus = "deleted"
	StatusRenamed    FileStatus = "renamed"
	StatusCopied     FileStatus = "copied"
	StatusUntracked  FileStatus = "untracked"
	StatusConflicted FileStatus = "conflicted"
)

type RepoStatus struct {
	Branch         string
	Ahead          int
	Behind         int
	HasUncommitted bool
	HasStaged      bool
	HasUnstaged    bool
	HasUntracked   bool
	Files          []FileChange
}

type FileChange struct {
	Path                string
	DisplayPath         string
	OriginalPath        string
	StagedStatus        FileStatus
	UnstagedStatus      FileStatus
	IsChangedAfterStage bool
	IsUntracked         bool
	IsDeleted           bool
	IsRenamed           bool
}

type CommitHistoryEntry struct {
	Hash      string
	Author    string
	Subject   string
	FullPatch string
}

func (c FileChange) HasStaged() bool {
	return c.StagedStatus != "" && c.StagedStatus != StatusUnmodified
}

func (c FileChange) HasUnstaged() bool {
	return c.UnstagedStatus != "" && c.UnstagedStatus != StatusUnmodified
}

func (c FileChange) StagedStatusLabel() string {
	if c.StagedStatus == "" || c.StagedStatus == StatusUnmodified {
		return "staged"
	}
	return string(c.StagedStatus)
}

func (c FileChange) UnstagedStatusLabel() string {
	if c.UnstagedStatus == "" || c.UnstagedStatus == StatusUnmodified {
		return "modified"
	}
	return string(c.UnstagedStatus)
}
