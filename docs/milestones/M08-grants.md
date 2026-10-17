# M08 Grants

Status: active
Version: v0.8.0

## Demo

A human grants the `deepgram` key to a test extension, which then receives it, while an extension without that grant never sees it and answers `KEY_NOT_GRANTED`.

## Scope

Implement [Grants](../north-star/contract.md#grants), the `grant` and `revoke` commands of [ncly auth](../north-star/core-spec.md#ncly-auth), and the grant checks of [Required keys](../north-star/core-spec.md#required-keys), as [D058](../north-star/decisions/D058-grants-by-path.md) decides. The extension host in `internal/cli` owns grants, as the [code map](../north-star/architecture.md#code-map) says, and writes the grants file through the durable writes and the locks of `sdk/platform`. An extension learns which declared key it lacks from `NCLY_MISSING_KEYS`, and a Go extension through `Call.Key` of `sdk/extension`. The grant prompt is the confirmation of [M07](M07-forms.md), in the look that Pascal picked.

The scenarios declare services in the manifests of shell fixtures and of the Go fixture `keyed`, which stays apart from the `hello` of M02 because verify-nicely installs `hello` in every run, and use the test keychain and the `pty` command of [M06](M06-auth.md). A fixture records a SHA-256 of each key that it receives, never the key, so no key appears on stdout or stderr.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M08-T1 | Grant and revoke | agent | — | done |
| M08-T2 | Keys reach granted extensions | agent | M08-T1 | done |
| M08-T3 | Source identity | agent | M08-T2 | done |
| M08-T4 | Status and doctor | agent | M08-T3 | done |
| M08-T5 | Optional keys | agent | M08-T4 | todo |

### M08-T1 Grant and revoke

- **Read:** [Grants](../north-star/contract.md#grants), [ncly auth](../north-star/core-spec.md#ncly-auth), [Modes](../north-star/contract.md#modes), [Global flags](../north-star/contract.md#global-flags), [ncly completion](../north-star/core-spec.md#ncly-completion), [D058](../north-star/decisions/D058-grants-by-path.md)
- **Proves:** `testdata/script/grant.txtar`, `testdata/script/auth_dry_run.txtar`, `testdata/script/completion.txtar`

Add `auth grant` and `auth revoke` to core in `internal/cli`, with the grants file of [Grants](../north-star/contract.md#grants). `grant` stores the path where discovery found the extension, so the file needs no new format when M08-T3 compares sources. Until then, a grant of a service to a domain that already holds one answers `unchanged`. Both declarations set `dry_run` in `modes`. Completion suggests the declared services after `grant` and `revoke`, as M06-T4 does for `login`, then the domains whose extension declares that service after `grant`, and the domains that hold a grant of it after `revoke`.
Add durable writes and locks to `sdk/platform/` for the grants file.

Under `pty`, `grant.txtar` covers a confirmed grant, which the grants file then holds with the fixture's path, a declined one, which exits 2 with `CONFIRMATION_REQUIRED` and changes nothing, a grant that exists, which answers `unchanged` without a prompt, and Ctrl-C at the prompt, which exits 130 and changes nothing. It also covers:

- No terminal, which exits 78 with `TERMINAL_REQUIRED`, and `--force`, which exits 2 with `USAGE_INVALID` as an unknown flag of `grant`
- An unknown domain and a service that the domain does not declare, which exit 2 with `NOT_FOUND`
- `revoke` with `--force`, after a confirmed and after a declined prompt, for a grant that does not exist, which exits 2 with `NOT_FOUND`, and for a domain that is no longer installed
- An unreadable grants file and one with another `format_version`, which fail both commands with `CONFIG_INVALID` before any change

`auth_dry_run.txtar` proves for both commands that a dry run needs no terminal, prompts for nothing, holds `auth.confirmation` in `pending_checks`, and changes nothing, as `snapshot` and `unchanged` show. `completion.txtar` proves that a typed prefix narrows the services and the domains. The scenarios run once under the pseudo-locale.

### M08-T2 Keys reach granted extensions

- **Read:** [Grants](../north-star/contract.md#grants), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Keys](../north-star/contract.md#keys), [Dispatch](../north-star/contract.md#dispatch), [SDK](../north-star/contract.md#sdk), [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill)
- **Proves:** `testdata/script/grant_dispatch.txtar`, `extensions/skill/testdata/script/skill_nicely.txtar`

When core in `internal/cli` starts an extension's command, it passes each declared key that a human granted to the extension's domain, and sets `NCLY_MISSING_KEYS`, as [Grants](../north-star/contract.md#grants) defines, with the verdict of each service of M06-T2. M08-T3 adds the comparison of sources. Register `KEY_NOT_GRANTED` in the registry of `sdk/contract`, and add `Call.Key` to `sdk/extension`. Add `KEY_NOT_GRANTED` to the `nicely` skill's rule for a missing key, within its 60 lines.

A shell fixture command records the names of the key variables that it receives, with a SHA-256 of each value, and `NCLY_MISSING_KEYS`, and a Go fixture command answers through `Call.Key`. `grant_dispatch.txtar` covers:

- A granted key from its variable and one from the test keychain, which reach the fixture, and a key exported in the shell for a service that no human granted, which never does, and which `NCLY_MISSING_KEYS` names with `KEY_NOT_GRANTED`
- A granted key in neither source, named with `AUTH_MISSING`, and under `NCLY_TEST_KEYCHAIN=unavailable`, with `KEYRING_UNAVAILABLE`, which the Go fixture answers with exit 78 and the hint and `next` of each code
- A Go fixture invocation that needs no key, which runs without it
- A dry run, a status command without `--live`, and a command that declares no key, after which the call log of the test keychain stays empty, and whose `NCLY_MISSING_KEYS` names only the services that no human granted
- A nested `ncly` call from the fixture, whose inner extension receives its own granted key from the keychain, and never a key granted only to the outer one

The scenario runs once under the pseudo-locale.

Make the `verifykeys` fixture of verify-nicely runnable. It declares `verifynicely` and reports a SHA-256 digest instead of the key. When `NCLY_MISSING_KEYS` names the service, it answers `KEY_NOT_GRANTED`. Extend the `auth` feature to run the fixture before and after granting its fake key. Add a recorded PTY step that sends the confirmation for `ncly auth grant verifynicely verifykeys` to this scratch fixture. `just verify step` has no terminal, and `grant` takes no `--force`. Never print the key.

### M08-T3 Source identity

- **Read:** [Grants](../north-star/contract.md#grants), [Discovery](../north-star/contract.md#discovery), [ncly auth](../north-star/core-spec.md#ncly-auth)
- **Proves:** `testdata/script/grant_source.txtar`

Bind each grant to its source, as [Grants](../north-star/contract.md#grants) defines: core passes a key only to the extension at the path that the grant holds, and a grant to the same domain at another path asks for confirmation with both paths, then replaces the old grant. `grant_source.txtar` covers:

- A copy of the shell fixture under the same name earlier in `PATH`, which receives none of the keys granted to the first, and whose `NCLY_MISSING_KEYS` names them with `KEY_NOT_GRANTED`
- A grant to that copy under `pty`, whose prompt shows both paths, and after which only the copy receives the key
- A new manifest at the same path, which keeps its grants, and a service that it declares for the first time, which reaches it only after a grant
- A link in `PATH` whose target changes, as a Homebrew upgrade does, which keeps the grant

### M08-T4 Status and doctor

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [Required keys](../north-star/core-spec.md#required-keys), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Grants](../north-star/contract.md#grants), [Compatibility](../north-star/contract.md#compatibility)
- **Proves:** `testdata/script/grant_status.txtar`, `testdata/script/doctor_status.txtar`, `sdk/contract/compat_test.go`

Each entry of `ncly auth status` adds `grants`, and `status` and the `auth.<service>` check of doctor fail with `KEY_NOT_GRANTED` when an extension that needs the key lacks a grant, as [ncly auth](../north-star/core-spec.md#ncly-auth) defines. Add the `auth.grants` check to doctor in `internal/cli`, register the warning `GRANT_UNUSED` in `sdk/contract`, and add a consumer case in `compat_test.go` that reads an entry of `services` with and without `grants`.

Update the `auth` feature's status and doctor checks for the required `verifynicely` key before and after its grant. With a stored key and no grant, both report `KEY_NOT_GRANTED`.

`grant_status.txtar` covers a key with every grant, a key without the grant of one of two extensions that declare it, a missing key without its grant, which reports `AUTH_MISSING` alone, a grant whose service the fixture no longer declares and one whose path no longer holds the fixture, each a `GRANT_UNUSED` entry, and an unreadable grants file, which fails `status` and the `auth.grants` check with `CONFIG_INVALID`. It compares the verdict of each service in `status` and in doctor, and runs once under the pseudo-locale.

Under `--live`, doctor reads the key of a status command within the bound of [Required keys](../north-star/core-spec.md#required-keys). `doctor_status.txtar` gains a `hello status` that declares a granted key: under `--live` it receives the key, and under `NCLY_TEST_KEYCHAIN=hang`, doctor waits once, after which the status command receives `KEYRING_UNAVAILABLE` and the `auth` checks find the keychain unavailable at once. A `--timeout` shorter than the keychain's shortened limit fails `hello.status` with `CHECK_TIMEOUT`, and an execution marker that never appears proves that the command never starts.

### M08-T5 Optional keys

- **Read:** [Manifest](../north-star/contract.md#manifest), [Grants](../north-star/contract.md#grants), [ncly auth](../north-star/core-spec.md#ncly-auth), [Required keys](../north-star/core-spec.md#required-keys), [ncly describe](../north-star/core-spec.md#ncly-describe), [Compatibility](../north-star/contract.md#compatibility)
- **Proves:** `testdata/script/extension_manifest.txtar`, `testdata/script/grant_optional.txtar`, `sdk/contract/compat_test.go`

Add the `keys` table to the manifest of `sdk/contract`, with its rules, and `optional` to the entries of `requires.keys` that `ncly describe` returns, with a consumer case in `compat_test.go` that reads an entry with and without it. `extension_manifest.txtar` proves that the manifest rejects an entry of `keys` that no command declares and an `optional` that is not a boolean.

`grant_optional.txtar` marks a service optional in the shell fixture's manifest. Without its key or its grant, `status` lists it with `optional` true and stays `ok`, doctor's `auth.<service>` passes with a message that says the key is missing, and the fixture command runs, with the service in `NCLY_MISSING_KEYS`. Under `NCLY_TEST_KEYCHAIN=unavailable`, the command still runs. A second extension that declares the same service without marking it optional still fails `status` without its grant. The scenario runs once under the pseudo-locale.
