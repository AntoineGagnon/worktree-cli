package ui

import (
	"fmt"
	"time"

	"github.com/agagnon/worktree-cli/internal/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, createBinds.Cancel):
		m.mode = modeList
		m.input.Blur()
		m.err = nil
		return m, nil
	case key.Matches(msg, createBinds.Confirm):
		branch := m.input.Value()
		if branch == "" {
			return m, nil
		}
		path, err := config.Resolve(m.cfg.PatternFor(m.repoRoot), m.repoRoot, branch)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.busy = true
		m.spinning = false
		m.busyMsg = "Creating worktree…"
		m.busyHint = path
		m.err = nil
		return m, tea.Batch(
			createWorktreeCmd(path, branch),
			showSpinnerAfter(200*time.Millisecond),
		)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) viewCreate() string {
	pattern := m.cfg.PatternFor(m.repoRoot)
	preview := "(enter a branch name)"
	if v := m.input.Value(); v != "" {
		if path, err := config.Resolve(pattern, m.repoRoot, v); err == nil {
			preview = path
		}
	}

	header := titleStyle.Render("New worktree")
	patternLine := dimStyle.Render(fmt.Sprintf("pattern: %s", pattern))
	previewLine := dimStyle.Render("→ ") + previewStyle.Render(preview)

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.input.View(),
		"",
		patternLine,
		previewLine,
	)
	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body, "", errStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, "", m.help.View(createBinds))
}
