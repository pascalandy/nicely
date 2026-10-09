# M00 Foundation

Status: done
Version: v0.0.1

## Demo

An installable `ncly` answers `--help`, `--version`, and `completion`. It fixes the shared contract, catalogs, config, checks, and release pipeline that every later milestone builds on, while each implementation waits for its first consumer.

## Prerequisites

All three were in place before M00-T6. The AUR account and its key `nicely-release-aur` moved to [M99](M99-parking-lot.md), with the AUR publication.

- The empty repository `pascalandy/homebrew-tap` exists
- The `pascalandy/nicely` repository has two secrets for `release.yml`: `HOMEBREW_TAP_SSH_KEY`, the private half of a deploy key with write access to `pascalandy/homebrew-tap`, and `AUR_SSH_KEY`, the private half of the `nicely-release-aur` key
- `main` requires the status check `signoff`

The first release, v0.0.1, waits for Pascal, as [Release](../../AGENTS.md#release) says. It is no card, so it never blocks a milestone.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M00-T1 | Skeleton | agent | — | done |
| M00-T2 | Contract | agent | M00-T1 | done |
| M00-T3 | Config | agent | M00-T2 | done |
| M00-T4 | Text | agent | M00-T2 | done |
| M00-T5 | Completion | agent | M00-T4 | done |
| M00-T6 | Release pipeline | agent | M00-T5 | done |

### M00-T1 Skeleton

- **Read:** [AGENTS.md](../../AGENTS.md#code)
- **Proves:** `testdata/script/version.txtar`, `justfile`, `lefthook.yml`

`go.mod` with the module path `github.com/pascalandy/nicely`, `cmd/ncly`, and `--version`. A `justfile` with `check`, `test`, `lint`, and `signoff`. Lefthook runs format, lint, and gitleaks before a commit, and tests and scenarios before a push. The testscript runner with the `exits` command and `HOME` and the XDG folders inside `$WORK`. Go tools such as `golangci-lint`, `govulncheck`, and `go-licenses` are pinned with `tool` lines in `tools/go.mod`, a separate module, so their dependencies never change the versions that `ncly` ships with.

### M00-T2 Contract

- **Read:** [Output](../north-star/contract.md#output), [Exit codes](../north-star/contract.md#exit-codes), [Command descriptions](../north-star/contract.md#command-descriptions)
- **Proves:** `testdata/script/usage_errors.txtar`, `testdata/script/output_errors.txtar`, `testdata/script/closed_pipe.txtar`, `testdata/script/dry_run_conformance.txtar`, `internal/contract/spec_test.go`

`internal/contract` implements Output, Compatibility, Retry safety, Exit codes, Error codes, Modes, and global flags from [contract.md](../north-star/contract.md). One command declaration supplies help, completion, validation, and the later discovery command. The root dispatches unknown names through an empty extension registry and skips reserved names. The M00 boundaries in [Contract coverage](#contract-coverage) are complete, including independent expected results. Tests check agreement with the spec on errors, exit codes, flags, and command declarations. The Operations responsibilities are fixed, while run records and resume wait for their first consumer.

### M00-T3 Config

- **Read:** [Configuration](../north-star/contract.md#configuration)
- **Proves:** `internal/config/config_test.go`, `testdata/script/language.txtar`

The shared and local config files, their precedence, and the XDG paths, including state. `CONFIG_INVALID` for a broken file. Unknown keys are collected for the `CONFIG_UNKNOWN_KEY` warning that doctor reports.

### M00-T4 Text

- **Read:** [Configuration](../north-star/contract.md#configuration)
- **Proves:** `testdata/script/help.txtar`, `testdata/script/language.txtar`, `testdata/script/error_block.txtar`, `testdata/script/terminal.txtar`, `internal/i18n/i18n_test.go`

The English catalog, with a description per entry, and the parity test against `en`. The pseudo-locale `en-XA`, with scenarios for the help and for an error. Plural rules, number, size, and date formats, and language matching in `internal/i18n`. Help text comes from the catalog, so `--lang` is read before the command tree is built. Styles and the error block in `internal/tui`.

### M00-T5 Completion

- **Read:** [ncly completion](../north-star/contract.md#ncly-completion)
- **Proves:** `testdata/script/completion.txtar`

`ncly completion zsh|bash|fish`.

### M00-T6 Release pipeline

- **Read:** [Local release preparation](#local-release-preparation)
- **Proves:** `scripts/release-check.sh`, `scripts/release-check-test.sh`, `tools/cmd/release-artifacts/main_test.go`

GoReleaser builds macOS and Linux binaries for arm64 and amd64. Each archive ships the third-party notices that `go-licenses` generates. The AUR package `ncly-bin` installs the binary and its completions. The Homebrew formula builds `ncly` from the release's source archive and generates its completions. `release.yml` starts only through `workflow_dispatch`, runs GoReleaser, and updates the formula. `just release-check` runs `govulncheck` and `goreleaser check` on top of `just check`. The AUR publication moved to M99.

## Local release preparation

The full local check runs on Linux with Go, Git, just, gitleaks, shellcheck, shfmt, Ruby, GNU `install`, `shasum`, and `tar`. The tools module requires Go 1.27.1, which Go downloads when automatic toolchain downloads are enabled. The application module still requires Go 1.27.0.

`just release-check vX.Y.Z` runs `just check`, the pinned `govulncheck` and `goreleaser check`, then prepares and checks a snapshot in its own temporary checkout. It accepts a release tag such as `v0.0.1`; an invalid tag fails before building. It uses no publishing credentials and invokes no publisher. A failed check returns a failure, never release readiness.

The snapshot contains four binary archives, for macOS and Linux on amd64 and arm64, plus a source archive and their SHA256 checksums. Each binary archive contains `ncly`, its MIT license, the runtime dependencies' licenses and adjacent NOTICE files, and completions for bash, zsh, and fish. Completions are generated once with a host binary. The license collector runs as a host program against each target's dependency graph.

The source archive contains the committed application source, `go.mod`, `go.sum`, and third-party notices. Building the extracted source must reproduce the requested version. Preparation uses an isolated committed checkout so that the source archive cannot silently omit the working changes being tested.

Stage additions and deletions before running the recipe. It captures current tracked files and rejects untracked files, missing tracked files, symbolic links, and source changes during capture. Each attempt keeps a fresh scratch directory, by default `/tmp/ncly-release-check.XXXXXXXX`, with artifacts under `tree/dist`. The recipe prints the retained path on success or failure after capture starts. Remove the scratch directory after inspecting its artifacts.

The generated `ncly-bin` PKGBUILD and `.SRCINFO` select the matching Linux archive, verify its checksum, and install its binary, notices, and completions. The Homebrew formula uses the source archive and its actual checksum, builds `ncly` with the release version, installs its notices, and generates its completions. Each local check has one proof owner.

| Owner | Local proof | Limit |
|---|---|---|
| Go [archive inspection](../../tools/cmd/release-artifacts/archive.go) and [source verification](../../tools/cmd/release-artifacts/verify.go) | Exactly four binary targets, archive contents, notices, completions, checksums, and byte-identical source rebuilds | AUR package proof belongs to Bash staging |
| Bash [AUR staging](../../scripts/verify-aur.sh) | Evaluated package metadata, local checksums, and staged installation contents and modes for both Linux architectures | AUR publication and installation moved to M99 |
| Formula syntax check in [release-check.sh](../../scripts/release-check.sh) | Ruby accepts the generated Homebrew formula | The Homebrew build and installation belong to the first release |
| Complete [release-check recipe](../../scripts/release-check.sh) | All local checks pass for the captured current tracked source, including final source stability and blocked-command checks | Publishing [prerequisites](#prerequisites) and the first release's manual check remain separate gates |

A zero exit from the full recipe proves the local rehearsal for that captured source. On success, review the retained artifacts. On failure, use any retained files to diagnose the failed check. Fix the cause, then rerun the full recipe. A component check on retained artifacts proves only that component's result. A source change requires a fresh full recipe run.

`release.yml` has only a manual `workflow_dispatch` trigger with a required tag input. It checks out and validates that exact tag before any publication. Preparing or pushing the workflow file does not run it. The first real dispatch, publishing authorization, macOS execution, and the installation check belong to the first release.

## Out of scope

- Every user command other than `--help`, `--version`, and `completion`.
- Extension discovery, beyond the empty extension registry of M00-T2: [M01](M01-extension-host.md).
- The form, spinner, progress bar, and step list of `internal/tui`. Each one lands with the first command that needs it.

## Contract coverage

The contract suite exercises behavior, not the spelling of source code. Each guarantee has one primary test owner. A later command adds coverage for its own effects and adapters instead of copying every shared test.

| Guarantee | M00 proof | First real proof |
|---|---|---|
| Parser failures | The real binary handles unknown commands, unknown flags, missing values, and invalid flag values in JSON mode, with `--json` before or after the bad argument and through `NCLY_JSON` | Each command adds its distinct validation cases |
| Streams and envelope | Parse the complete JSON answer, check `contract_version`, `ok`, exact exit code, and the correct stream. Test verbose diagnostics, warnings, and text-only help/version/completion exceptions | Doctor in M05 and `auth status` in M06 keep their reports on stdout when a check fails |
| Partial results and retry | Direct tests of the actual shared verdict and rendering components retain successful items and reject 75 after non-repeatable writes, paid dispatch, or unknown effects. Reverse error arrival order to prove the verdict stays the same | `agent run` in M10 and transcript in M13 inject failures around real operation boundaries and verify effects, not just fields |
| Compatibility | Independent consumer cases cover types, units, nullability, required fields, enum policy, defined array order, and ignored optional fields | Extensions in M01 and the transcript Python program in M13 test incompatible messages before any new effect |
| Dry run | Define the shared preparation/result contract and its conformance cases, including the bounded cache exception | Auth in M06 and transcript in M13 compare user files, config, keychain, run records, and external calls before and after dry run |
| Cancellation and resume | Define the signal verdicts and required partial-result representation | M01 checks process cleanup on an extension. M10, M13, and M15 check persisted evidence, lost stdout, and explicit resume |

No M00 key carries a unit, so the first key with one, such as `duration_ms`, adds its own consumer case. The scenario commands `snapshot` and `unchanged` are the dry-run conformance check: they compare every file under `$HOME` except Nicely's cache.

Direct component tests are justified where M00 exposes no real command that reaches the behavior. Test support stays in tests and calls production components. It does not implement an alternate CLI. Effects and resume are not marked verified until the first real command exercises them, from M10 to M15.

## CLI examples

```
# The version goes to stdout and nothing goes to stderr
exec ncly --version
stdout '^ncly v'
! stderr .

# An unknown command fails as one JSON line
exits 2 ncly nope --json
! stdout .
stderr '"ok":false'
stderr '"contract_version":1'
stderr '"code":"USAGE_INVALID"'

# Parser errors keep machine output even when --json follows the bad flag
exits 2 ncly --unknown-flag --json
! stdout .
stderr '"code":"USAGE_INVALID"'

# A missing flag value uses the same envelope
exits 2 ncly --json --lang
! stdout .
stderr '"code":"USAGE_INVALID"'

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

These snippets illustrate the scenarios. M00-T2 also decodes complete JSON answers so a matching substring cannot hide extra text, a wrong type, or an extra object. M00-T4 sharpens the pseudo-locale scenario so that it fails on any word outside the markers other than command names, flags, and examples.
