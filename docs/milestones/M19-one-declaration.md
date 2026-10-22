# M19 One declaration

Status: planned
Version: v0.19.0

## Demo

A dry run without a grant names the grant command, `ncly describe` lists an extension's codes, and a handler gets its defaults from the manifest.

## Scope

[D033](../north-star/decisions/D033-one-declaration.md) is the spine: one declaration supplies each surface that an agent reads, and nothing copies its values.
The agent drives of 2026-10-21 found surfaces that skip the declaration, and values copied by hand that drift.
Core gains no command: each card completes one projection of the declaration, or deletes one copy.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M19-T1 | Check the grant in the dry run | agent | — | todo |
| M19-T2 | Describe the codes of an extension | agent | — | todo |
| M19-T3 | Give each usage error its accepted values | agent | — | todo |
| M19-T4 | Pass declared defaults to the handler | agent | — | todo |

### M19-T1 Check the grant in the dry run

- **Read:** [Grants](../north-star/contract/grants.md), [Keys](../north-star/contract/keys.md), [Dry-run answer](../north-star/contract/dry-run-answer.md), [Operations](../north-star/contract/operations.md), [Required keys](../north-star/core-spec.md#required-keys), [ncly auth](../north-star/core-spec.md#ncly-auth), [SDK](../north-star/contract/sdk.md), [D033](../north-star/decisions/D033-one-declaration.md), [D058](../north-star/decisions/D058-grants-by-path.md)
- **Proves:** `testdata/script/dry_run_keys.txtar`, `testdata/script/doctor_auth.txtar`, `testdata/script/auth_status.txtar`, and `extensions/transcript/testdata/script/transcript_dry_run.txtar`. They cover a dry run whose key has no grant, and one whose key has a grant, which keeps its check pending. The call log of the test keychain stays empty. They also cover a key both missing and ungranted

In the operator drive, a dry run without a key or a grant answered `ok`, so the failure waited for the real call.
The hint of `ncly auth status` named the login alone, although the grant was missing too.

A dry run answers `KEY_NOT_GRANTED` for a declared key that no human granted at the extension's path, as execution does.
It never reads the keychain or a key's value, as [Keys](../north-star/contract/keys.md) defines, so a granted key keeps its check pending.
When doctor or `ncly auth status` finds a key both missing and ungranted, the hint names `ncly auth login`, then `ncly auth grant`.
Its `next` holds the login, because a `next` names one command, as [D053](../north-star/decisions/D053-next-command.md) decides.

### M19-T2 Describe the codes of an extension

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Error codes](../north-star/contract/error-codes.md), [Manifest](../north-star/contract/manifest.md), [Compatibility](../north-star/contract/compatibility.md), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `testdata/script/describe_command.txtar` and `testdata/script/describe_schema.txtar`. They cover the codes of an extension and their exit codes in ncly describe --json

In the operator and author drives, `ncly describe` showed no error code, so only a skill told an agent the codes of an extension.

List in `ncly describe --json` the codes that an extension's manifest declares in `[[errors]]`, with their exit codes.
A code belongs to the extension, not to one command, as [Manifest](../north-star/contract/manifest.md) defines.
The new key is additive, as [Compatibility](../north-star/contract/compatibility.md) allows.
Remove the lists of codes from the skills, and keep a code only where a skill says when to act on it.

### M19-T3 Give each usage error its accepted values

- **Read:** [Usage](../north-star/contract/usage.md), [Output](../north-star/contract/output.md), [Command descriptions](../north-star/contract/command-descriptions.md), [D053](../north-star/decisions/D053-next-command.md), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `testdata/script/usage_errors.txtar`. It covers the accepted values of a flag, and the `next` of a usage error that named a help page

In the operator drive, 11 of 22 usage errors named a help page instead of the fix, and `--effort huge` named no accepted value.

Give a usage error about a flag's value the values that the declaration accepts.
When a usage error has no concrete correction, its `next` is `ncly describe <path> --json`, never a help page.
Keep each concrete correction that a hint gives today, with its `next`.

### M19-T4 Pass declared defaults to the handler

- **Read:** [Manifest](../north-star/contract/manifest.md), [SDK](../north-star/contract/sdk.md), [ncly agent](../../extensions/agent/spec.md#ncly-agent), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `sdk/extension/`, `testdata/extensions/standalone/testdata/script/defaults.txtar`, `extensions/agent/agent.go`, and `template/`. They cover a flag that the caller did not set, which reaches the handler with its declared default

In the author drive, a default changed in the manifest alone passed `just check`, and the handler kept the old value.
The `agent` extension states one default in `manifest.toml`, `agent.go`, and `spec.md`.

Make `extension.Main` fill each flag that the caller did not set with its declared default, with no new API.
Delete each copy of a default from the handlers and the specs, such as in `agent` and the template.
Change the default rule of [Manifest](../north-star/contract/manifest.md), which leaves the default to the command today.
