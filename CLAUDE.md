# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...                         # compile check
go install ./cmd/worktree              # install `worktree` binary to $GOBIN (or ~/go/bin)
gofmt -w .                             # format
go mod tidy                            # sync deps after import changes
```

There is no test suite yet. Do not invent test commands.

The binary subdirectory is `cmd/worktree/` (not the repo root) so `go install` produces a binary named `worktree` rather than `worktree-cli`.

## Architecture

This is a Bubble Tea TUI for managing git worktrees. Two design choices drive most of the file layout:

### Mode-per-file UI pattern

`internal/ui/model.go` owns the `Model` struct and is **only** a dispatcher: `Update` routes `tea.KeyMsg` to the active mode's `update<Mode>` method, and `View` routes to the active mode's `view<Mode>` method. Each mode (list, create, delete, tmuxmenu) lives in its own file and defines those two methods on `Model` — they share state through the shared receiver, not through interfaces.

To add a new mode:
1. Append a value to the `mode` enum in `model.go`.
2. Create `internal/ui/<mode>.go` with `func (m Model) update<Mode>(msg tea.KeyMsg) (tea.Model, tea.Cmd)` and `func (m Model) view<Mode>() string`.
3. Add a `case` in the `Update` switch (key dispatch) and the `View` switch in `model.go`.
4. Add a binding in `keys.go` and a trigger in `updateList` (or wherever the entry point lives).

`internal/ui/actions.go` holds shell-out helpers and message types — anything that produces a `tea.Cmd` for an external process (editor, tmux split, async git ops) lives there alongside its result `*DoneMsg` type.

### Shell wrapper integration

A child process can't change its parent shell's cwd, so navigation works via:
1. The TUI prints the selected worktree path to **stdout** on exit.
2. The TUI renders to **`/dev/tty`** (see `cmd/worktree/cli.go:runTUI`) — this keeps stdout free for the path AND ensures lipgloss color detection runs against the real terminal even when stdout is a captured pipe. Without this, `eval "$(worktree shell)"`-style integration would either swallow the TUI rendering or strip all ANSI colors.
3. `internal/shell/init.go` contains a zsh-compatible function template printed by the `worktree shell` subcommand. Users install via `eval "$(worktree shell)"`. The function captures stdout via `$(command worktree)` and `cd`s into the result. Both `worktree` and `wt` are exposed.

When changing how the binary writes to stdout/stderr, remember that **anything on stdout is consumed by the wrapper** — only the final selected path should go there.

### Async work + delayed spinner

`git worktree add` / `git worktree remove` block on large monorepos. The pattern in `actions.go` + the create/delete handlers + `model.go`:
- The handler sets `m.busy = true` and returns `tea.Batch(<workCmd>, showSpinnerAfter(200ms))`.
- The work cmd is a goroutine-flavored `tea.Cmd` that returns `createDoneMsg` / `deleteDoneMsg`.
- The 200ms tick fires `showSpinnerMsg`; the dispatcher only flips `m.spinning = true` if still busy. This avoids spinner flash on fast operations.
- All keys except `ctrl+c` are swallowed while `m.busy` (gated in `model.go`'s `Update`).
- The done-msg handler in `model.go` does the post-success list refresh and mode transition — the per-mode files only kick the work off.

Any future long-running operation should follow the same pattern (work cmd + delayed spinner + done msg handled in the dispatcher).

### Subcommand dispatch

`cmd/worktree/main.go` is a one-liner. `cmd/worktree/cli.go` has a `subcommands` slice (name → handler func) consumed by `dispatch`. The default (no args) is the TUI. Adding a subcommand is "append to the slice and write `runFoo`".

### Config

Per-repo TOML config at `$XDG_CONFIG_HOME/worktree-cli/config.toml` (default `~/.config/worktree-cli/config.toml`). Repos are keyed by absolute repo-root path. `config.PatternFor(root)` falls back to `[default].pattern`, then to the hard-coded `DefaultPattern` constant. `config.Resolve` does `{repo}` / `{branch}` / `~` / relative-path expansion. `config.EnsureExists` writes a fully-commented template if the file is missing — the `c` keybind in the TUI uses this to bootstrap first-time users.

## Constraints

- No mention of Claude / Claude Code in commits, PR titles, PR bodies, or anything user-facing. (Inherited from `~/.claude/CLAUDE.md`.)
- Never push without explicit permission.
