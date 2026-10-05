# Decisions

Each entry records what was decided, why, and what was rejected. The rules themselves live in the [guide](guide.md) and in [cli-spec.md](cli-spec.md), and an entry links to its rule instead of repeating it.

Until v0.0.1 ships, fix an entry in place, because git keeps the old text. From v0.0.1, add a new entry that names the entry it replaces, and mark the old entry `Replaced by Dxxx`.

## D001 Build from scratch in Go with Cobra and Charm

Decided 2026-10-05. Nicely is written in Go on Cobra, with Lip Gloss, Huh, Bubble Tea, Bubbles, and Glamour for the interface.

**Why.** Cobra already carries the practices of `gh`, `kubectl`, and the Stripe CLI. The Charm libraries cover forms, spinners, styles, and Markdown in one family.

**Rejected.** Forking the Stripe CLI, because most of its code is Stripe-specific, such as login, webhooks, API resources, and fixtures. Fang, because its source hard-codes the help headings `usage`, `flags`, `examples`, and `ERROR` in English and takes `-v` for `--version`, which Nicely uses for `--verbose`. Checked on 2026-10-05.

## D002 Name the project Nicely and the command ncly

Decided 2026-10-05. The project is Nicely, the command is `ncly`, the repository is `pascalandy/nicely`, and the Go module is `github.com/pascalandy/nicely`.

**Why.** `nicely` is taken on PyPI and npm, but neither package installs a `nicely` command, and Nicely ships only through Homebrew and the AUR, where the name is free. `ncly` is free on PyPI, npm, crates.io, and Homebrew, and it is short to type.

## D003 Put the domain first and keep one verb per action

