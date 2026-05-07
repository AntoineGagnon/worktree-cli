package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type editorDoneMsg struct {
	err          error
	reloadConfig bool
}

func openEditor(path string) tea.Cmd {
	editor := resolveEditor()
	cmd := exec.Command(editor, path)
	cmd.Dir = path
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

func openConfig(path string) tea.Cmd {
	editor := resolveEditor()
	cmd := exec.Command(editor, path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorDoneMsg{err: err, reloadConfig: true}
	})
}

func resolveEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vim"
}

func openTmux(path string) error {
	return runTmuxSplit(path, "")
}

func openTmuxCmd(path, command string) error {
	return runTmuxSplit(path, command)
}

func runTmuxSplit(path, command string) error {
	if os.Getenv("TMUX") == "" {
		return errors.New("not inside a tmux session")
	}
	args := []string{"split-window", "-c", path}
	if command != "" {
		args = append(args, command)
	}
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux: %w: %s", err, out)
	}
	return nil
}
