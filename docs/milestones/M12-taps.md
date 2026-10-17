# M12 Install an extension from a tap

Status: planned
Version: v0.12.0

## Demo

A user without a source checkout installs one selected first-party extension from the official tap and calls it through `ncly`.

## Scope

The first tap installs one explicitly selected first-party extension from the official tap, the nicely repository on GitHub. Its executable, manifest, catalogs, and skill form one installation. Record the source and resolved revision, and validate the complete installation before exposing it to discovery.

Keep `ncly tap add` and `ncly tap list` for this path. Without `--json`, `list` keeps the look that Pascal picked in [M07](M07-forms.md): one line per tap, in aligned columns, with the tap, its source identity, and its revision, then its extensions on indented lines under it, each with its domain and path. The make-ready spec defines selection of the extension and the release artifact. Use the user's existing Git or download authentication. Nicely adds no credential manager.

A repeat add of the same installed revision is a no-op. An existing different installation or user-owned destination is a conflict. Do not overwrite it or silently upgrade it. A failed installation leaves no partially discoverable extension.

Installing never grants a key. The user calls the existing `ncly auth grant` separately after reviewing the installed extension. Grants retain the discovery-path identity of M08. A tap-specific grant identity is unnecessary for this first immutable installation.

[Taps beyond installation](M99-parking-lot.md#taps) hold arbitrary Git URLs, sync, remove, upgrades, tap skills, skill linking, source precedence, fleet migration, and cross-tap conflicts. No card below implements those operations or a transaction engine for them. `just install` remains the developer path.

## Open questions

Settle these in a bounded `ncly tap` section of [core-spec.md](../north-star/core-spec.md) before making this milestone ready.

- How the call selects one extension and its versioned release artifact, with the required architecture and platform metadata.
- How the official repository publishes that artifact. Prefer its release binaries; compare with a local build only if release packaging cannot serve the first install.
- Where a staged installation lives and how one rename makes its executable, manifest, catalogs, and skill visible together.
- Where source and revision receipts live, so an interrupted add and a repeated add cannot overwrite an unmanaged destination.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M12-T1 | Add and list one extension | agent | — | todo |
| M12-T2 | Publish a complete installation | agent | M12-T1 | todo |
| M12-T3 | Official extension install | agent | M12-T2 | todo |

### M12-T1 Add and list one extension

- **Read:** [Scope](#scope), [Configuration](../north-star/contract.md#configuration), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `testdata/script/tap_add.txtar`

Implement the settled selection, source receipt, and `list` result in `internal/cli`. A local fixture represents the download without requiring another service. Dry run describes the selected extension and destination without changing user files, config, or grants. Reject an unmanaged destination.

### M12-T2 Publish a complete installation

- **Read:** [Scope](#scope), [Discovery](../north-star/contract.md#discovery), [Grants](../north-star/contract.md#grants)
- **Proves:** `testdata/script/tap_extensions.txtar`

Stage and validate the extension as one installation before publication. An interrupted add leaves either the complete previous state or the complete new installation. Repeating an unchanged add is a no-op. A different installed revision is a conflict, with no mixed manifest and executable. Keys reach the installed extension only after its own explicit grant.

### M12-T3 Official extension install

- **Read:** [Scope](#scope), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `testdata/script/tap_official.txtar`

Publish and install the `agent` artifact through the chosen official path. Verify `ncly agent --help`, `describe`, and doctor from an installation without the source checkout. Later first-party extensions reuse this packaging. Do not add sync or a private-fleet migration to this proof.
