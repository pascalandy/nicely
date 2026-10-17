# M12 Taps

Status: ready
Version: v0.12.0

## Demo

`ncly tap add` brings a Git repository's skills and extensions to this machine, and installs `agent` from the official tap's release.

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
- `add`, `list`, and `remove` manage sources and storage through [Configuration and commands](../north-star/core-spec.md#configuration-and-commands)
- `sync` updates Git publications and official release archives through [Extension packages and releases](../north-star/core-spec.md#extension-packages-and-releases)
- Publications, config edits, and links follow [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), with ownership checks and repeatable effects
- Discovery selects one publication per call. Grants use paths through `current`, as [Discovery and grants](../north-star/core-spec.md#discovery-and-grants) defines
- Tap skills and harness links follow [ncly skill](../../extensions/skill/spec.md#ncly-skill) and [ncly skill link](../../extensions/skill/spec.md#ncly-skill-link)
- `list` retains the aligned tap columns and indented extension lines that Pascal picked in [M07](M07-forms.md)
- README.md and the `nicely` skill invite users to add the official tap. An interactive invitation stays in [M99](M99-parking-lot.md)

A public registry of extensions and signed extensions stay out of scope.

Scenarios build a local Git repository as a fixture tap and add it with a file path URL. A counted extension stub and filesystem checks prove dry run and retry behavior. Metadata alone does not prove conformance.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M12-T1 | Tap add and list | agent | — | todo |
| M12-T2 | Tap sync and remove | agent | M12-T1 | todo |
| M12-T3 | Discover tap contents | agent | M12-T2 | todo |
| M12-T4 | Skill link | agent | M12-T3 | todo |
| M12-T5 | The official tap | agent | M12-T2, M12-T3 | todo |

### M12-T1 Tap add and list

- **Read:** [ncly tap](../north-star/core-spec.md#ncly-tap), [Configuration](../north-star/contract.md#configuration), [Retry safety](../north-star/contract.md#retry-safety), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_add.txtar`

`internal/cli/` owns add, list, bare clones, and `git archive` publications. `sdk/config/` edits `tap.sources` with `tomledit`.
`sdk/platform/` supplies locks, writes, and boundary hooks. These packages reach only lower layers.
Prove host-separated IDs, GitHub aliases, query and fragment rejection, comment preservation, and local overrides.
Prove aligned listing, dry run, and repeated add. Kill the command at write and link-swap boundaries. Repeat add to finish from disk state.
T3 adds grants, and T5 adds official archives.

### M12-T2 Tap sync and remove

- **Read:** [Configuration and commands](../north-star/core-spec.md#configuration-and-commands), [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), [D035](../north-star/decisions/D035-sync-ownership.md), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_sync.txtar`, `testdata/script/tap_sync_boundary.txtar`, `testdata/script/tap_remove.txtar`

`internal/cli/` syncs and removes taps with T1's publication code. `sdk/platform/` adds the trash helper.
Prove current and previous retention, older-publication cleanup, unchanged sync, and edit conflicts.
Kill the command at write, link-swap, and trash boundaries. Repeat the invocation. Count writes, trash moves, and launches, including a later tap's temporary failure.
Prove whole-invocation exit 75.
Prove absent-tap `NOT_FOUND` before confirmation. T4 adds harness-link removal, and T5 adds official archive sync.

### M12-T3 Discover tap contents

- **Read:** [Discovery and grants](../north-star/core-spec.md#discovery-and-grants), [Extensions](../north-star/contract.md#extensions), [Grants](../north-star/contract.md#grants), [ncly skill](../../extensions/skill/spec.md#ncly-skill), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_extensions.txtar`, `testdata/script/tap_revision.txtar`, `extensions/skill/testdata/script/skill_taps.txtar`

`internal/cli/` adds taps to host discovery and reports grant paths through `current`. It reuses M08's grant confirmation for interactive `add`.
`extensions/skill/` reads standalone tap skills through `sdk/config/` and installed extension skills through `ncly describe`, available since M03.
The extension never imports `internal/`. Prove origins, tap order, tap precedence over the developer copy, and qualified skill selection.
Prove that updates keep path grants, other sources inherit none, and new keys require human confirmation.
Non-interactive add grants nothing, and dry run lists grants in `pending_checks`.
Publish B between A's manifest loading and launch. The counted stub runs A with A's catalogs, then B on the next call.

### M12-T4 Skill link

- **Read:** [ncly skill link](../../extensions/skill/spec.md#ncly-skill-link), [Publication and recovery](../north-star/core-spec.md#publication-and-recovery), [code map](../north-star/architecture.md#code-map)
- **Proves:** `extensions/skill/testdata/script/skill_link.txtar`, `testdata/script/tap_remove_links.txtar`

`extensions/skill/` uses `sdk/config/` for harness destinations and `sdk/platform/` for link swaps with revalidation.
`internal/cli/` removes links whose targets lie inside the tap's storage, without importing the extension or reading a state file.
Prove matching links, managed source switches, skipped bundled `nicely`, and user-file conflicts. Kill link-swap boundaries and repeat the call.
Prove removal leaves a user-modified link to a target outside tap storage alone. Dry run changes no link, config, or trash destination.

### M12-T5 The official tap

- **Read:** [Extension packages and releases](../north-star/core-spec.md#extension-packages-and-releases), [D041](../north-star/decisions/D041-first-party-extensions.md), [D048](../north-star/decisions/D048-static-manifest-and-catalogs.md), [D054](../north-star/decisions/D054-extension-skills-through-describe.md), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/tap_official.txtar`, `testdata/script/tap_archives.txtar`, release inspection in `tools/cmd/release-artifacts/`

`internal/cli/` installs archives from the installed `ncly` tag's checksum file, without a clone.
`.goreleaser.yaml` and `.github/workflows/release.yml` publish them. `tools/cmd/release-artifacts/` verifies their contents and checksums without a core domain dependency.
The test binary overrides the release base URL and the detected system. A `file://` fixture release proves all four system selections on one machine.
With no checkout, Git, or Go on `PATH`, install `agent`. Run `agent`.
Prove checksum and manifest failures, unchanged sync, and sync after an `ncly` upgrade.
Prove a build without a release version answers `NOT_FOUND`. Count launches to exclude builds and install scripts.
Update README.md and the `nicely` skill text with the recommended install command. Every later first-party extension ships through this same archive path.

## After this milestone

These steps live outside this repository and block no card. Pascal decides when.

- An agent turns Pascal's `sync-fleet` script into `ncly-fleet` in his private tap, his first personal extension, and Pascal runs `ncly fleet sync` on macOS and on Omarchy
- `ncly skill link` replaces `just install-skills` on one machine, and the `sync`, `install-skills`, and `sync-fleet` recipes of `pascalandy/skills` retire
