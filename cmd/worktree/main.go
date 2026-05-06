package main

import (
	"fmt"
	"io"
	"os"

	"github.com/agagnon/worktree-cli/internal/shell"
	"github.com/agagnon/worktree-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "shell":
			fmt.Print(shell.Script)
			return
		case "-h", "--help", "help":
			printHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			printHelp()
			os.Exit(2)
		}
	}

	runTUI()
}

func printHelp() {
	fmt.Print(`worktree — interactive git worktree manager

Usage:
  worktree            Launch the interactive TUI (alias: wt)
  worktree shell      Print shell integration (use: eval "$(worktree shell)")
  worktree help       Show this help

Inside the TUI:
  enter   navigate to selected worktree
  n       new worktree
  d       delete worktree (y confirm · f force · n cancel)
  e       open in $EDITOR
  t       open in a new tmux pane (must be inside tmux)
  ?       toggle help
  q       quit
`)
}

func runTUI() {
	// Render to /dev/tty so the wrapper can capture stdout for the selected
	// path without breaking the TUI or color detection. Fall back to stderr.
	var out io.Writer = os.Stderr
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer tty.Close()
		out = tty
	}
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(out))

	m, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(out))
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
