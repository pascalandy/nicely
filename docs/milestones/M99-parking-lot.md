# M99 Parking lot

Status: open

Ideas that are kept but belong to no delivery milestone. To start one, select a small useful slice, move it into a milestone, and remove only that slice here. These are options, not MVP acceptance criteria. [D061](../north-star/decisions/D061-smallest-useful-capability.md) records the scope audit of PR #156.

| Idea | What it would do | Revisit when |
|---|---|---|
| Routines | `ncly routine run <name>` runs a named sequence of commands from the config | A repeated sequence in an installed extension needs a named shortcut |
| Remote commands | `--host` runs a command on another machine over SSH, including `ncly doctor --host` | A remote task needs Nicely, or a remote command needs a key |
| Dotfiles for Linux machines | A dotfiles repository that carries the shared config to each Omarchy machine | Pascal adds a second Linux machine. This lives outside Nicely |
| Launcher | A bare `ncly` in interactive mode opens a fuzzy command palette, and prints help everywhere else | Several everyday commands ship and finding one becomes a recurring problem |
| Help by intent | The root help groups commands by purpose, such as tools, inspect and recover, and set up, and `ncly help --search <text>` finds a command by its words. Agents already search `ncly describe --json` | The root help lists more than ten commands, after a mockup that Pascal picks |
| `--explain` | Prints the underlying command and a link to the wrapped tool | The everyday domains wrap several tools |
| `ncly tools` | A showcase of every curated tool, why it was chosen, and its credits | The everyday domains ship |
| Aliases | Git-style aliases in the config, such as `md = "markdown view"` | A command feels too long in daily use |
| Completion after `--` | `ncly -- <Tab>` offers nothing, because ncly rejects a command name after `--`. Cobra offers the commands, so the fix answers Cobra's completion protocol in `Main` | A user hits it, or a command group starts taking arguments |
| `ncly ignite` | Bootstraps a project with Pascal's setup: Git, just, Lefthook, and signoff, through Copier templates | Pascal starts his next new project |
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |
| MCP skills server | Serves skills over the MCP Skills extension, SEP-2640, from the `skill` domain package rather than from CLI output. Per-file manifests join `view --json` as optional fields, never `list`, which agents read whole. Research in [#10](https://github.com/pascalandy/nicely/issues/10) | A host Pascal uses supports SEP-2640 and cannot read skills from disk. `ncly skill pack` answers the same need |
| Record check in core | Core reads the run record of an answer that exits 75 and turns it into a protocol failure when the record breaks [Retry safety](../north-star/contract/retry-safety.md), as [D045](../north-star/decisions/D045-step-effects.md) defers | A tap carries an extension that does not use the Go SDK |
| Answer limits | The answer check bounds the size of a held answer and of a held stderr line, its JSON depth, and rejects invalid UTF-8 and duplicate keys, then stops the extension and answers `RUNTIME`. An extension built on the SDK cannot produce these answers | A tap carries an extension that does not use the Go SDK |
| One test path for first-party extensions | First-party extensions run the static checks of the test kit, and `just install` writes their sidecars through `extension.Export`, as the kit does | A first-party extension breaks a rule that the kit checks, or its install differs from the kit's |
| Hooks | Lets an extension change how core or another extension behaves, which [D044](../north-star/decisions/D044-additive-extensions.md) excludes | A need appears that neither core, the SDK, nor the extension that owns the step can meet |
| Progress events | Long commands expose live JSON Lines events and a step list. Run records already keep effect boundaries and support inspection after interruption | Live progress adds value beyond the saved run record |
| Other extension languages | An extension written in a language other than Go, with an SDK and a test kit of its own. The wire contract already permits it, as [D040](../north-star/decisions/D040-wire-contract.md) decides | A user needs an extension that a Go handler and a worker cannot serve |
| Worker cache | Keeps an extracted worker between runs, instead of one extraction for each run | A measured extraction cost slows a command |
| Transcript in Go | Ports the Python program of the transcript extension to Go | Python blocks a release or a translation |
| Structured summaries | A bundled prompt writes, beside its Markdown, a versioned JSON file of decisions, actions, and open questions, each citing a segment of the transcript that a local check verifies | Pascal wants to search decisions or actions across summaries |
| Transcript on Omarchy | The paid end-to-end check of `ncly transcript` passes on Omarchy, with YouTube cookies from a Linux browser. [#23](https://github.com/pascalandy/nicely/issues/23) tracks it | Pascal needs transcripts on Omarchy |
| One duration parser | `ncly doctor` reads its `--timeout` through `sdk/command.Duration`, as `ncly agent run` does, instead of its own `parseTimeout` in `internal/cli/doctor.go`, so both commands share one rule | A change touches `ncly doctor --timeout`, or a third command takes `--timeout` |
| Nested deadlines and cleanup | The program runner passes `NCLY_DEADLINE`, an absolute UTC time, to every program, and each `ncly` ends its work before that time minus its 15 seconds of cleanup. A nested `ncly` also gets a shorter cleanup than its parent. It then times out, or writes its record after the signal, before the parent's SIGKILL | An approved summary slice actually calls a nested recorded command |
| More languages | Spanish, Japanese, and others, through the M22 workflow | A user asks, or `fr-CA` proves the workflow |
| Catalog folders | One folder per language, such as `locales/en/`, lets a catalog split under the size limit. [Text of an extension](../north-star/contract/text-of-an-extension.md) asks for one file per language, so `just limits` skips catalogs | A catalog grows too large to read whole. The change needs a contract decision |
| Terminal demos | VHS recordings of key commands for the README and the docs | The everyday domains ship |
| First-run experience | What a new visitor sees first, and the marketing around it | The everyday domains ship |
| Project config | A `.nicely.toml` in a repository, read between the environment and the local config | A project needs its own command settings or paths |
| `ncly config` | `ncly config get`, `set`, and `list` edit the config with comments kept | Agents edit the config by hand often enough to break it |
| Key checks for skills | `ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation` | A selected tap slice carries skills with prerequisite manifests |
| Man pages | Generated from Cobra and shipped in the packages | A user asks, or Nicely applies to homebrew-core |
| Build provenance | GitHub artifact attestations for the release archives | Strangers install from the release archives |
| AUR package | Publishes `ncly-bin` to the AUR, so that `yay -S ncly-bin` installs `ncly` and `ncly <Tab>` completes in bash on Omarchy. It needs an AUR account with the public key `nicely-release-aur`. GoReleaser builds the package that `just release-check` verifies, and `skip_upload: true` in the `aurs` section of `.goreleaser.yaml` keeps it off the AUR. To ship it, delete that line, and say in step 3 of the `release-nicely` skill that the workflow publishes the AUR package | AUR registration reopens |
| homebrew-core | A formula in Homebrew's main repository | Nicely meets Homebrew's notability rules |
| Windows | Support for Windows | Never planned. Revisit only on real demand |

## Parked capabilities

Each parked capability has its own file:

- [Agent extras](M99-parking-lot/agent-extras.md)
- [Taps](M99-parking-lot/taps.md)
- [Transcript summary](M99-parking-lot/transcript-summary.md)
- [Resume](M99-parking-lot/resume.md)
- [Transcript extras](M99-parking-lot/transcript-extras.md)
- [Everyday domains](M99-parking-lot/everyday-domains.md)
- [Docs hub](M99-parking-lot/docs-hub.md)
