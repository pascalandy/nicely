# M3 Extensions

Status: planned
Release: v0.3.0

## Goal

Any executable becomes an `ncly` domain, and taps carry skills and extensions to every machine. Fleet sync becomes Pascal's first personal extension.

## Depends on

M2, so extensions inherit the agent depth and the shared profiles.

## Scope

**Spec first.** Add sections for extensions, `tap`, `skill link`, and `nicely.toml` to [cli-spec.md](../north-star/cli-spec.md) before writing Go.

**Discovery.**

- An executable named `ncly-<domain>` becomes `ncly <domain>`. `ncly` looks in `~/.local/share/nicely/extensions/` first, then in `PATH` order.
- Built-in commands always win. A new `extensions` component of `ncly doctor` lists every extension it finds, where it came from, and any extension that a built-in or an earlier extension hides.
- `ncly --help` shows extensions in their own section, with the description from their manifest.

**Protocol.** `ncly` runs an extension with these environment variables, and the extension follows the agent contract of cli-spec.md.

| Variable | Value |
|---|---|
| `NCLY_VERSION` | Version of `ncly` |
| `NCLY_LANG` | The resolved language |
| `NCLY_JSON` | `1` when `--json` was given |
| `NCLY_NO_INPUT` | `1` in non-interactive mode |
| `NCLY_AGENT_DEPTH` | The current agent depth |
| Declared keys | Only the keys the manifest lists, such as `DEEPGRAM_API_KEY` |

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
ncly tap add <owner/repo | git-url>
ncly tap list [--json]
ncly tap sync [<tap>]
ncly tap remove <tap> [--dry-run] [--force]
```

- `owner/repo` means a GitHub repository. Any Git URL works.
- A tap is cloned to `~/.local/share/nicely/taps/<owner>/<repo>/` with the user's own Git credentials. Nicely handles no Git authentication.
- A tap holds `skills/<name>/SKILL.md` and `extensions/ncly-<domain>`.
- Tap skills join `ncly skill list`. When two sources hold the same name, `ncly skill show <tap>/<name>` picks one.

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

## Open questions

1. Should an extension found in `PATH` need an opt-in? Recommendation: no, the same as Git. `ncly doctor extensions` shows every extension and where it came from.
