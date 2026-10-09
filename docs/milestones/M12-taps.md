# M12 Taps

Status: planned
Version: v0.12.0

## Demo

`ncly tap add` brings a Git repository's skills and extensions to this machine, and installs `agent` from the official tap.

## Scope

A tap is a Git repository that carries skills and extensions, like a Pi package. The nicely repository is the official tap of the first-party extensions.

```
ncly tap add <owner/repo | git-url> [--dry-run]
ncly tap list [--json]
ncly tap sync [<tap>] [--dry-run]
ncly tap remove <tap> [--dry-run] [--force]
ncly skill link [<name>] [--harness <name>] [--dry-run]
```

- `owner/repo` means a GitHub repository. Any Git URL works
- A tap has a canonical source identity that distinguishes hosts and repositories. Record its resolved revision and each extension's origin. Two executables with the same domain do not share an identity merely because their names match
- The shared config lists the taps. `add` writes the tap into it and clones the tap below `~/.local/share/nicely/taps/`, keyed by its canonical source identity, with the user's own Git credentials. Nicely handles no Git authentication
- `add` lists the keys that the tap's extensions declare and asks the human to grant each one. Without a terminal, `add` grants nothing, and `--force` never grants a key
- `sync` stages updates for the configured taps, validates their manifests, and publishes each managed destination under a lock, as [D035](../north-star/decision-records.md#d035-track-ownership-when-synchronizing-files) decides. It records the resolved revisions. Each destination has its own atomic publication boundary; unrelated taps are not one transaction. Local edits are preserved and reported as conflicts. When an update declares a new key, the extension does not start until a human grants it
- `remove` takes the tap out of the shared config and moves its clone to the trash
- A tap holds `skills/<name>/SKILL.md` and `extensions/ncly-<domain>` with `extensions/ncly-<domain>.toml`. A skill keeps its requirements in `nicely.toml` beside its `SKILL.md`, in the manifest format of M01
- Explicit `[skill] paths` win in their configured order, followed by taps in shared configuration order after local overrides. `ncly skill view <tap>/<name>` selects a tap skill. Discovery reports its origin, revision, and any shadowing
- `skill link` symlinks skills into the skill folder of each harness, one line per harness in `[skill.link]`, such as `claude = "~/.claude/skills"`. Linking twice leaves the same managed links. A source switch updates only a link that Nicely owns and that the user has not changed. An existing user file or unmanaged link is a conflict. Tap removal checks its links first: a user-modified link blocks removal, and unchanged managed links go with the tap. Dry run lists these effects before any config edit or trash operation

`link` joins the verb table of [Usage](../north-star/cli-spec.md#usage). A public registry of extensions and signed extensions stay out of scope.

Scenarios build a local Git repository as a fixture tap and add it with a file path URL. A counted extension stub and filesystem checks prove dry run and retry behavior. Metadata alone does not prove conformance.

## Open questions

Settle each one in a new `ncly tap` section of [cli-spec.md](../north-star/cli-spec.md), then delete this section.

- The canonical tap ID for arbitrary Git URLs, avoiding collisions between hosts, and its mapping to local storage
- The editor for config files. `tomledit` keeps comments and formatting
- Recovery of an interrupted tap, config, or link update, with ownership evidence and retry rules for the whole invocation
- How a tap ships a Go extension: a binary published for each system, or a build on the user's machine
- How `ncly` finds the first-party extensions in the official tap, and how `just install` gives way to it
- How Pascal's Python scripts move from the `{"ok":false,"errors":["…"]}` strings of his script-output convention to the error objects of D027

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M12-T1 | Tap add and list | agent | — | todo |
| M12-T2 | Tap sync and remove | agent | M12-T1 | todo |
| M12-T3 | Extensions and skills from taps | agent | M12-T1 | todo |
| M12-T4 | Skill link | agent | M12-T3 | todo |
| M12-T5 | The official tap | agent | M12-T3 | todo |

### M12-T1 Tap add and list

- **Read:** [Configuration](../north-star/cli-spec.md#configuration), [Grants](../north-star/cli-spec.md#grants)
- **Proves:** `testdata/script/tap_add.txtar`

`ncly tap add` clones a fixture tap, writes it into the shared config, and asks for each declared key in a terminal. `ncly tap list` shows each tap's identity and revision.

### M12-T2 Tap sync and remove

- **Read:** [D035](../north-star/decision-records.md#d035-track-ownership-when-synchronizing-files)
- **Proves:** `testdata/script/tap_sync.txtar`, `testdata/script/tap_remove.txtar`

`sync` publishes each destination under its lock, preserves local edits as conflicts, and survives a repeated or interrupted sync. `remove` moves the clone to the trash.

### M12-T3 Extensions and skills from taps

- **Read:** [Extensions](../north-star/cli-spec.md#extensions), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `testdata/script/tap_extensions.txtar`

A tap's extensions become domains and its skills join `ncly skill list`, with deterministic precedence. Two taps that publish the same domain are reported, and an update that declares a new key gets nothing until a human grants it.

### M12-T4 Skill link

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `testdata/script/skill_link.txtar`

`ncly skill link` with managed links, conflicts, and the checks that tap removal makes first.

### M12-T5 The official tap

- **Read:** [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/tap_official.txtar`

A user without the repository installs `agent` from the official tap, and `just install` gives way to it, as the `ncly tap` section settles. Every later first-party extension, such as `transcript`, ships the same way.

## After this milestone

These steps live outside this repository and block no card. Pascal decides when.

- An agent turns Pascal's `sync-fleet` script into `ncly-fleet` in his private tap, his first personal extension, and Pascal runs `ncly fleet sync` on macOS and on Omarchy
- `ncly skill link` replaces `just install-skills` on one machine, and the `sync`, `install-skills`, and `sync-fleet` recipes of `pascalandy/skills` retire
