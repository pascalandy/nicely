# D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions. For the first release, Homebrew installs them and the Linux archive ships them. The AUR package will install them when AUR registration reopens. Completion suggests values such as skill names and profiles, and [D055](D055-completion-values-from-a-list-command.md) decides how it gets those that only an extension knows.

**Why.** Completion is how a terminal user discovers flags and values. Cobra builds it from the command tree, so it never drifts from the commands.

**Rejected.** Hand-written completion scripts.
