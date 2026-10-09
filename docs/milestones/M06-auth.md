# M06 Auth

Status: planned
Version: v0.6.0

## Demo

`ncly auth login deepgram --stdin` stores a key in the keychain, and `ncly auth status` and `ncly doctor auth` report where each key comes from.

## Scope

Implement [Keys](../north-star/contract.md#keys) and [ncly auth](../north-star/contract.md#ncly-auth). Core declares no service of its own: any service name works the same way. The masked field and the confirmation form arrive in [M07](M07-forms.md), and grants, which pass a key to an extension, in [M08](M08-grants.md).

The test keychain is stored in a file under `$WORK` and registered only in the test binary, because go-keyring's mock lives in one process and a scenario runs `ncly` several times. `NCLY_TEST_KEYCHAIN=unavailable`, which only the test binary reads, makes it stop answering.

## Open questions

Settle each one in [ncly auth](../north-star/contract.md#ncly-auth), then delete this section.

- Which services `auth status` lists now that core declares none: the services that installed manifests declare, plus the services that hold a key?
- What `login` does in a terminal without `--stdin` until the masked field of M07 exists: read the key with echo off, or fail with a hint that shows `--stdin`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M06-T1 | Keychain and status | agent | — | todo |
| M06-T2 | Login and logout | agent | M06-T1 | todo |
| M06-T3 | Dry run | agent | M06-T2 | todo |
| M06-T4 | Doctor auth | agent | M06-T1 | todo |

### M06-T1 Keychain and status

- **Read:** [Keys](../north-star/contract.md#keys), [ncly auth](../north-star/contract.md#ncly-auth)
- **Proves:** `testdata/script/auth_status.txtar`

The keychain through go-keyring, with a time limit of 60 seconds in interactive mode and 10 seconds otherwise, and the test keychain. `ncly auth status` reports whether the keychain answers and where each key comes from, never its value. The environment wins over the keychain.

### M06-T2 Login and logout

- **Read:** [ncly auth](../north-star/contract.md#ncly-auth), [Modes](../north-star/contract.md#modes)
- **Proves:** `testdata/script/auth_login.txtar`, `testdata/script/auth_logout.txtar`

`login --stdin` and `logout`, with a confirmation or `--force` to replace or remove a key. Without a terminal and without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED`, which wins over `KEYRING_UNAVAILABLE`. An unavailable keychain exits 78 with `KEYRING_UNAVAILABLE`, and the hint names the variable to export instead.

### M06-T3 Dry run

- **Read:** [ncly auth](../north-star/contract.md#ncly-auth), [Global flags](../north-star/contract.md#global-flags)
- **Proves:** `testdata/script/auth_dry_run.txtar`

As the first writing command, auth proves dry run against the keychain and the files with `snapshot` and `unchanged`. A dry run reads no stdin and no keychain entry, prompts for nothing, and reports `auth.keychain` as pending.

### M06-T4 Doctor auth

- **Read:** [ncly doctor](../north-star/contract.md#ncly-doctor), [Keys](../north-star/contract.md#keys)
- **Proves:** `testdata/script/doctor_auth.txtar`

The `auth` component of `ncly doctor`. A missing key fails on stdout with exit 78. When the keychain is unavailable, the check is `warn` if every needed key comes from the environment, and `fail` with `KEYRING_UNAVAILABLE` otherwise. Completion suggests service names.
