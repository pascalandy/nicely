# M3 Extensions

Status: planned
Release: v0.3.0

## Goal

An executable with a compatible manifest becomes an `ncly` domain, and taps carry skills and extensions to every machine. Fleet sync becomes Pascal's first personal extension.

## Depends on

M1, which brings declarations, discovery, run records, and `internal/run`. M2 is not required.

## Scope

**Spec first.** Move the sections for extensions, `tap`, `skill link`, and `nicely.toml` into [cli-spec.md](../north-star/cli-spec.md) before writing Go. `link` joins the verb table.

**Discovery.**

- An executable named `ncly-<domain>` is a candidate for `ncly <domain>`. `ncly` looks in `~/.local/share/nicely/extensions/` first, then in `PATH` order. Dispatch requires a manifest with a compatible protocol. Discovery reports a missing or incompatible manifest without starting the executable
- Built-in commands and reserved names always win. A new `extensions` component of `ncly doctor` lists every extension it finds, where it came from, any extension that a built-in, a reserved name, or an earlier extension hides, and any grant that no extension declares anymore.
- `ncly --help` shows extensions in their own section, with the description from their manifest.

**Protocol.** The [Compatibility](../north-star/cli-spec.md#compatibility-m0), [Operations](../north-star/cli-spec.md#operations-m0-contract-m1-execution), and Programs sections govern dispatch. The manifest declares effects and supported modes, including dry run, recorded execution, and resume. A missing capability is unsupported. Error codes carry the domain prefix, and the extension translates its own text. Invalid JSON, a mismatched exit code, or an incompatible response is a protocol failure, never a successful run.

**New error code.**

| Code | Exit | When |
|---|---|---|
| `KEY_NOT_GRANTED` | 78 | An extension declares a key that no human granted to it |

A Python extension can be one file with the shebang `#!/usr/bin/env -S uv run --script` and its dependencies inline.

**Manifest.** A skill's `nicely.toml` and an extension's `<executable>.toml` declare their requirements using the same versioned format. For `ncly-fleet`, the manifest is `ncly-fleet.toml` beside the resolved executable. Distinct sidecar names let several extensions share a directory. `doctor`, help, completion, and `describe` read the manifest without executing the extension.

```toml
format_version = 1
contract_version = 1
domain = "fleet"
description = "Sync repositories across Pascal's machines"

[requires]
bins = ["git", "ssh"]
keys = []
```

This excerpt shows identity and requirements. Settle the complete command declarations, effects, capability fields, and result contracts before code. The schema must express multiple commands without duplicating help or completion definitions.

Skills move their requirements from frontmatter to `nicely.toml` when they join a tap.

**Taps.**

```
ncly tap add <owner/repo | git-url> [--dry-run]
ncly tap list [--json]
ncly tap sync [<tap>] [--dry-run]
ncly tap remove <tap> [--dry-run] [--force]
```

- `owner/repo` means a GitHub repository. Any Git URL works.
- A tap has a canonical source identity that distinguishes hosts and repositories. Record its resolved revision and each extension's origin. Two executables with the same domain do not share an identity merely because their names match
- The shared config lists the taps. `add` writes the tap into it and clones the tap below `~/.local/share/nicely/taps/`, keyed by its canonical source identity, with the user's own Git credentials. Nicely handles no Git authentication
- `add` lists the keys that the tap's extensions declare and asks the human to grant each one. Without a terminal, `add` grants nothing, and `--force` never grants a key.
- `sync` stages updates for the configured taps, validates their manifests, and publishes each managed destination under a lock. It records the resolved revisions. Each destination has its own atomic publication boundary; unrelated taps are not one transaction. Local edits are preserved and reported as conflicts. When an update declares a new key, the extension does not start until a human grants it
- `remove` takes the tap out of the shared config and moves its clone to the trash.
- A tap holds `skills/<name>/SKILL.md`, `extensions/ncly-<domain>`, and `extensions/ncly-<domain>.toml`
- Explicit `[skill] paths` win in their configured order, followed by taps in shared configuration order after local overrides. Qualified `ncly skill view <tap>/<name>` selects a tap skill. Discovery reports its origin, revision, and any shadowing
- Grants live in local state and bind the key to the extension's domain and canonical source. A replacement source needs new consent. Updating the same source preserves existing grants only for the keys already granted. Grants are consent, not a sandbox

**Skill link.**

```
ncly skill link [<name>] [--harness <name>] [--dry-run]
```

It symlinks skills into the skill folder of each harness, one line per harness in the config:

```toml
[skill.link]
claude = "~/.claude/skills"
```

Linking twice leaves the same managed links. A source switch updates only a link that Nicely owns and that the user has not changed. An existing user file or unmanaged link is a conflict. Tap removal checks its links first. A user-modified link blocks removal until the conflict is resolved; unchanged managed links are removed with the tap. Dry run lists those effects before any config edit or trash operation.

**First personal extension.** Pascal's `sync-fleet` script becomes `ncly-fleet` in his private tap, called as `ncly fleet sync`.

**Follow-up in `pascalandy/skills`.** The `sync`, `install-skills`, and `sync-fleet` recipes can retire once taps and `skill link` cover them. Pascal decides when.

## To define when M3 starts

- The canonical tap ID for arbitrary Git URLs, avoiding collisions between hosts, and its mapping to local storage
- The editor for config files. `tomledit` keeps comments and formatting.
- Whether a manifest description carries translations.
- The command that grants and revokes a key, such as `ncly auth grant <service> <domain>`.
- The local grant record format and consent flow for a changed extension source, using the identity rules above
- The complete manifest schema, response validation, and run-record integration for extensions that declare partial or paid work
- Recovery of an interrupted tap/config/link update, with ownership evidence and retry rules for the whole invocation
- How Pascal's Python scripts move from the `{"ok":false,"errors":["…"]}` strings of his script-output convention to the error objects of D027.
- Whether an extension run in a terminal prompts for a missing grant, or always exits with `KEY_NOT_GRANTED` and points to the grant command.

## Out of scope

- A public registry or marketplace of extensions.
- Signed extensions.
- `ncly skill pack` and `ncly skill run`: [M99](M99-parking-lot.md).

## Acceptance

Scenarios build a local Git repository as a fixture tap and add it with a file path URL. Fixture extensions have compatible manifests and return the current JSON envelope. Any environment values used by the assertions are data inside that envelope, not raw diagnostic lines.

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

Add scenarios for incompatible manifests without execution, malformed extension output, changed-source grants, deterministic skill precedence, two taps publishing the same domain, preserved local edits, and a repeated or interrupted sync. A counted extension stub and filesystem checks prove dry run and retry behavior. Metadata alone does not prove conformance.

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] `ncly fleet sync` runs from Pascal's private tap on macOS and on Omarchy
- [ ] `ncly skill link` replaces `just install-skills` on one machine
- [ ] v0.3.0 is tagged and released
