package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agagnon/worktree-cli/internal/config"
	"github.com/agagnon/worktree-cli/internal/git"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type editorDoneMsg struct{ err error }

type mode int

const (
	modeList mode = iota
	modeCreate
	modeDelete
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

type item struct {
	wt git.Worktree
}

func (i item) Title() string       { return filepath.Base(i.wt.Path) }
func (i item) Description() string { return fmt.Sprintf("%s · %s", i.wt.Branch, i.wt.Path) }
func (i item) FilterValue() string { return i.wt.Path + " " + i.wt.Branch }

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
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeCreate:
			return m.updateCreate(msg)
		case modeDelete:
			return m.updateDelete(msg)
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
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

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
		if err := git.Add(path, branch); err != nil {
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
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

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

func (m Model) View() string {
	switch m.mode {
	case modeCreate:
		return m.viewCreate()
	case modeDelete:
		return m.viewDelete()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	body := m.list.View()
	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body, errStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.help.View(listBinds))
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

func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, path)
	cmd.Dir = path
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

func openTmux(path string) error {
	if os.Getenv("TMUX") == "" {
		return errors.New("not inside a tmux session")
	}
	cmd := exec.Command("tmux", "split-window", "-c", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux: %w: %s", err, out)
	}
	return nil
}

var (
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	previewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)
