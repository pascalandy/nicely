# M19 One declaration

Status: planned
Version: v0.19.0

## Demo

A cold agent's dry run gives its key's grant command, `ncly describe` lists a command's codes, and an author changes a default in one place.

## Scope

[D033](../north-star/decisions/D033-one-declaration.md) is the spine: one declaration supplies each surface that an agent reads, and no handler, extension spec, or skill copies it.
In the agent drives of 2026-10-21, a cold operator made 85 calls and read 72,224 bytes.
In the author drive, one command touched 8 files, and its 13 facts sat in 63 places.
Core gains no command: each card completes a projection of the declaration or deletes a copy.
A card that grows a file over the cap of `just limits`, such as `core-spec.md`, first moves the section into a folder named after the file.
Each card removes the markers that name it, such as "from M19-T1", once it ships.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M19-T1 | Check keys and grants in the dry run | agent | — | todo |
| M19-T2 | Describe the codes that a command answers | agent | — | todo |
| M19-T3 | Give each usage error its fix | agent | — | todo |
| M19-T4 | Read declared defaults, and name each rule failure, in the SDK | agent | — | todo |
| M19-T5 | Hold every extension to one bar | agent | M19-T2 | todo |

### M19-T1 Check keys and grants in the dry run

- **Read:** [Grants](../north-star/contract/grants.md), [Keys](../north-star/contract/keys.md), [Dry-run answer](../north-star/contract/dry-run-answer.md), [Operations](../north-star/contract/operations.md), [Required keys](../north-star/core-spec.md#required-keys), [ncly auth](../north-star/core-spec.md#ncly-auth), [SDK](../north-star/contract/sdk.md), [D033](../north-star/decisions/D033-one-declaration.md), [D058](../north-star/decisions/D058-grants-by-path.md)
- **Proves:** `testdata/script/dry_run_keys.txtar`, `testdata/script/doctor_auth.txtar`, `testdata/script/auth_status.txtar`, and `extensions/transcript/testdata/script/transcript_dry_run.txtar`. They cover a dry run whose key has no grant, and one whose key has a grant, which keeps its check pending. The call log of the test keychain stays empty. They also cover ncly doctor transcript with its key check, and a key both missing and ungranted

In the operator drive, a dry run without a key or a grant answered `ok`, and `ncly doctor transcript` skipped `auth.deepgram`.
The hint of `ncly auth status` named the login alone, although the grant was missing too.

A dry run answers the code of a declared key that the call cannot get, with its hint and `next`, as the real call would.
A Go extension gets that problem from the SDK.
A dry run never reads the keychain or a key's value, as [Keys](../north-star/contract/keys.md) defines, so of a key it knows only the grant.
It answers `KEY_NOT_GRANTED` when no human granted the key at the extension's path, as execution does.
A granted key keeps its check pending, because only execution reads the keychain and sends the key.

`ncly doctor <domain>` includes the key checks of that domain, which read the keychain, as [Required keys](../north-star/core-spec.md#required-keys) defines.
When doctor or `ncly auth status` finds a key both missing and ungranted, the hint names `ncly auth login`, then `ncly auth grant`.
Its `next` holds the login, because a `next` names one command, as [D053](../north-star/decisions/D053-next-command.md) decides.

### M19-T2 Describe the codes that a command answers

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Error codes](../north-star/contract/error-codes.md), [Manifest](../north-star/contract/manifest.md), [Compatibility](../north-star/contract/compatibility.md), [SDK](../north-star/contract/sdk.md), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `docs/north-star/core-spec/`, `testdata/script/describe_command.txtar`, `testdata/script/describe_schema.txtar`, and `testdata/extensions/hello/testdata/script/sdk_extension.txtar`. They cover the codes and exit codes of a command in ncly describe --json. A handler that answers a code that the manifest lacks exits 1 with RUNTIME

In the operator and author drives, `ncly describe` showed no error code, so only a skill told an agent the codes of an extension.
A handler that answered an undeclared code panicked.

List in `ncly describe <path> --json` the codes that each command may answer, with their exit codes, from the declaration and the manifest's `errors`.
The new key is additive, as [Compatibility](../north-star/contract/compatibility.md) allows.
Make the SDK answer `RUNTIME` for a code that the manifest lacks, with a message that names the missing `[[errors]]` entry, instead of a panic.
Remove the lists of codes from the skills, and keep a code only where a skill says when to act on it.
Move the `ncly describe` section into `docs/north-star/core-spec/` before you grow it.

### M19-T3 Give each usage error its fix

- **Read:** [Usage](../north-star/contract/usage.md), [Output](../north-star/contract/output.md), [Command descriptions](../north-star/contract/command-descriptions.md), [ncly agent](../../extensions/agent/spec.md#ncly-agent), [D053](../north-star/decisions/D053-next-command.md), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `testdata/script/usage_errors.txtar`, `testdata/script/help.txtar`, and `extensions/agent/testdata/script/agent_dry_run.txtar`. They cover the accepted values of a flag, and the closest name of a mistyped command or flag. They also cover the next of a usage error in JSON mode, and the agent entry of the root help. Each example of agent run passes in a dry run

In the operator drive, 11 of 22 usage errors named a help page instead of the fix, and `--effort huge` named no accepted value.
`ncly describe` showed `--tools` as optional, although `agent run` fails without it, and its examples failed as written.

Give a usage error the accepted values of its flag and the closest command or flag name.
In JSON mode, its `next` is `ncly describe <path> --json`, never a help page.
Name the agent entry in the root help: `NCLY_JSON=1`, `NCLY_NO_INPUT=1`, and `ncly skill view nicely`.
Declare what `agent run` requires, so `ncly describe` and its first answer agree, and make its examples run as written.

### M19-T4 Read declared defaults, and name each rule failure, in the SDK

- **Read:** [SDK](../north-star/contract/sdk.md), [Manifest](../north-star/contract/manifest.md), [Test kit](../north-star/contract/test-kit.md), [Text of an extension](../north-star/contract/text-of-an-extension.md), [D033](../north-star/decisions/D033-one-declaration.md)
- **Proves:** `sdk/extension/`, `sdk/extensiontest/extensiontest_test.go`, `testdata/extensions/standalone/testdata/script/defaults.txtar`, and `template/`. They cover a manifest default that reaches a handler, and each rule failure of extension.Check as an English sentence. They also cover an answer whose message is a bare catalog ID, and a fake program that answers its version probe

In the author drive, a default changed in the manifest alone and a mistyped catalog ID both passed `just check`.
Three failures of `extension.Check` printed a raw catalog ID, and the fake program of the kit could not answer a version probe.

Give `extension.Call` typed accessors that return the manifest default, so a handler never restates it.
Make `extension.Check` print the English sentence of each rule failure, never a catalog ID.
Make the test kit fail an answer whose `message` or `hint` is a bare catalog ID.
Let its fake program answer the version probe that a manifest declares.
Use the accessors in the template.

### M19-T5 Hold every extension to one bar

- **Read:** [Test kit](../north-star/contract/test-kit.md), [What a great extension does](../north-star/contract/extensions.md#what-a-great-extension-does), [SDK](../north-star/contract/sdk.md), [D063](../north-star/decisions/D063-sdk-test-kit.md), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `extensions/agent/agent_test.go`, `extensions/skill/skill_test.go`, `extensions/transcript/transcript_test.go`, `justfile`, `scripts/template-test.sh`, and `extensions/skill/testdata/script/skill_nicely.txtar`. Each first-party extension passes the static checks of the test kit, and just install writes the sidecars through extension.Export. The rendered template passes just check without its example command, and the nicely skill gives a release tag

The maintainer drive found that first-party extensions run `extension.Check` alone, beside a separate skill check in core.
Their `just install` copies the sidecars from the source tree.
The author needed 11 files and 3 failed runs to remove the example command.
The `--vcs-ref` that the skill gives failed on a source build.

Run the static checks of the test kit on each first-party extension, and retire core's `mentions` check.
`mentions` also checks the codes that a skill names, so it retires after M19-T2 removes the lists of codes.
Make `just install` write the sidecars through `extension.Export`, as the kit does.
Say in the template's `AGENTS.md` how to remove the example command, and move the shared usage helper out of `inspect.go`.
Make the `nicely` skill give the release tag for `--vcs-ref`, never the pseudo-version of a source build.

## After this milestone

Once v0.19.0 ships, rerun the operator, author, and maintainer drives against the release.
Compare their calls, bytes, files touched, and places per fact with the baseline in [Scope](#scope).