Decided 2026-10-05. Commands follow the grammar in [cli-spec.md](cli-spec.md#usage-m0): `ncly <domain> [<resource>] <verb>`, with singular nouns and one verb per action across domains.

**Why.** An agent explores one domain at a time, so `ncly video --help` costs fewer tokens than a flat list of fifty verbs. Options stay specific to their domain. Names become predictable: after `ncly image convert`, an agent guesses `ncly video convert`. The verbs are those of `gh`, which agents already know.

**Rejected.** The verb-first form, such as `ncly convert`. Keeping the interface of a script that joins Nicely, such as `transcript list prompts`, because one mixed shape breaks the guess.

## D004 Split core and extensions

Decided 2026-10-05. Core holds what a stranger would use. Personal tools are extensions. Core says where a feature lives, and the milestones decide when it ships.

**Why.** Nicely becomes the center of Pascal's tooling, but it must not carry his personal workflows into everyone's install.

## D005 Add extensions the Git way and carry them in taps

Decided 2026-10-05. An executable named `ncly-<domain>` becomes `ncly <domain>`. A tap is a Git repository that carries skills and extensions. Built-in commands always win over an extension with the same name. An extension receives the global flags as the variables in [cli-spec.md](cli-spec.md#global-flags-m0), and it prefixes its error codes with its domain.

**Why.** Any script becomes a command with a name and `chmod +x`, in any language. One private tap moves Pascal's tools to every machine. Git, `kubectl`, and cargo find their plugins the same way.

**Rejected.** A tap for skills only. Translating extension text from core, because core cannot know an extension's strings.

## D006 Serve agents through the CLI itself

Decided 2026-10-05. Agents operate `ncly` through the shell. `ncly` launches agents through their own non-interactive modes. Nicely uses no agent SDK.

**Why.** A shell command is the one interface every harness shares, so Nicely stays independent of any single agent.

**Rejected.** Limiting Nicely to Claude.

## D007 Keep one stable agent contract

Decided 2026-10-05. Command names, flags, JSON keys, error codes, and exit codes stay in English and stable. The exit codes and the error registry live in [cli-spec.md](cli-spec.md#exit-codes-m0). Each core error code maps to one exit code. Exit `78` means that a human must act, and exit `75` means that the same command is safe to rerun.

**Why.** The transcript CLI already proved `0`, `1`, `2`, `75`, `130`, and `143` with agents. `78` is `EX_CONFIG` in `sysexits.h`, and it lets an agent tell "a human must act" from "the call was wrong". Agents branch on `code`, so a translated `message` never breaks a parser.

**Rejected.** A dedicated exit code 3 for a missing key. Exit 2 for a step that only a human can do, such as `ncly auth login` without a terminal or a broken config file, because the agent cannot fix those by changing the call.

## D008 Use one flag for overwrite and delete

Decided 2026-10-05. `--force` covers every overwrite and deletion. There is no `--yes`.

**Why.** clig.dev asks for `-f` or `--force` when a confirmation cannot be asked. Every confirmation in Nicely guards an overwrite or a deletion, so a second flag would name the same step twice.

**Rejected.** `--yes` beside `--force`, as in Pascal's script conventions, which serve scripts that have both a prompt and a separate safety check.

## D009 Ask in a terminal, fail elsewhere

Decided 2026-10-05. A missing value opens a form in interactive mode and exits 2 in non-interactive mode. After a form, `ncly` prints the equivalent command after `Next time:`. `CI` and `NCLY_NO_INPUT` also select non-interactive mode.

**Why.** Agents never meet a prompt. Beginners get a guided path, and the last line teaches them the flags. Some agent tools run commands in a pseudo-terminal, so a terminal alone does not prove that a human is there.

## D010 Keep keys in the keychain or the environment

Decided 2026-10-05. Keys live in environment variables or in the OS keychain through go-keyring, under the service `nicely` and an account named after the service. The environment wins. Any service name works, and its variable name derives from it, as [cli-spec.md](cli-spec.md#keys-m1) says. Agents relay the hint and never handle a key.

**Why.** Each program owns its own keychain entries, as `gh` does with `gh:github.com`, so `logout` touches only Nicely's keys. The environment lets any other secret store feed `ncly`. Open service names let an extension declare a key that core does not know.

**Rejected.** A fallback file. Sharing entries written by another tool, such as chezmoi's `service=deepgram, user=api_key`, because two programs would then own one entry. Remote commands that need keys, parked in M99.

## D011 Check locally by default, test live on request

Decided 2026-10-05. `ncly doctor` stays on the machine by default. `--live` calls free endpoints only. An install runs only after a yes in a terminal.

**Why.** Doctor must be safe to run at any time. A live check proves a key works without a bill.

## D012 Make every word of core translatable from day one

Decided 2026-10-05. Core text lives in go-i18n catalogs with a description per entry. A test compares every catalog with `en`, and a pseudo-locale scenario exposes text outside the catalog from M0. Plurals, numbers, sizes, and dates go through `internal/i18n`. English ships first, and Canadian French follows in M6.

**Why.** Adding translation later means touching every string and every number format. AI makes catalogs cheap to keep complete, and a CLI in the user's language reaches people most CLIs ignore. The pseudo-locale catches a missing string in M0 instead of M6.

**Rejected.** A static check for string literals, because the pseudo-locale also catches truncated layouts.

## D013 Write core in Go

Decided 2026-10-05. Core aims to be fully Go. A Python program may sit in core during its move, but it returns JSON only, and Go owns display, translation, and exit codes.

**Why.** One binary and one display layer keep the experience identical across commands.

**Rejected.** Letting Python keep its own display.

## D014 Publish from the first commit

Decided 2026-10-05. The repository is public from day one, under the MIT license, with no telemetry. Issues and pull requests are welcome.

**Why.** Public work keeps good habits from the start. Everything personal lives elsewhere: private taps, the config folder, and the docs hub.

## D015 Target macOS and Linux with the same paths

Decided 2026-10-05. Nicely runs on macOS and on Linux, with Omarchy as the reference. Both systems use XDG paths under `nicely`. On macOS, a formula in the Homebrew tap `pascalandy/homebrew-tap` builds `ncly` from source. On Arch, the AUR package `ncly-bin` installs the prebuilt binary.

**Why.** Pascal moves toward Linux machines over the years, and the same paths everywhere keep his future dotfiles simple. A binary built on the user's machine never meets Gatekeeper, so Nicely needs no Apple Developer account, and Homebrew generates the completions itself. Homebrew builds `gh` from source the same way. The AUR is the native channel on Arch.

**Rejected.** Signing and notarizing, which costs US$99 per year. A cask that strips the quarantine flag with `xattr`, which GoReleaser discourages. GoReleaser's `brews` section, deprecated in v2.10.

## D016 Run CI locally and release by hand

Decided 2026-10-05. Checks run through `just`, Lefthook, and `gh signoff`. GitHub Actions runs only when an agent starts the release workflow. External pull requests go through the maintainer's agent and `just signoff`.

**Why.** Pascal often hits GitHub Actions limits, and local checks are faster.

## D017 Write acceptance criteria as testscript scenarios

Decided 2026-10-05. Each milestone states its acceptance criteria as testscript scenarios. End-to-end scenarios come before unit tests.

**Why.** Pascal writes the specifications and agents write the Go. A scenario reads like a terminal session, so Pascal checks behavior without reading Go.

## D018 Name the agent command `agent` and keep one profile registry

Decided 2026-10-05. `ncly agent` runs a task through a harness from M2. Profiles live in the config from M1, as [cli-spec.md](cli-spec.md#agent-profiles-m1) defines, and every command that runs an agent reads them there. A depth limit stops agents from launching agents without end. `ncly agent` absorbs the `headless` skill.

**Why.** Model IDs change every few months, so they belong in the config, not in code. With one registry from M1, `--profile` keeps its meaning when `ncly agent` arrives. "Harness" stays the internal term.

**Rejected.** "Infer", "chat", and "prompt", which mean little to most people or describe too little. Each project declaring its own inference settings. Transcript keeping its own profiles until M2, which would change `--profile` in v0.2.0.

## D019 Put transcript in core without showcasing it

Decided 2026-10-05. `transcript` is a core domain: `ncly transcript run youtube` and `ncly transcript run zoom`.

**Why.** It needs a Deepgram key, so a first-time visitor should start elsewhere. It handles Zoom audio as well as YouTube.

**Rejected.** `ncly video transcribe`, the wrong domain for Zoom audio.

## D020 Ship doctor, auth, skill, and transcript on day one

Decided 2026-10-05. M1 contains `doctor`, `auth`, `skill`, and `transcript`, plus the agent profiles in the config. `ncly agent` waits for M2.

**Why.** Day one covers the setup of a new machine, skill discovery, and the first real command. Pascal has no need for an agent command on day one.

## D021 Move skill requirements to a file

Decided 2026-10-05. Skills declare their requirements in frontmatter today. From M3, skills and extensions declare them in a `nicely.toml` file next to them. Until then, `ncly doctor skill` checks only the structure of a skill.

**Why.** One file format serves skills and extensions alike. A frontmatter reader built in M1 would be thrown away in M3.

## D022 Keep docs private by default

Decided 2026-10-05. Each project in the docs hub is `public`, `private`, or `exclude`, and `private` is the default. The hub stays plain Markdown, so any engine can render it.

**Why.** Personal details slip into docs without anyone noticing. Publishing a project must be a choice made for that project alone.

## D023 Configure with a shared TOML file and a local one

Decided 2026-10-05. The shared config describes the setup the user wants, and the local config overrides it on one machine, as [cli-spec.md](cli-spec.md#configuration-m0) defines. `ncly` edits a config file only through a command whose job is to change the setup, and it keeps comments and formatting. An unknown key is a warning. Environment variables start with `NCLY_`.

**Why.** TOML is easy to read and edit, and Pascal's tools already use it. One shared file lets a new machine reach the same setup, for example with `ncly tap sync`. The local file keeps machine paths out of the shared one. A warning on unknown keys keeps older versions of `ncly` working.

**Rejected.** YAML. A config that only humans write, with Nicely's own records kept per machine, because each new machine would then need every tap added again.

## D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions, and the packages install them. Completion suggests values such as skill names and profiles.

**Why.** Completion is how a terminal user discovers flags and values. Cobra builds it from the command tree, so it never drifts from the commands.

**Rejected.** Hand-written completion scripts.

## D025 Keep rules in north-star and work in milestones

Decided 2026-10-05. `docs/north-star/` holds the guide, the CLI spec, and this log. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. When a milestone starts, it moves its draft spec into cli-spec.md and links to it.

**Why.** Each rule has one source, so milestones cannot drift from the spec. Agents read the guide every session and the rest only when they need it.

**Rejected.** A spec inside each milestone file.

## D026 Write the repository in English

Decided 2026-10-05. Code, comments, docs, and commit messages are in English. User-facing text goes through the catalogs.

**Why.** The repository is public, and English reaches the most contributors.

## D027 Answer in one JSON line with `ok` and `errors`

Decided 2026-10-05. With `--json`, every answer is one JSON object on one line with `ok`. A failure lists `errors`, each with `code`, `message`, and `hint`, and the first error decides the exit code. A report command keeps its report on stdout. JSON keys are only added. [cli-spec.md](cli-spec.md#output-m0) holds the details.

**Why.** Stripe, npm, JSON:API, and GraphQL all answer with error objects that carry a stable code. The key `errors` matches Pascal's script-output convention, so his Python scripts can become extensions by turning their error strings into objects. `ok` stays readable when an agent merges stdout and stderr. ESLint, ShellCheck, `terraform validate -json`, and `npm audit --json` keep their reports on stdout when a check fails. One line keeps `tail -n1 | jq` working.

**Rejected.** A flat error object without `ok`. A single `error` object. JSON by default without `--json`, because humans use the same commands. A report on stderr when a check fails.

## D028 Keep outside text away from agent tools

Decided 2026-10-05. Text from outside Nicely, such as a transcript or a web page, reaches an agent only with the agent's tools turned off.

**Why.** A video can carry instructions. The transcript CLI already summarizes with a tool-free `claude --print` for that reason.

**Rejected.** Relying on `--read-only`, which detects a change only after it happened.

## D029 Decide early what is costly to change

Decided 2026-10-05. A rule that a later milestone would otherwise rewrite goes into the north-star files before M0 code is written. Its code waits for the milestone that uses it.

**Why.** Extensions, Canadian French, the shared config, and agent profiles all put constraints on the contract, the config, and the catalogs. A rule written now costs a paragraph. The same rule found in M3 or M6 costs a rewrite of M0 code.

**Rejected.** Deciding each rule only when its milestone starts.
