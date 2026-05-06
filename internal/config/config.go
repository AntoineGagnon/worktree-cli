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
}

type Default struct {
	Pattern string `toml:"pattern"`
}

type Repo struct {
	Pattern string `toml:"pattern"`
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
