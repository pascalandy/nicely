# Decisions

Each entry records what was decided, why, and what it replaced. Entries stay as written. To revise one, add a new entry that names the entry it supersedes.

## D001 Build from scratch in Go with Cobra and Charm

Decided 2026-10-05. Nicely is written in Go on Cobra, with Fang and the Charm libraries for the interface. It does not fork the Stripe CLI.

Most of the Stripe CLI is Stripe-specific, such as login, webhooks, API resources, and fixtures. Removing that code would cost more than building. Cobra already carries the practices of the Stripe CLI, `gh`, and `kubectl`.

## D002 Name the project Nicely and the command ncly

Decided 2026-10-05. The project is Nicely, the command is `ncly`, the repository is `pascalandy/nicely`, and the Go module is `github.com/pascalandy/nicely`.

`nicely` is taken on PyPI and npm, but neither package installs a `nicely` command, and Nicely ships only through Homebrew and the AUR, where the name is free. `ncly` is free on PyPI, npm, crates.io, and Homebrew, and it is short to type.

## D003 Put the domain first

Decided 2026-10-05. Commands read `ncly <domain> <action>`. Every script that joins Nicely is reshaped into domains.

An agent explores one domain at a time, so `ncly video --help` costs fewer tokens than a flat list of fifty verbs. Options stay specific to their domain. Names become predictable: after `ncly image convert`, an agent guesses `ncly video convert`. This replaced the verb-first form, such as `ncly convert`.

## D004 Split core and extensions

Decided 2026-10-05. Core holds what a stranger would use. Personal tools are extensions. Core says where a feature lives, and the milestones decide when it ships.

Nicely becomes the center of Pascal's tooling, but it must not carry his personal workflows into everyone's install.

## D005 Add extensions the Git way and carry them in taps

Decided 2026-10-05. An executable named `ncly-<domain>` becomes `ncly <domain>`. A tap is a Git repository that carries skills and extensions. Built-in commands always win over an extension with the same name.

Any script becomes a command with a name and `chmod +x`, in any language. One private tap moves Pascal's tools to every machine. This replaced a tap for skills only.

## D006 Serve agents through the CLI itself

Decided 2026-10-05. Agents operate `ncly` through the shell. `ncly` launches agents through their own non-interactive modes. Nicely uses no agent SDK.

A shell command is the one interface every harness shares, so Nicely stays independent of any single agent. Limiting Nicely to Claude would have been a design error.

## D007 Keep one stable agent contract

Decided 2026-10-05. Command names, flags, JSON keys, error codes, and exit codes stay in English and stable. Exit codes come from the transcript CLI, `0`, `1`, `2`, `75`, `130`, and `143`, plus `78` for setup a human must do. Errors are one JSON object with `code`, `message`, and `hint`.

The transcript CLI already proved these codes with agents. `78` lets an agent tell "a human must act" from "the call was wrong". Agents branch on `code`, so a translated `message` never breaks a parser. This replaced an earlier idea of a dedicated exit code 3 for a missing key.

## D008 Use one flag for overwrite and delete

Decided 2026-10-05. `--force` covers every overwrite and deletion. There is no `--yes`.

One concept gets one flag.

## D009 Ask in a terminal, fail elsewhere

Decided 2026-10-05. A missing value opens a form in interactive mode and exits 2 in non-interactive mode. After a form, `ncly` prints the equivalent command after `Next time:`.

Agents never meet a prompt. Beginners get a guided path, and the last line teaches them the flags.

## D010 Keep keys in the keychain or the environment

Decided 2026-10-05. Keys live in the OS keychain through go-keyring, or in environment variables. Agents relay the hint and never handle a key. There is no fallback file.

Pascal installs `ncly` locally on each machine and uses it in a desktop session, where the keychain answers. Remote commands that need keys are parked in M99.

## D011 Check locally by default, test live on request

Decided 2026-10-05. `ncly doctor` stays on the machine by default. `--live` calls free endpoints only. An install runs only after a yes in a terminal.

Doctor must be safe to run at any time. A live check proves a key works without a bill.

## D012 Make every word translatable from day one

