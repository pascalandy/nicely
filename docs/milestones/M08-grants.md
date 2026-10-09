# M08 Grants

Status: planned
Version: v0.8.0

## Demo

A human grants the `deepgram` key to a test extension, which then receives it, while an extension without that grant never sees it.

## Scope

Implement [Grants](../north-star/cli-spec.md#grants). A key reaches an extension only after a human grants it in a terminal, as [D030](../north-star/decision-records.md#d030-treat-extensions-as-trusted-code-and-grant-keys-one-by-one) decides. Grants live in local state and bind the key to the extension's domain and source. A replacement source needs new consent. An update that declares a new key gets nothing until a human grants it. `--force` never grants a key. Grants are consent, not a sandbox.

## Open questions

Settle each one in [Grants](../north-star/cli-spec.md#grants), then delete this section.

- The command that grants and revokes a key, such as `ncly auth grant <service> <domain>`
- The local grant record and the consent flow for a changed extension source
- The source identity of an extension found in `PATH` or the data folder, before taps exist
- Whether an extension started in a terminal asks for a missing grant, or always exits with `KEY_NOT_GRANTED` and names the grant command
- How a manifest declares an optional key, such as a harness key, which reaches the extension only when granted and present
- When core reads a granted key: before dispatch, except for a dry run, which core recognizes from the declaration, because a dry run reads no key

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M08-T1 | Grant and revoke | agent | — | todo |
| M08-T2 | Keys reach granted extensions | agent | M08-T1 | todo |
| M08-T3 | Source identity | agent | M08-T2 | todo |
| M08-T4 | Status and doctor | agent | M08-T2 | todo |

### M08-T1 Grant and revoke

- **Read:** [Grants](../north-star/cli-spec.md#grants), [Modes](../north-star/cli-spec.md#modes)
- **Proves:** `testdata/script/grant.txtar`

A human grants a key to an extension in a terminal, and revokes it. Without a terminal, granting exits 78 with `TERMINAL_REQUIRED`, and `--force` never grants. A dry run changes no grant.

### M08-T2 Keys reach granted extensions

- **Read:** [Grants](../north-star/cli-spec.md#grants), [Programs that ncly runs](../north-star/cli-spec.md#programs-that-ncly-runs), [Keys](../north-star/cli-spec.md#keys)
- **Proves:** `testdata/script/grant_dispatch.txtar`

An extension receives a granted key, from the environment or the keychain, and never a key that no human granted to it. When its manifest requires a key that is not granted, core exits 78 with `KEY_NOT_GRANTED` before starting it. A dry run reads no key.

### M08-T3 Source identity

- **Read:** [Grants](../north-star/cli-spec.md#grants)
- **Proves:** `testdata/script/grant_source.txtar`

An extension from another source, under the same name, inherits no grant. An update of the same extension that declares a new key exits with `KEY_NOT_GRANTED` until a human grants it, and keeps the grants it had.

### M08-T4 Status and doctor

- **Read:** [ncly auth](../north-star/cli-spec.md#ncly-auth), [ncly doctor](../north-star/cli-spec.md#ncly-doctor)
- **Proves:** `testdata/script/grant_status.txtar`

`ncly auth status` lists the services that extensions declare, with their grants. The `extensions` component of `ncly doctor` reports every grant that no extension declares anymore.
