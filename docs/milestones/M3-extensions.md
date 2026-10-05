# M3 Extensions

Status: planned
Release: v0.3.0

## Goal

Any executable becomes an `ncly` domain, and taps carry skills and extensions to every machine. Fleet sync becomes Pascal's first personal extension.

## Depends on

M1, which brings `internal/run`, the agent depth, and the profiles that extensions inherit.

## Scope

**Spec first.** Move the sections for extensions, `tap`, `skill link`, and `nicely.toml` into [cli-spec.md](../north-star/cli-spec.md) before writing Go. `link` joins the verb table.

**Discovery.**

- An executable named `ncly-<domain>` becomes `ncly <domain>`. `ncly` looks in `~/.local/share/nicely/extensions/` first, then in `PATH` order, the same as Git, `kubectl`, and cargo. An extension found in `PATH` needs no opt-in.
- Built-in commands and reserved names always win. A new `extensions` component of `ncly doctor` lists every extension it finds, where it came from, any extension that a built-in, a reserved name, or an earlier extension hides, and any grant that no extension declares anymore.
- `ncly --help` shows extensions in their own section, with the description from their manifest.

**Protocol.** `ncly` runs an extension as the Programs section of cli-spec.md defines: its environment, its grants, and its cancellation. The extension follows the agent contract. Its error codes carry its domain as a prefix, and it translates its own text into the language of `NCLY_LANG`.

**New error code.**

| Code | Exit | When |
|---|---|---|
| `KEY_NOT_GRANTED` | 78 | An extension declares a key that no human granted to it |

A Python extension can be one file with the shebang `#!/usr/bin/env -S uv run --script` and its dependencies inline.

**Manifest.** A `nicely.toml` next to an extension or a `SKILL.md` declares what it needs. `doctor`, help, and completion read it.

```toml
domain = "fleet"
description = "Sync repositories across Pascal's machines"

[requires]
bins = ["git", "ssh"]
keys = []
```

Skills move their requirements from frontmatter to `nicely.toml` when they join a tap.

**Taps.**

```
ncly tap add <owner/repo | git-url> [--dry-run]
ncly tap list [--json]
ncly tap sync [<tap>] [--dry-run]
ncly tap remove <tap> [--dry-run] [--force]
```

- `owner/repo` means a GitHub repository. Any Git URL works.
- The shared config lists the taps. `add` writes the tap into it and clones the tap to `~/.local/share/nicely/taps/<owner>/<repo>/` with the user's own Git credentials. Nicely handles no Git authentication.
- `add` lists the keys that the tap's extensions declare and asks the human to grant each one. Without a terminal, `add` grants nothing, and `--force` never grants a key.
- `sync` clones each tap of the shared config that this machine lacks and pulls the others, so a new machine reaches the same setup. When an update declares a new key, `sync` warns, and the extension does not start until a human grants the key.
- `remove` takes the tap out of the shared config and moves its clone to the trash.
- A tap holds `skills/<name>/SKILL.md` and `extensions/ncly-<domain>`.
- Tap skills join `ncly skill list`. When two sources hold the same name, `ncly skill view <tap>/<name>` picks one.

**Skill link.**

```
ncly skill link [<name>] [--harness <name>] [--dry-run]
```

It symlinks skills into the skill folder of each harness, one line per harness in the config:

```toml
[skill.link]
claude = "~/.claude/skills"
```

**First personal extension.** Pascal's `sync-fleet` script becomes `ncly-fleet` in his private tap, called as `ncly fleet sync`.

**Follow-up in `pascalandy/skills`.** The `sync`, `install-skills`, and `sync-fleet` recipes can retire once taps and `skill link` cover them. Pascal decides when.

## To define when M3 starts

- The order of skill sources between `[skill] paths` and taps.
- The editor for config files. `tomledit` keeps comments and formatting.
- Whether a manifest description carries translations.
- The command that grants and revokes a key, such as `ncly auth grant <service> <domain>`.
- Where grants live: in the shared config, so a new machine asks nothing again, or on each machine.
- How Pascal's Python scripts move from the `{"ok":false,"errors":["…"]}` strings of his script-output convention to the error objects of D027.

## Out of scope

- A public registry or marketplace of extensions.
- Signed extensions.
- `ncly skill pack` and `ncly skill run`: [M99](M99-parking-lot.md).

## Acceptance

Scenarios build a local Git repository as a fixture tap and add it with a file path URL.

```
# A tap brings its skills
exec ncly tap add file://$WORK/tap-fixture
exec ncly skill list --json
stdout '"name":"fixture-skill"'

# An extension becomes a domain and receives the protocol
exec ncly hello --json
stdout 'NCLY_JSON=1'

# An extension never sees a key that no human granted to it
env DEEPGRAM_API_KEY=test-key
exec ncly hello --json
! stdout 'DEEPGRAM_API_KEY'

# An update that declares a new key gets nothing until a human grants it
exec git -C $WORK/tap-fixture merge -q --ff-only asks-for-key
exec ncly tap sync
exits 78 ncly hello --json
stderr '"code":"KEY_NOT_GRANTED"'

# A built-in wins over an extension with the same name
exec ncly doctor extensions --json
stdout 'ncly-doctor'
stdout '"status":"warn"'
```

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] `ncly fleet sync` runs from Pascal's private tap on macOS and on Omarchy
- [ ] `ncly skill link` replaces `just install-skills` on one machine
- [ ] v0.3.0 is tagged and released