Decided 2026-10-05. User-facing text lives in go-i18n catalogs with a description per entry. A test compares every catalog with `en`. English ships first, and Canadian French follows in M6.

Adding translation later means touching every string. AI makes catalogs cheap to keep complete, and a CLI in the user's language reaches people most CLIs ignore.

## D013 Write core in Go

Decided 2026-10-05. Core aims to be fully Go. A Python component may sit in core during its move, but it returns JSON only, and Go owns display, translation, and exit codes.

One binary and one display layer keep the experience identical across commands. This replaced the option of letting Python keep its own display.

## D014 Publish from the first commit

Decided 2026-10-05. The repository is public from day one, under the MIT license, with no telemetry. Issues and pull requests are welcome.

Public work keeps good habits from the start. Everything personal lives elsewhere: private taps, the config folder, and the docs hub.

## D015 Target macOS and Linux with the same paths

Decided 2026-10-05. Nicely runs on macOS and on Linux, with Omarchy as the reference. Both systems use XDG paths under `nicely`. Releases go to the Homebrew tap `pascalandy/homebrew-tap` and to the AUR as `ncly-bin`.

Pascal moves toward Linux machines over the years. The same paths everywhere keep his future dotfiles simple. The AUR is the native channel on Arch.

## D016 Run CI locally and release by hand

Decided 2026-10-05. Checks run through `just`, Lefthook, and `gh signoff`. GitHub Actions runs only when an agent starts the release workflow. External pull requests go through the maintainer's agent and `just signoff`.

Pascal often hits GitHub Actions limits, and local checks are faster.

## D017 Write acceptance criteria as testscript scenarios

Decided 2026-10-05. Each milestone states its acceptance criteria as testscript scenarios. End-to-end scenarios come before unit tests.

Pascal writes the specifications and agents write the Go. A scenario reads like a terminal session, so Pascal checks behavior without reading Go.

## D018 Name the agent command `agent` and keep one profile registry

Decided 2026-10-05. `ncly agent` runs a task through a harness. Profiles, each a harness, a model, and an effort, live in the Nicely config. Every tool reads them from there. A depth limit stops agents from launching agents without end. `ncly agent` absorbs the `headless` skill.

"Infer", "chat", and "prompt" mean little to most people or describe too little. "Harness" stays the internal term. Before this, each project declared its own inference settings.

## D019 Put transcript in core without showcasing it

Decided 2026-10-05. `transcript` is a core domain, and its command tree mirrors the existing transcript CLI: `ncly transcript run youtube` and `ncly transcript run zoom`.

It needs a Deepgram key, so a first-time visitor should start elsewhere. It handles Zoom audio as well as YouTube, so the earlier `ncly video transcribe` was the wrong domain.

## D020 Ship doctor, auth, transcript, and skill on day one

Decided 2026-10-05. M1 contains `doctor`, `auth`, `transcript`, and `skill`. `agent` waits for M2, so transcript keeps its own profiles until then.

## D021 Move skill requirements to a file

Decided 2026-10-05. Skills declare their requirements in frontmatter today. From M3, skills and extensions declare them in a `nicely.toml` file next to them.

One file format serves skills and extensions alike.

## D022 Keep docs private by default

Decided 2026-10-05. Each project in the docs hub is `public`, `private`, or `exclude`, and `private` is the default. The hub stays plain Markdown, so any engine can render it.

Personal details slip into docs without anyone noticing. Publishing a project must be a choice made for that project alone.

## D023 Configure with TOML

Decided 2026-10-05. The config file is `~/.config/nicely/config.toml`. Environment variables start with `NCLY_`. Precedence runs flags, environment, config file, defaults.

## D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions, and the packages install them. Completion suggests values such as skill names and profiles.

## D025 Organize planning around north-star and milestones

Decided 2026-10-05. `docs/north-star/` holds the guide, the CLI spec, and this log. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. A milestone adds its commands to the CLI spec before code.

One spec avoids drift between milestones. Agents read the guide every session and the rest only when they need it.

## D026 Write the repository in English

Decided 2026-10-05. Code, comments, docs, and commit messages are in English. User-facing text goes through the catalogs.

The repository is public, and English reaches the most contributors.
