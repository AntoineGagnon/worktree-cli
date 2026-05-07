package ui

import (
	"fmt"

	"github.com/agagnon/worktree-cli/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateTmuxMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "ctrl+c" {
		m.mode = modeList
		m.err = nil
		return m, nil
	}
	for _, c := range m.cfg.Tmux.Commands {
		if c.Key == msg.String() {
			it, ok := m.list.SelectedItem().(item)
			if !ok {
				m.mode = modeList
				return m, nil
			}
			if err := openTmuxCmd(it.wt.Path, c.Command); err != nil {
				m.err = err
			} else {
				m.err = nil
			}
			m.mode = modeList
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewTmuxMenu() string {
	header := titleStyle.Render("Run in new tmux pane")

	var body string
	if len(m.cfg.Tmux.Commands) == 0 {
		path, _ := config.Path()
		body = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			dimStyle.Render("No commands configured."),
			dimStyle.Render("Add to "+path+":"),
			"",
			dimStyle.Render(`  [[tmux.commands]]`),
			dimStyle.Render(`  key = "c"`),
			dimStyle.Render(`  label = "Claude"`),
			dimStyle.Render(`  command = "claude"`),
		)
	} else {
		lines := []string{header, ""}
		for _, c := range m.cfg.Tmux.Commands {
			lines = append(lines, fmt.Sprintf("  %s  %s  %s",
				keyStyle.Render(c.Key),
				c.Label,
				dimStyle.Render("→ "+c.Command),
			))
		}
		body = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body, "", errStyle.Render(m.err.Error()))
	}
	hint := dimStyle.Render("press a key · esc to cancel")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", hint)
}
