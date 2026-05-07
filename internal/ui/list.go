package ui

import (
	"fmt"
	"path/filepath"

	"github.com/agagnon/worktree-cli/internal/config"
	"github.com/agagnon/worktree-cli/internal/git"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	wt git.Worktree
}

func (i item) Title() string       { return filepath.Base(i.wt.Path) }
func (i item) Description() string { return fmt.Sprintf("%s · %s", i.wt.Branch, i.wt.Path) }
func (i item) FilterValue() string { return i.wt.Path + " " + i.wt.Branch }

func buildList() (list.Model, error) {
	trees, err := git.List()
	if err != nil {
		return list.Model{}, err
	}
	items := make([]list.Item, 0, len(trees))
	for _, t := range trees {
		items = append(items, item{wt: t})
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Worktrees"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return l, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	switch {
	case key.Matches(msg, listBinds.Quit):
		return m, tea.Quit
	case key.Matches(msg, listBinds.Help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil
	case key.Matches(msg, listBinds.Enter):
		if it, ok := m.list.SelectedItem().(item); ok {
			m.selected = it.wt.Path
			return m, tea.Quit
		}
	case key.Matches(msg, listBinds.New):
		m.mode = modeCreate
		m.err = nil
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	case key.Matches(msg, listBinds.Delete):
		if it, ok := m.list.SelectedItem().(item); ok {
			wt := it.wt
			m.pendingWt = &wt
			m.mode = modeDelete
			m.err = nil
		}
		return m, nil
	case key.Matches(msg, listBinds.Edit):
		if it, ok := m.list.SelectedItem().(item); ok {
			m.err = nil
			return m, openEditor(it.wt.Path)
		}
		return m, nil
	case key.Matches(msg, listBinds.Tmux):
		if it, ok := m.list.SelectedItem().(item); ok {
			if err := openTmux(it.wt.Path); err != nil {
				m.err = err
			} else {
				m.err = nil
			}
		}
		return m, nil
	case key.Matches(msg, listBinds.TmuxRun):
		if _, ok := m.list.SelectedItem().(item); ok {
			m.mode = modeTmuxMenu
			m.err = nil
		}
		return m, nil
	case key.Matches(msg, listBinds.Config):
		path, err := config.EnsureExists()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.err = nil
		return m, openConfig(path)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) viewList() string {
	body := m.list.View()
	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body, errStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.help.View(listBinds))
}
