# Nicely glossary

Use one word per concept in code, docs, help, and commit messages. Add a term here before you use a new one.

- **Agent**: an AI program that works through a shell. The `agent` extension launches one.
- **Agent contract**: everything a script or an agent parses, defined in [contract.md](contract.md) and the extension specs.
- **Artifact**: a result produced by an operation, referenced by its path and the evidence needed to verify it
- **Bundled extension**: an extension compiled into `ncly`, with the same manifest and contract as an external one. Only `skill` is bundled.
- **Card**: one pull request of a milestone, with its owner, its dependencies, what to read, and what proves it.
- **Core**: the host that every install gets: dispatch, help, completion, `describe`, `doctor`, `auth`, `run`, and `tap`.
- **Domain**: a top-level noun that groups actions, such as `video`.
- **Extension**: a domain added to core through a manifest, bundled or external.
- **External extension**: an executable named `ncly-<domain>` with a compatible manifest beside it.
- **First-party extension**: an extension in `extensions/` of this repository, such as `agent` or `transcript`.
- **Grant**: the permission, given by a human in a terminal, for one extension to receive one key.
- **Harness**: the program an agent runs in: Claude Code, Codex, Pi, OpenCode, or Grok.
- **Interactive mode**: the mode in which `ncly` may ask questions. [contract.md](contract.md#modes) defines when it applies. Every other case is non-interactive mode.
- **Local config**: `config.local.toml`, which holds what belongs to one machine and overrides the shared config.
- **Manifest**: the file that declares an extension's commands, effects, modes, and requirements, read without starting the extension.
- **Milestone**: a numbered stage of work in `docs/milestones/`, one capability that a sentence can demonstrate. Its status is `planned`, `ready`, `active`, or `done`. M99 is the parking lot, and its status stays `open`.
- **Operation**: the work prepared and performed by a command. The domain owns its steps
- **Profile**: a named combination of harness, provider, model, and effort, stored in the config.
- **Pseudo-locale**: a generated test language that marks every catalog string, so text outside the catalog stands out.
- **Python program**: Python code that an extension embeds while it moves to Go, such as transcript's.
- **Report command**: a command whose result is a report, such as `ncly doctor`. Its report goes to stdout even when a check fails.
- **Resume**: an explicit request to continue a recorded run after validating its inputs, artifacts, and remaining steps
- **Run record**: the durable evidence of one operation, identified by `run_id`, including its step outcomes and artifact references
- **SDK**: the public Go packages in `sdk/` that implement the contract and the shared services, for core and for every Go extension.
- **Shared config**: `config.toml`, the setup the user wants, which can travel between machines.
- **Skill**: a folder with a `SKILL.md` that teaches an agent a task.
- **Tap**: a Git repository that holds skills and extensions, like a Pi package. The Homebrew tap that ships `ncly` is a different thing, always called the Homebrew tap.
