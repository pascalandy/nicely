# M99 Parking lot

Status: open

Ideas that are kept but belong to no milestone. Add one row per idea. To start one, move it into a milestone file and delete its row here.

| Idea | What it would do | Revisit when |
|---|---|---|
| Routines | `ncly routine run <name>` runs a named sequence of commands from the config | The fleet extension of M12 works and a repeated sequence shows up |
| Remote commands | `--host` runs a command on another machine over SSH, including `ncly doctor --host` | A remote task needs Nicely, or a remote command needs a key |
| Dotfiles for Linux machines | A dotfiles repository that carries the shared config to each Omarchy machine | Pascal adds a second Linux machine. This lives outside Nicely |
| Launcher | A bare `ncly` in interactive mode opens a fuzzy command palette, and prints help everywhere else | The everyday domains of M18 to M20 exist |
| `--explain` | Prints the underlying command and a link to the wrapped tool | The everyday domains wrap several tools |
| `ncly tools` | A showcase of every curated tool, why it was chosen, and its credits | The everyday domains ship |
| Aliases | Git-style aliases in the config, such as `md = "markdown view"` | A command feels too long in daily use |
| Completion after `--` | `ncly -- <Tab>` offers nothing, because ncly rejects a command name after `--`. Cobra offers the commands, so the fix answers Cobra's completion protocol in `Main` | A user hits it, or a command group starts taking arguments |
| `ncly ignite` | Bootstraps a project with Pascal's setup: Git, just, Lefthook, and signoff, through Copier templates | Pascal starts his next new project |
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |
| MCP skills server | Serves skills over the MCP Skills extension, SEP-2640, from the `skill` domain package rather than from CLI output. Per-file manifests join `view --json` as optional fields, never `list`, which agents read whole. Research in [#10](https://github.com/pascalandy/nicely/issues/10) | A host Pascal uses supports SEP-2640 and cannot read skills from disk. `ncly skill pack` answers the same need |
| Agent code review | Exposes the built-in review mode of each harness through the `agent` extension | Pascal reviews code through `ncly agent` weekly |
| Public docs build | `ncly docs build --public` builds and hosts only the `public` projects | A project needs public docs |
| Record check in core | Core reads the run record of an answer that exits 75 and turns it into a protocol failure when a step that has left `pending` declares `non_repeatable` or `paid`, as [D045](../north-star/decisions/D045-step-effects.md) defers | A tap carries an extension that does not use the Go SDK |
| Hooks | Lets an extension change how core or another extension behaves, which [D044](../north-star/decisions/D044-additive-extensions.md) excludes | A need appears that neither core, the SDK, nor the extension that owns the step can meet |
| Progress events | Long commands expose live JSON Lines events and a step list. Run records already keep effect boundaries and support inspection after interruption | Live progress adds value beyond the saved run record |
| Transcript in Go | Ports the Python program of the transcript extension to Go | Python blocks a release or a translation |
| Transcript on Omarchy | The paid end-to-end check of `ncly transcript` passes on Omarchy, with YouTube cookies from a Linux browser. [#23](https://github.com/pascalandy/nicely/issues/23) tracks it | Pascal needs transcripts on Omarchy |
| Nested deadlines and cleanup | The program runner passes `NCLY_DEADLINE`, an absolute UTC time, to every program, and each `ncly` ends its work before that time minus its 15 seconds of cleanup. A nested `ncly` also gets a shorter cleanup than its parent. It then times out, or records `interrupted`, before the parent's SIGKILL | M14 makes transcript call `ncly agent run` |
| More languages | Spanish, Japanese, and others, through the M22 workflow | A user asks, or `fr-CA` proves the workflow |
| Terminal demos | VHS recordings of key commands for the README and the docs | The everyday domains ship |
| First-run experience | What a new visitor sees first, and the marketing around it | The everyday domains ship |
| Project config | A `.nicely.toml` in a repository, read between the environment and the local config | A project needs its own profile or paths |
| `ncly config` | `ncly config get`, `set`, and `list` edit the config with comments kept | Agents edit the config by hand often enough to break it |
| Key checks for skills | `ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation` | Taps carry skills with manifests, in M12 |
| Man pages | Generated from Cobra and shipped in the packages | A user asks, or Nicely applies to homebrew-core |
| Build provenance | GitHub artifact attestations for the release archives | Strangers install from the release archives |
| AUR package | Publishes `ncly-bin` to the AUR, so that `yay -S ncly-bin` installs `ncly` and `ncly <Tab>` completes in bash on Omarchy. It needs an AUR account with the public key `nicely-release-aur`. Before the first release, set `skip_upload: true` in the `aurs` section of `.goreleaser.yaml`, and update step 3 of Release in AGENTS.md. GoReleaser then still builds the package that `just release-check` verifies, and pushes nothing to the AUR. Without `skip_upload: true`, GoReleaser stops after the GitHub release and before the Homebrew tap update | AUR registration reopens |
| homebrew-core | A formula in Homebrew's main repository | Nicely meets Homebrew's notability rules |
| Windows | Support for Windows | Never planned. Revisit only on real demand |
