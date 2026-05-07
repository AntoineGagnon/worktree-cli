# worktree-cli

An interactive TUI for managing git worktrees, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). List, create, delete, and `cd` into worktrees with single keystrokes; spawn editor or tmux panes scoped to a worktree; configure path patterns per-repo.

## Install

Clone and install the binary:

```sh
git clone git@github.com:AntoineGagnon/worktree-cli.git
cd worktree-cli
go install ./cmd/worktree
```

This puts a `worktree` binary in `$GOBIN` (or `~/go/bin` if `GOBIN` is unset). Make sure that directory is on your `$PATH`.

Add the shell integration so `enter` in the TUI actually changes your shell's directory:

```sh
echo 'eval "$(worktree shell)"' >> ~/.zshrc
source ~/.zshrc
```

You can now run it as `worktree` or the shorter `wt`.

## Usage

Run inside any git repository:

```sh
worktree     # or: wt
```

| Key     | Action                                                   |
| ------- | -------------------------------------------------------- |
| `enter` | `cd` into the selected worktree                          |
| `n`     | new worktree (prompts for branch name)                   |
| `d`     | delete worktree (`y` confirm · `f` force · `n` cancel)   |
| `e`     | open the worktree in `$EDITOR`                           |
| `t`     | open a new tmux pane in the worktree                     |
| `T`     | "Run commands" menu — pick a configured command          |
| `c`     | open the config file in `$EDITOR` (template if missing)  |
| `/`     | filter the list                                          |
| `?`     | toggle full help                                         |
| `q`     | quit                                                     |

## Configuration

`~/.config/worktree-cli/config.toml` (or `$XDG_CONFIG_HOME/worktree-cli/config.toml`). Press `c` from the TUI to edit it — a commented template is generated on first open.

```toml
[default]
pattern = "../{repo}-{branch}"

[repos."/Users/you/src/example"]
pattern = "~/src/worktrees/example/{branch}"

[[tmux.commands]]
key = "c"
label = "Claude"
command = "claude"
```

Pattern variables: `{repo}` (basename of the repo root) and `{branch}`. Leading `~` expands to `$HOME`; relative paths resolve against the repo root.

## Requirements

- Go 1.21+ (build only; not needed at runtime)
- zsh or bash for the shell wrapper
- `tmux` (optional, for `t` / `T`)
