# D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions. Homebrew installs them; Linux archives ship them. AUR installation waits for [COR-105](../../backlog/COR-105-aur-package.md). [D055](D055-completion-values-from-a-list-command.md) defines how extensions supply dynamic values.

**Why.** Generated completion follows the command tree and helps users discover flags and values.

**Rejected.** Hand-written completion scripts.
