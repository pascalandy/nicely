# M0 Foundation

Status: planned
Release: v0.0.1

## Goal

An installable `ncly` that does almost nothing yet, built on finished plumbing. Every later command lands on the same contract, catalogs, config, interactive parts, tests, and release pipeline.

## Depends on

Nothing.

## Scope

**Repository.** `pascalandy/nicely` with the module path `github.com/pascalandy/nicely`, plus LICENSE, README.md, AGENTS.md, CONTRIBUTING.md, and `docs/`.

**Layout.**

```
cmd/ncly/           entry point, a few lines
internal/cli/       one package per domain or standalone command
internal/contract/  exit codes, error registry, JSON writers
internal/i18n/      catalogs in locales/, lookup, plural rules
internal/config/    TOML file, environment, flags, XDG paths
internal/platform/  OS differences: open, clipboard, trash, keychain
internal/tui/       the shared interactive parts
python/             Python components while they move in, empty in M0
testdata/script/    testscript scenarios
```

**Libraries.** Cobra and Fang for commands. Huh, Bubble Tea, Bubbles, Lip Gloss, and Glamour for the interface. go-i18n for catalogs, koanf for config, and testscript for scenarios.

**Agent contract.** Global flags, modes, output, exit codes, error registry, and configuration, as the M0 sections of [cli-spec.md](../north-star/cli-spec.md) specify.

**Translation.** An English catalog. Every entry has an ID and a description of where it appears. A test fails when another catalog misses a key from `en`. Help headings come from the catalog too.

**Interactive parts.** `internal/tui` holds the only components commands use:

- a form for missing values
- a spinner for work of unknown length
- a progress bar for work of known length
- a step list for pipelines
- a final summary with the `Next time:` line
- an error block that says what failed, why, and how to fix it

**Completion.** `ncly completion zsh|bash|fish`, with hooks ready for the values of M1.

**Local CI.**

- `justfile` recipes: `check`, `test`, `lint`, `signoff`, and `release-check`.
- Lefthook pre-commit: format, lint, and gitleaks with rules for home paths, hostnames, and local IPs.
- Lefthook pre-push: tests and testscript scenarios.
- The signoff status is a required check on `main`.

**Release.**

- GoReleaser builds macOS and Linux binaries for arm64 and amd64.
- It publishes to the Homebrew tap `pascalandy/homebrew-tap` and to the AUR as `ncly-bin`, both with completions.
- `go-licenses` generates the third-party notices that ship in each archive.
- The GitHub Actions workflow `release.yml` runs GoReleaser and starts only through `workflow_dispatch`.
- `just release-check` runs `govulncheck` and `goreleaser check` on top of `just check`.

**Testing helper.** A custom testscript command, `exits <code> <command>`, asserts an exact exit code, because the built-in `! exec` only asserts a failure.

## Out of scope

Every user command other than `--help`, `--version`, and `completion`.

## Acceptance

```
# The version goes to stdout and nothing goes to stderr
exec ncly --version
stdout '^ncly v'
! stderr .

# An unknown command fails as a JSON usage error
exits 2 ncly nope --json
! stdout .
stderr '"code":"USAGE_INVALID"'

# A language without a catalog falls back to English
exec ncly --lang xx --help
stdout 'Usage'

# NO_COLOR removes escape codes
env NO_COLOR=1
exec ncly --help
! stdout '\x1b\['
```

Manual checks on Pascal's machines:

1. On macOS, `brew install pascalandy/tap/ncly`, then `ncly <Tab>` completes in zsh.
2. On Omarchy, `yay -S ncly-bin`, then the same completion check.

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] Both manual checks pass
- [ ] v0.0.1 is tagged and released through `release.yml`

## Open questions

1. Can Fang take its help headings from the catalog? Test this first. If it cannot, keep Cobra's templates and style them with Lip Gloss.
2. Which Homebrew publisher does GoReleaser recommend for prebuilt binaries today? Use it, and confirm the installed binary runs without a Gatekeeper prompt on macOS.
