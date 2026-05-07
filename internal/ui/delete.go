package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingWt == nil {
		m.mode = modeList
		return m, nil
	}
	switch {
	case key.Matches(msg, deleteBinds.Cancel):
		m.mode = modeList
		m.pendingWt = nil
		m.err = nil
		return m, nil
	case key.Matches(msg, deleteBinds.Confirm):
		return m.runDelete(false)
	case key.Matches(msg, deleteBinds.Force):
		return m.runDelete(true)
	}
	return m, nil
}

func (m Model) runDelete(force bool) (tea.Model, tea.Cmd) {
	m.busy = true
	m.spinning = false
	m.busyMsg = "Deleting worktree…"
	m.busyHint = m.pendingWt.Path
	m.err = nil
	return m, tea.Batch(
		deleteWorktreeCmd(m.pendingWt.Path, force),
		showSpinnerAfter(200*time.Millisecond),
	)
}

func (m Model) viewDelete() string {
	wt := m.pendingWt
	header := warnStyle.Render("Delete worktree?")
	name := titleStyle.Render(filepath.Base(wt.Path))
	branch := dimStyle.Render(fmt.Sprintf("branch: %s", wt.Branch))
	path := dimStyle.Render(wt.Path)

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		name,
		branch,
		path,
	)
	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body, "", errStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, "", m.help.View(deleteBinds))
}
