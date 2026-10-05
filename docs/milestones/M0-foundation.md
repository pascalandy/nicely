# M0 Foundation

Status: planned
Release: v0.0.1

## Goal

An installable `ncly` that does almost nothing yet, built on finished plumbing. Every later command lands on the same contract, catalogs, config, tests, and release pipeline, and no later milestone has to reshape them.

## Depends on

Nothing.

## Prerequisites

Pascal does these before T6. T1 to T5 do not need them.

- [ ] The empty repository `pascalandy/homebrew-tap` exists
- [ ] An AUR account exists, with an SSH key registered for it
- [ ] The `pascalandy/nicely` repository has two secrets for `release.yml`: a token that can push to `pascalandy/homebrew-tap`, and the private half of the AUR SSH key
- [ ] After T1, `main` requires the status check `signoff`

## Tasks

Do the tasks in this order. [AGENTS.md](../../AGENTS.md#code) gives the layout.

- [ ] **T1 Skeleton.** `go.mod` with the module path `github.com/pascalandy/nicely`, `cmd/ncly`, and `--version`. A `justfile` with `check`, `test`, `lint`, and `signoff`. Lefthook runs format, lint, and gitleaks before a commit, and tests and scenarios before a push. The testscript runner with the `exits` command and `HOME` and the XDG folders inside `$WORK`. Go tools such as `govulncheck` and `go-licenses` are pinned with `tool` lines in `go.mod`.
- [ ] **T2 Contract.** `internal/contract` implements the M0 sections of [cli-spec.md](../north-star/cli-spec.md): Output, Exit codes, Error codes, Modes, and the global flags with their variables. The root command sends an unknown name to an extension registry that holds nothing yet and skips reserved names. Help and completion read the same registry, so M3 fills it without reshaping the root. A test fails when cli-spec.md and the code disagree on the error registry, the exit codes, the global flags, or the command tree.
- [ ] **T3 Config.** The shared and local config files, their precedence, and the XDG paths, including state. `CONFIG_INVALID` for a broken file. Unknown keys are collected for the `CONFIG_UNKNOWN_KEY` warning that doctor reports in M1.
- [ ] **T4 Text.** The English catalog, with a description per entry, and the parity test against `en`. The pseudo-locale `en-XA`, with scenarios for the help and for an error. Plural rules, number, size, and date formats, and language matching in `internal/i18n`. Help text comes from the catalog, so `--lang` is read before the command tree is built. Lip Gloss styles and the error block in `internal/tui`.
- [ ] **T5 Completion.** `ncly completion zsh|bash|fish`.
- [ ] **T6 Release.** GoReleaser builds macOS and Linux binaries for arm64 and amd64. Each archive ships the third-party notices that `go-licenses` generates. The AUR package `ncly-bin` installs the binary and its completions. The Homebrew formula builds `ncly` from the release's source archive and generates its completions. `release.yml` starts only through `workflow_dispatch`, runs GoReleaser, and updates the formula. `just release-check` runs `govulncheck` and `goreleaser check` on top of `just check`.
- [ ] **T7 First release.** v0.0.1 is tagged and released through `release.yml`, and both manual checks pass.

## Out of scope

- Every user command other than `--help`, `--version`, and `completion`.
- Extension discovery, beyond the empty extension registry of T2: M3.
- The form, spinner, progress bar, and step list of `internal/tui`. Each one lands with the first command that needs it.

## Acceptance

```
# The version goes to stdout and nothing goes to stderr
exec ncly --version
stdout '^ncly v'
! stderr .

# An unknown command fails as one JSON line
exits 2 ncly nope --json
! stdout .
stderr '^\{"ok":false,"errors":\[\{"code":"USAGE_INVALID"'

# A language without a catalog falls back to English
exec ncly --lang xx --help
stdout '(?i)usage'

# The pseudo-locale marks the catalog strings of the help and of an error
exec ncly --lang en-XA --help
stdout '⟦'
exits 2 ncly --lang en-XA nope
stderr '⟦'

# A broken config never blocks the help
env NCLY_CONFIG=$WORK/broken.toml
exec ncly --help
stdout '(?i)usage'

# NO_COLOR removes escape codes
env NO_COLOR=1
exec ncly --help
! stdout '\x1b\['

-- broken.toml --
lang =
```

T4 sharpens the pseudo-locale scenario, so that it fails on any word outside the markers other than command names, flags, and examples.

Manual checks on Pascal's machines:

1. On macOS, `brew install pascalandy/tap/ncly` builds `ncly`, then `ncly <Tab>` completes in zsh.
2. On Omarchy, `yay -S ncly-bin`, then `ncly <Tab>` completes in bash.

## Done when

- [ ] Every prerequisite and every task is ticked
- [ ] Every acceptance scenario passes in `just check`
- [ ] Both manual checks pass
- [ ] v0.0.1 is tagged and released through `release.yml`
