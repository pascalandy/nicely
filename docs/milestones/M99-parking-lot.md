# M99 Parking lot

Status: open

Ideas that are kept but belong to no milestone. Add one row per idea. To start one, move it into a milestone file and delete its row here.

| Idea | What it would do | Revisit when |
|---|---|---|
| Routines | `ncly routine run <name>` runs a named sequence of commands from the config | The fleet extension works and a repeated sequence shows up |
| Remote commands | `--host` runs a command on another machine over SSH, including `ncly doctor --host` | A remote task needs Nicely, or a remote command needs a key |
| Dotfiles for Linux machines | A dotfiles repository that carries the shared config to each Omarchy machine | Pascal adds a second Linux machine. This lives outside Nicely |
| Launcher | A bare `ncly` in interactive mode opens a fuzzy command palette, and prints help everywhere else | M4 domains exist |
| `--explain` | Prints the underlying command and a link to the wrapped tool | M4 commands wrap several tools |
| `ncly tools` | A showcase of every curated tool, why it was chosen, and its credits | M4 ships |
| Aliases | Git-style aliases in the config, such as `md = "markdown view"` | A command feels too long in daily use |
| `ncly ignite` | Bootstraps a project with Pascal's setup: Git, just, Lefthook, and signoff, through Copier templates | Pascal starts his next new project |
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |
| Agent code review | Exposes the built-in review mode of each harness through `ncly agent` | Pascal reviews code through `ncly agent` weekly |
| Public docs build | `ncly docs build --public` builds and hosts only the `public` projects | A project needs public docs |
| Progress events | Long commands expose live JSON Lines events and a step list. M1 already records effect boundaries internally and supports inspection after interruption | Live progress adds value beyond the saved run record |
| Transcript in Go | Ports the Python transcript program to Go | Python in core blocks a release or a translation |
| More languages | Spanish, Japanese, and others, through the M6 workflow | A user asks, or `fr-CA` proves the workflow |
| Terminal demos | VHS recordings of key commands for the README and the docs | M4 ships |
| First-run experience | What a new visitor sees first, and the marketing around it | M4 ships |
| Project config | A `.nicely.toml` in a repository, read between the environment and the local config | A project needs its own profile or paths |
| `ncly config` | `ncly config get`, `set`, and `list` edit the config with comments kept | Agents edit the config by hand often enough to break it |
| Key checks for skills | `ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation` | M3 manifests exist |
| Man pages | Generated from Cobra and shipped in the packages | A user asks, or Nicely applies to homebrew-core |
| Build provenance | GitHub artifact attestations for the release archives | Strangers install from the release archives |
| homebrew-core | A formula in Homebrew's main repository | Nicely meets Homebrew's notability rules |
| Windows | Support for Windows | Never planned. Revisit only on real demand |
