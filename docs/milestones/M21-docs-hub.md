# M21 Docs hub

Status: planned
Version: v0.21.0

## Demo

`ncly docs` gathers the docs of every project into one private local site that never leaks a key.

## Scope

Each project keeps its own `docs/` folder, and the `docs` extension gathers them, as [D022](../north-star/decision-records.md#d022-keep-docs-private-by-default) decides.

```
ncly docs scan <dir> [--dry-run]
ncly docs list [--json]
ncly docs sync [--dry-run]
ncly docs serve
ncly docs check [--json]
```

- `scan` finds every repository under `<dir>` that has a `docs/` folder and adds it to the config as `private`. It names the config file it changed, and `--dry-run` shows the change
- `list` shows each project, its path, and its visibility
- `sync` stages each included project's docs and publishes its managed copy under a lock, then rebuilds navigation. Atomic replacement applies per project and to navigation separately. The run record identifies completed and pending work if publication stops between them
- `serve` opens a local preview in the browser. Nothing leaves the machine
- `check` runs gitleaks, with any rules the user adds locally, on the hub before anything is published. It is a report command. A leak exits 78 with `DOCS_LEAK_FOUND`, because a human must review and fix it before anything moves

`scan`, `serve`, and `check` join the verb table of [Usage](../north-star/cli-spec.md#usage).

```toml
[docs]
hub = "~/code/docs-hub"
default = "private"

[docs.projects]
transcript = { path = "~/code/transcript", visibility = "public" }
client-x   = { path = "~/code/client-x" }
old-test   = { path = "~/code/old-test", visibility = "exclude" }
```

| Visibility | In the hub | In a future public build |
|---|---|---|
| `public` | yes | yes |
| `private` | yes | no |
| `exclude` | no | no |

**Ownership and repeatability.** Keep an inventory of generated copies and their fingerprints. A repeat sync with unchanged sources is a no-op. Delete stale generated copies when a source disappears or becomes `exclude`, using the trash rule. Preserve a locally edited destination and report the conflict. A source path is never a cleanup target. The configured hub must not overlap any source docs folder. Machine-specific paths belong in local config, and every config-writing command names its target file before changing it.

**Partial work.** Sync uses the operation contract. Revalidate sources, ownership, and artifacts before a resume. The overall verdict covers earlier published projects and navigation too. Dry run reports copies, replacements, removals, and unresolved conflicts without editing config or the run record.

## Open questions

Settle each one in `extensions/docs/spec.md`, then delete this section.

- The rendering engine for `serve`. The hub stays plain Markdown, so Obsidian opens it as a vault. Mintlify's local preview is the first candidate, but it needs Node, so weigh it against an engine that a stranger can run with nothing but `ncly` installed
- The scenarios: include a project and then exclude it, remove a source file, edit a generated destination, and interrupt publication before navigation is updated

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M21-T1 | Scan and list | agent | — | todo |
| M21-T2 | Sync | agent | M21-T1 | todo |
| M21-T3 | Serve | agent | M21-T2 | todo |
| M21-T4 | Check | agent | M21-T2 | todo |

### M21-T1 Scan and list

- **Read:** [Scope](#scope)
- **Proves:** `testdata/script/docs_scan.txtar`

`scan` adds projects as `private` and names the config file it changed. `list` shows each project's path and visibility.

### M21-T2 Sync

- **Read:** [Scope](#scope), [D035](../north-star/decision-records.md#d035-track-ownership-when-synchronizing-files)
- **Proves:** `testdata/script/docs_sync.txtar`

`sync` never copies an excluded project, preserves user edits and source files, and resumes an interrupted publication whose navigation references only published copies.

### M21-T3 Serve

- **Read:** [Scope](#scope)
- **Proves:** `testdata/script/docs_serve.txtar`

`serve` previews the hub locally with the chosen engine.

### M21-T4 Check

- **Read:** [Scope](#scope)
- **Proves:** `testdata/script/docs_check.txtar`

`check` reports a planted fake key on stdout and exits 78 with `DOCS_LEAK_FOUND`.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal reads the docs of all his projects in one local site
- Pascal's global agent instructions say that "docs" means `ncly docs`
