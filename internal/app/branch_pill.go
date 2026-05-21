package app

func (m *Model) renderBranchPill() string {
	return m.styles.BranchPill(m.repo.Branch)
}
