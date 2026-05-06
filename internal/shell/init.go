package shell

const Script = `# worktree-cli shell integration
# Install: eval "$(worktree shell)"
worktree() {
  if [ $# -eq 0 ]; then
    local __wt_out
    __wt_out=$(command worktree) || return $?
    if [ -n "$__wt_out" ] && [ -d "$__wt_out" ]; then
      cd "$__wt_out" || return $?
    fi
  else
    command worktree "$@"
  fi
}
wt() { worktree "$@"; }
`
