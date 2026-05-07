package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultPattern = "../{repo}-{branch}"

type Config struct {
	Default Default         `toml:"default"`
	Repos   map[string]Repo `toml:"repos"`
	Tmux    Tmux            `toml:"tmux"`
}

type Default struct {
	Pattern string `toml:"pattern"`
}

type Repo struct {
	Pattern string `toml:"pattern"`
}

type Tmux struct {
	Commands []TmuxCommand `toml:"commands"`
}

type TmuxCommand struct {
	Key     string `toml:"key"`
	Label   string `toml:"label"`
	Command string `toml:"command"`
}

const Template = `# worktree-cli config
# Uncomment and edit the sections you want to use.

# Default path pattern for new worktrees.
# Variables: {repo} (basename of repo root), {branch}.
# Leading ~ expands to $HOME; relative paths resolve against the repo root.
# [default]
# pattern = "../{repo}-{branch}"

# Per-repo overrides, keyed by absolute path of the repo root.
# [repos."/Users/you/src/example"]
# pattern = "~/src/worktrees/example/{branch}"

# Commands available from the T (Run commands) menu.
# Each entry maps a single key to a shell command, run in a new tmux pane
# whose cwd is the worktree. The pane closes when the command exits — append
# "; exec $SHELL" if you want to drop into a shell after.
# [[tmux.commands]]
# key = "c"
# label = "Claude"
# command = "claude"
#
# [[tmux.commands]]
# key = "v"
# label = "Neovim"
# command = "nvim ."
`

// EnsureExists creates the config file with a commented template if it does
// not already exist. Returns the resolved path either way.
func EnsureExists() (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(Template), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func Path() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "worktree-cli", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "worktree-cli", "config.toml"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	cfg := &Config{Repos: map[string]Repo{}}
	_, err = toml.DecodeFile(path, cfg)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if cfg.Repos == nil {
		cfg.Repos = map[string]Repo{}
	}
	return cfg, nil
}

func (c *Config) PatternFor(repoRoot string) string {
	if r, ok := c.Repos[repoRoot]; ok && r.Pattern != "" {
		return r.Pattern
	}
	if c.Default.Pattern != "" {
		return c.Default.Pattern
	}
	return DefaultPattern
}

// Resolve expands a pattern into an absolute path.
// Variables: {repo} (basename of repoRoot), {branch}.
// Leading ~ expands to $HOME. Relative paths resolve against repoRoot.
func Resolve(pattern, repoRoot, branch string) (string, error) {
	p := pattern
	p = strings.ReplaceAll(p, "{repo}", filepath.Base(repoRoot))
	p = strings.ReplaceAll(p, "{branch}", branch)

	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}

	if !filepath.IsAbs(p) {
		p = filepath.Join(repoRoot, p)
	}
	return filepath.Clean(p), nil
}
