# M08 Grants

Status: planned
Version: v0.8.0

## Demo

A human grants the `deepgram` key to a test extension, which then receives it, while an extension without that grant never sees it.

## Scope

Implement [Grants](../north-star/contract.md#grants). A key reaches an extension only after a human grants it in a terminal, as [D030](../north-star/decisions/D030-trusted-extensions.md) decides. Grants live in local state and bind the key to the extension's domain and source. A replacement source needs new consent. An update that declares a new key gets nothing until a human grants it. `--force` never grants a key. Core never refuses to start an extension over a key; the extension reports the key it lacks. Grants are consent, not a sandbox.

## Open questions

Settle each one in [Grants](../north-star/contract.md#grants), then delete this section.

- The command that grants and revokes a key, such as `ncly auth grant <service> <domain>`
- The local grant record and the consent flow for a changed extension source
- The source identity of an extension found in `PATH` or the data folder, before taps exist
- Whether an extension started in a terminal asks for a missing grant, or always exits with `KEY_NOT_GRANTED` and names the grant command
- How a manifest declares an optional key, such as a harness key, which reaches the extension only when granted and present
- How core tells an extension which of its declared keys no human granted, so the extension answers `KEY_NOT_GRANTED` rather than `AUTH_MISSING`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M08-T1 | Grant and revoke | agent | — | todo |
| M08-T2 | Keys reach granted extensions | agent | M08-T1 | todo |
| M08-T3 | Source identity | agent | M08-T2 | todo |
| M08-T4 | Status and doctor | agent | M08-T2 | todo |

### M08-T1 Grant and revoke

- **Read:** [Grants](../north-star/contract.md#grants), [Modes](../north-star/contract.md#modes)
- **Proves:** `testdata/script/grant.txtar`

A human grants a key to an extension in a terminal, and revokes it. Without a terminal, granting exits 78 with `TERMINAL_REQUIRED`, and `--force` never grants. A dry run changes no grant.

### M08-T2 Keys reach granted extensions

- **Read:** [Grants](../north-star/contract.md#grants), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Keys](../north-star/contract.md#keys)
- **Proves:** `testdata/script/grant_dispatch.txtar`

An extension receives a granted key that its command declares, from the environment or the keychain, and never a key that no human granted to it. A fixture that needs a key it did not receive answers `KEY_NOT_GRANTED` or `AUTH_MISSING` with exit 78, and one that needs no key for this invocation runs. A dry run, and a command that declares no key, read no key. A nested `ncly` call from a fixture gives the inner extension its keys from the keychain.

### M08-T3 Source identity

- **Read:** [Grants](../north-star/contract.md#grants)
- **Proves:** `testdata/script/grant_source.txtar`

An extension from another source, under the same name, inherits no grant. An update of the same extension that declares a new key receives nothing for it until a human grants it, and keeps the grants it had.

### M08-T4 Status and doctor

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [ncly doctor](../north-star/core-spec.md#ncly-doctor)
- **Proves:** `testdata/script/grant_status.txtar`

`ncly auth status` lists the services that extensions declare, with their grants. The `extensions` component of `ncly doctor` reports every grant that no extension declares anymore.
