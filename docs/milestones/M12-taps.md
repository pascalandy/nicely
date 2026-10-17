# M12 Taps

Status: ready
Version: v0.12.0

## Demo

A user installs `agent` from the official tap, syncs a fixture tap's skills and extensions, links its skills, and removes it safely.

## Scope

A tap carries skills and extensions. [ncly tap](../north-star/core-spec.md#ncly-tap) defines its commands, formats, release selection, publication, and recovery.
The Nicely repository remains the official tap, as [D041](../north-star/decisions/D041-first-party-extensions.md) decides.

```
ncly tap add <owner/repo | git-url> [--dry-run]
ncly tap list [--json]
ncly tap sync [<tap>] [--dry-run]
ncly tap remove <tap> [--dry-run] [--force]
ncly skill link [<name>] [--harness <name>] [--dry-run]
```

- Tap IDs distinguish hosts and map to local storage, as [Tap identity and storage](../north-star/core-spec.md#tap-identity-and-storage) defines
- `add`, `list`, and `remove` manage the configured source, Git clone, and installed publication through [Configuration and commands](../north-star/core-spec.md#configuration-and-commands)
- `sync` updates Git executables and release archives through [Extension packages and releases](../north-star/core-spec.md#extension-packages-and-releases). The official tap follows the installed `ncly` release
- Complete publications, config edits, and links follow [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), with ownership checks and interruption evidence
- Extension discovery pins one publication per call. Tap grants bind its ID and domain, as [Discovery and grants](../north-star/core-spec.md#discovery-and-grants) defines
- Tap skills and harness links follow [ncly skill](../../extensions/skill/spec.md#ncly-skill) and [ncly skill link](../../extensions/skill/spec.md#ncly-skill-link)
- `list` retains the aligned tap columns and indented extension lines that Pascal picked in [M07](M07-forms.md)
- README.md and the `nicely` skill invite users to add the official tap. An interactive invitation stays in [M99](M99-parking-lot.md)

A public registry of extensions and signed extensions stay out of scope.

Scenarios build a local Git repository as a fixture tap and add it with a file path URL. A counted extension stub and filesystem checks prove dry run and retry behavior. Metadata alone does not prove conformance.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M12-T1 | Tap add, list, and remove | agent | — | todo |
| M12-T2 | Tap sync | agent | M12-T1 | todo |
| M12-T3 | Discover tap contents | agent | M12-T2 | todo |
| M12-T4 | Skill link | agent | M12-T3 | todo |
| M12-T5 | The official tap | agent | M12-T2, M12-T3 | todo |

### M12-T1 Tap add, list, and remove

- **Read:** [ncly tap](../north-star/core-spec.md#ncly-tap), [Configuration](../north-star/contract.md#configuration), [Retry safety](../north-star/contract.md#retry-safety), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_add.txtar`, `testdata/script/tap_remove.txtar`, `testdata/script/tap_recovery.txtar`

`internal/cli/` owns tap commands and initial Git publication. `sdk/config/` edits `tap.sources` with `tomledit`. `sdk/platform/` supplies locks, durable writes, and trash.
These packages reach only lower layers. Add the trash helper here, with recoverable evidence for each move.
Prove host-separated IDs, GitHub aliases, comment preservation, local overrides, aligned listing, dry run, and repeated add and remove.
Boundary kills prove config and publication recovery, edit conflicts, and one trash move. T4 adds the harness-link removal cases, and T5 adds archives.

### M12-T2 Tap sync

- **Read:** [Configuration and commands](../north-star/core-spec.md#configuration-and-commands), [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), [D035](../north-star/decisions/D035-sync-ownership.md), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_sync.txtar`, `testdata/script/tap_sync_boundary.txtar`

`internal/cli/` fetches and validates Git updates with the initial publication code. It uses `sdk/platform/` for publication locks and durable intent.
Prove one complete publication per tap, stale-file removal, unchanged sync, and conflicts that preserve edits.
Kill each boundary and repeat the invocation. Count effects after one tap succeeds and a later tap hits a lock or network failure.
The overall exit permits 75 only when every prior effect is safe to repeat. T5 extends the same update path to release archives.

### M12-T3 Discover tap contents

- **Read:** [Discovery and grants](../north-star/core-spec.md#discovery-and-grants), [Extensions](../north-star/contract.md#extensions), [Grants](../north-star/contract.md#grants), [ncly skill](../../extensions/skill/spec.md#ncly-skill), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_extensions.txtar`, `testdata/script/tap_revision.txtar`, `extensions/skill/testdata/script/skill_taps.txtar`

`internal/cli/` adds taps to host discovery and grant-source matching. It reuses M08's grant confirmation for interactive `add`.
`extensions/skill/` reads standalone tap skills through `sdk/config/` and installed extension skills through `ncly describe`, available since M03.
The extension never imports `internal/`. Prove selected and hidden origins, tap order, the developer copy's precedence, and qualified skill selection.
Prove that updates keep grants, other sources inherit none, and new keys receive nothing without human confirmation.
Read an existing path grant and a tap grant through independent consumer cases. No conversion may grant access.
Publish revision B between revision A's manifest loading and process launch.
The counted stub must run A with A's catalogs, then B on the next call.

### M12-T4 Skill link

- **Read:** [ncly skill link](../../extensions/skill/spec.md#ncly-skill-link), [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), [code map](../north-star/architecture.md#code-map)
- **Proves:** `extensions/skill/testdata/script/skill_link.txtar`, `testdata/script/tap_remove_links.txtar`

`extensions/skill/` uses `sdk/config/` for harness destinations and `sdk/platform/` for link publication and evidence.
`internal/cli/` checks that stored evidence before tap removal, without importing the extension.
Prove unchanged links, source switches, bundled `nicely`, and user-file conflicts. Kill each link-update boundary and repeat the call.
Removal preserves a user-modified link and its config entry. Otherwise it removes managed links before the tap's trash move.
Dry run proves all these paths without an edit, link replacement, or trash move.

### M12-T5 The official tap

- **Read:** [Extension packages and releases](../north-star/core-spec.md#extension-packages-and-releases), [D041](../north-star/decisions/D041-first-party-extensions.md), [D048](../north-star/decisions/D048-static-manifest-and-catalogs.md), [D054](../north-star/decisions/D054-extension-skills-through-describe.md), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_official.txtar`, `testdata/script/tap_archives.txtar`, release inspection in `tools/cmd/release-artifacts/`

`internal/cli/` adds release-archive preparation to tap add and sync. `.goreleaser.yaml` and `.github/workflows/release.yml` publish the first-party extension archives.
`tools/cmd/release-artifacts/` verifies each system's executable, manifest, catalogs, skill, and checksum coverage. Packaging reads extension-owned assets without adding a core domain dependency.
With no checkout or Go on `PATH`, add the official tap and run `agent` from the installed `ncly` tag's archive.
Prove all four system selections, checksum and manifest failures before publication, same-version sync, and sync after an `ncly` upgrade.
Prove that other taps follow their latest stable release. Count launches to rule out builds and install scripts.
Update README.md and the `nicely` skill text with the recommended install command. Every later first-party extension ships through this same archive path.

## After this milestone

These steps live outside this repository and block no card. Pascal decides when.

- An agent turns Pascal's `sync-fleet` script into `ncly-fleet` in his private tap, his first personal extension, and Pascal runs `ncly fleet sync` on macOS and on Omarchy
- `ncly skill link` replaces `just install-skills` on one machine, and the `sync`, `install-skills`, and `sync-fleet` recipes of `pascalandy/skills` retire
