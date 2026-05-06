package main

import (
	"fmt"
	"os"

	"github.com/agagnon/worktree-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if fm, ok := final.(ui.Model); ok {
		if path := fm.Selected(); path != "" {
			fmt.Println(path)
		}
	}
}
