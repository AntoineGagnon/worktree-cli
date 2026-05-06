package ui

import "github.com/charmbracelet/bubbles/key"

type listKeys struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	New    key.Binding
	Delete key.Binding
	Edit   key.Binding
	Tmux   key.Binding
	Help   key.Binding
	Quit   key.Binding
}

func (k listKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.New, k.Delete, k.Edit, k.Tmux, k.Help, k.Quit}
}

func (k listKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Enter, k.New, k.Delete},
		{k.Edit, k.Tmux},
		{k.Help, k.Quit},
	}
}

type createKeys struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func (k createKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

func (k createKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Cancel}}
}

var listBinds = listKeys{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "go to")),
	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Tmux:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tmux")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

var createBinds = createKeys{
	Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create")),
	Cancel:  key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "cancel")),
}

type deleteKeys struct {
	Confirm key.Binding
	Force   key.Binding
	Cancel  key.Binding
}

func (k deleteKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Force, k.Cancel}
}

func (k deleteKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Force, k.Cancel}}
}

var deleteBinds = deleteKeys{
	Confirm: key.NewBinding(key.WithKeys("y", "enter"), key.WithHelp("y", "confirm")),
	Force:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "force")),
	Cancel:  key.NewBinding(key.WithKeys("n", "esc", "ctrl+c"), key.WithHelp("n/esc", "cancel")),
}
