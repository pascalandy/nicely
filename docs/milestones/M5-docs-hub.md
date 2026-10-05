# M5 Docs hub

Status: planned
Release: v0.5.0

## Goal

One private place to read the docs of every project. Each project keeps its own `docs/` folder, and `ncly docs` gathers them.

## Depends on

M1.

## Scope

**Spec first.** Move the `ncly docs` commands below into [cli-spec.md](../north-star/cli-spec.md) before writing Go. `scan`, `serve`, and `check` join the verb table.

**Commands.**

```
ncly docs scan <dir> [--dry-run]
ncly docs list [--json]
ncly docs sync [--dry-run]
ncly docs serve
ncly docs check [--json]
```

- `scan` finds every repository under `<dir>` that has a `docs/` folder and adds it to the config as `private`. It names the config file it changed, and `--dry-run` shows the change.
- `list` shows each project, its path, and its visibility.
- `sync` copies the docs of every project that is not `exclude` into the hub folder and rebuilds the navigation.
- `serve` opens a local preview in the browser. Nothing leaves the machine.
- `check` runs gitleaks, with any rules the user adds locally, on the hub before anything is published. It is a report command. A leak exits 78 with `LEAK_FOUND`, because a human must review and fix it before anything moves.

**Config.**

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

**Engine.** The hub stays plain Markdown, so Obsidian opens it as a vault. Pick the rendering engine for `serve` during this milestone. Mintlify's local preview is the first candidate. It needs Node, so weigh it against an engine that a stranger can run with nothing but `ncly` installed.

**The word "docs" in agent sessions.** Add this line to Pascal's global agent instructions: when Pascal says "docs", he means `ncly docs`.

## Out of scope

- A public build and hosting: [M99](M99-parking-lot.md).
- Translating docs.

## Acceptance

```
# Scan adds projects as private
exec ncly docs scan $WORK/code
exec ncly docs list --json
stdout '"visibility":"private"'

# Sync never copies an excluded project
exec ncly docs sync
! exists $WORK/hub/old-test

# Check reports a planted fake key on stdout
exits 78 ncly docs check --json
stdout '"ok":false'
stdout '"code":"LEAK_FOUND"'
```

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] Pascal reads the docs of all his projects in one local site
- [ ] v0.5.0 is tagged and released
