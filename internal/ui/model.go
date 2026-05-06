package ui

import (
	"fmt"
	"path/filepath"

	"github.com/agagnon/worktree-cli/internal/git"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	list     list.Model
	help     help.Model
	width    int
	height   int
	selected string
	err      error
}

type item struct {
	wt git.Worktree
}

func (i item) Title() string       { return filepath.Base(i.wt.Path) }
func (i item) Description() string { return fmt.Sprintf("%s · %s", i.wt.Branch, i.wt.Path) }
func (i item) FilterValue() string { return i.wt.Path + " " + i.wt.Branch }

func New() (Model, error) {
	trees, err := git.List()
	if err != nil {
		return Model{}, err
	}

	items := make([]list.Item, 0, len(trees))
	for _, t := range trees {
		items = append(items, item{wt: t})
	}

	d := list.NewDefaultDelegate()
	l := list.New(items, d, 0, 0)
	l.Title = "Worktrees"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return Model{
		list: l,
		help: help.New(),
	}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Selected() string { return m.selected }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-3)
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, keys.Enter):
			if it, ok := m.list.SelectedItem().(item); ok {
				m.selected = it.wt.Path
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return errStyle.Render(m.err.Error())
	}
	helpView := m.help.View(keys)
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), helpView)
}

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
