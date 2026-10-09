# D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions. For v0.0.1, Homebrew installs them and the Linux archive ships them. The AUR package will install them when AUR registration reopens. Completion suggests values such as skill names and profiles.

**Why.** Completion is how a terminal user discovers flags and values. Cobra builds it from the command tree, so it never drifts from the commands.

**Rejected.** Hand-written completion scripts.
