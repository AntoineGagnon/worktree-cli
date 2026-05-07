package ui

import (
	"fmt"

	"github.com/agagnon/worktree-cli/internal/config"
	"github.com/agagnon/worktree-cli/internal/git"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeList mode = iota
	modeCreate
	modeDelete
	modeTmuxMenu
)

type Model struct {
	mode      mode
	list      list.Model
	input     textinput.Model
	help      help.Model
	cfg       *config.Config
	repoRoot  string
	width     int
	height    int
	selected  string
	pendingWt *git.Worktree
	err       error
}

func New() (Model, error) {
	root, err := git.RepoRoot()
	if err != nil {
		return Model{}, err
	}

	cfg, err := config.Load()
	if err != nil {
		return Model{}, fmt.Errorf("load config: %w", err)
	}

	l, err := buildList()
	if err != nil {
		return Model{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "branch-name"
	ti.Prompt = "branch › "
	ti.CharLimit = 200

	return Model{
		list:     l,
		input:    ti,
		help:     help.New(),
		cfg:      cfg,
		repoRoot: root,
	}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Selected() string { return m.selected }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-3)
		m.input.Width = msg.Width - 12
		m.help.Width = msg.Width
		return m, nil

	case editorDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		if msg.reloadConfig {
			if cfg, err := config.Load(); err != nil {
				m.err = fmt.Errorf("reload config: %w", err)
			} else {
				m.cfg = cfg
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeCreate:
			return m.updateCreate(msg)
		case modeDelete:
			return m.updateDelete(msg)
		case modeTmuxMenu:
			return m.updateTmuxMenu(msg)
		default:
			return m.updateList(msg)
		}
	}

	if m.mode == modeCreate {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	switch m.mode {
	case modeCreate:
		return m.viewCreate()
	case modeDelete:
		return m.viewDelete()
	case modeTmuxMenu:
		return m.viewTmuxMenu()
	default:
		return m.viewList()
	}
}
