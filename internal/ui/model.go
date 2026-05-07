package ui

import (
	"fmt"

	"github.com/agagnon/worktree-cli/internal/config"
	"github.com/agagnon/worktree-cli/internal/git"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	spinner   spinner.Model
	cfg       *config.Config
	repoRoot  string
	width     int
	height    int
	selected  string
	pendingWt *git.Worktree
	busy      bool
	spinning  bool
	busyMsg   string
	busyHint  string
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

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	return Model{
		list:     l,
		input:    ti,
		help:     help.New(),
		spinner:  sp,
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

	case showSpinnerMsg:
		if m.busy {
			m.spinning = true
			return m, m.spinner.Tick
		}
		return m, nil

	case spinner.TickMsg:
		if !m.spinning {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case createDoneMsg:
		m.busy = false
		m.spinning = false
		if msg.err != nil {
			m.err = msg.err
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
		m.input.Blur()
		return m, nil

	case deleteDoneMsg:
		m.busy = false
		m.spinning = false
		if msg.err != nil {
			m.err = msg.err
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
		return m, nil

	case tea.KeyMsg:
		if m.busy {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
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
	if m.busy && m.spinning {
		return m.viewBusy()
	}
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

func (m Model) viewBusy() string {
	header := m.spinner.View() + "  " + titleStyle.Render(m.busyMsg)
	if m.busyHint == "" {
		return header
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		dimStyle.Render("  → "+m.busyHint),
	)
}
