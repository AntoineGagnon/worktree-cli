package ui

import (
	"fmt"
	"path/filepath"

	"github.com/agagnon/worktree-cli/internal/git"
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
	if err := git.Remove(m.pendingWt.Path, force); err != nil {
		m.err = err
		return m, nil
	}
	l, err := buildList()
	if err != nil {
		m.err = err
		return m, nil
	}
	l.SetSize(m.width, m.height-3)
	m.list = l
	m.mode = modeList
	m.pendingWt = nil
	m.err = nil
	return m, nil
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
