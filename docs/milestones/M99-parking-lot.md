# M99 Parking lot

Status: open

Ideas that are kept but belong to no milestone. Add one row per idea. To start one, move it into a milestone file and delete its row here.

| Idea | What it would do | Revisit when |
|---|---|---|
| Routines | `ncly run <name>` runs a named sequence of commands from the config | The fleet extension works and a repeated sequence shows up |
| Remote commands | `--host` runs a command on another machine over SSH, including `ncly doctor --host` | A remote task needs Nicely, or a remote command needs a key |
| Dotfiles for Linux machines | A dotfiles repository that carries `~/.config/nicely` to each Omarchy machine | Pascal adds a second Linux machine. This lives outside Nicely |
| Launcher | A bare `ncly` in a terminal opens a fuzzy command palette | M4 domains exist |
| `--explain` | Prints the underlying command and a link to the wrapped tool | M4 commands wrap several tools |
| `ncly tools` | A showcase of every curated tool, why it was chosen, and its credits | M4 ships |
| Aliases | Git-style aliases in the config, such as `md = "markdown view"` | A command feels too long in daily use |
| `ncly ignite` | Bootstraps a project with Pascal's setup: Git, just, Lefthook, and signoff, through Copier templates | Pascal starts his next new project |
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |
| Agent code review | Exposes the built-in review mode of each harness through `ncly agent` | Pascal reviews code through `ncly agent` weekly |
| Public docs build | `ncly docs build --public` builds and hosts only the `public` projects | A project needs public docs |
| Progress events | Python components stream step events that Go shows as a step list | A transcript run feels long without feedback |
| Transcript in Go | Ports the Python transcript component to Go | Python in core blocks a release or a translation |
| More languages | Spanish, Japanese, and others, through the M6 workflow | A user asks, or `fr-CA` proves the workflow |
| Terminal demos | VHS recordings of key commands for the README and the docs | M4 ships |
| First-run experience | What a new visitor sees first, and the marketing around it | M4 ships |
| Windows | Support for Windows | Never planned. Revisit only on real demand |
