# M06 Auth

Status: ready
Version: v0.6.0

## Demo

With an extension that declares the `deepgram` service, `ncly auth login deepgram --stdin` stores its key in the keychain, and `ncly auth status` and `ncly doctor auth` report that the key comes from the keychain, and `verify-nicely` repeats the login with a test service on the real keychain of Omarchy.

## Scope

Implement [Keys](../north-star/contract.md#keys), [ncly auth](../north-star/core-spec.md#ncly-auth), and [Required keys](../north-star/core-spec.md#required-keys). Core declares no service of its own: each service comes from the `requires.keys` of an installed manifest, which [M03-T2](M03-describe.md#m03-t2-command-detail) brings, and any name works the same way. `auth` is the first command that writes. Until the masked field and the confirmation form of [M07](M07-forms.md), login reads a key through a plain prompt without echo, and a confirmation is a plain `y/N` prompt. Grants, which pass a key to an extension, arrive in [M08](M08-grants.md), so no answer of this milestone holds `grants`.

The keychain lives in `internal/keychain`, at layer 0 of the [code map](../north-star/architecture.md#code-map), and only `internal/cli` reaches it. It wraps go-keyring with the time limits of Keys. `cmd/ncly` passes core the keychain's backend in its call, as it passes the bundled extensions: go-keyring in `ncly`, and a test keychain in the test binary alone. The test keychain is a file under `$HOME`, outside Nicely's cache, because go-keyring's mock lives in one process while a scenario runs `ncly` several times, and because `unchanged` then sees a write. It records each call in that file, so a scenario counts the calls. `NCLY_TEST_KEYCHAIN=unavailable`, which only the test binary reads, makes it refuse every call at once, `NCLY_TEST_KEYCHAIN=hang` makes it never answer, and `NCLY_TEST_KEYCHAIN=write-hang` makes it answer every read, then apply a write to its file and never answer that write. The test binary shortens the keychain's limits, as it shortens the 15 seconds of cleanup, so a scenario proves a limit without waiting 10 seconds.

The scenarios declare services in the manifest of the shell fixture of M01.

`auth` is also the first command that reaches a real system that the scenarios replace, the keychain. M06-T1 therefore adds `verify-nicely`, the skill that drives the real `ncly`, before the `auth` cards. Each later card adds its feature to the skill's map. The `auth` feature logs a test service in and out of the real keychain, and never touches the key of a real service.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M06-T1 | Verification skill | agent | — | todo |
| M06-T2 | Keychain and status | agent | M06-T1 | todo |
| M06-T3 | Login | agent | M06-T2 | todo |
| M06-T4 | Logout and service completion | agent | M06-T3 | todo |
| M06-T5 | Doctor auth | agent | M06-T2 | todo |

### M06-T1 Verification skill

- **Read:** [Work on a card](../../AGENTS.md#work-on-a-card), [Discovery](../north-star/contract.md#discovery), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [D015](../north-star/decisions/D015-macos-and-linux.md)
- **Proves:** the run, quoted in the pull request, of an agent that did not write the change, which builds `ncly`, runs `ncly doctor`, drives one feature of the map end to end, cleans up, and finds its evidence after the cleanup

Scenarios prove each command in a sandbox, with the test keychain and stubs for harnesses and paid services. They cannot prove the real keychain, a real terminal, or a real harness. Add the project skill `.agents/skills/verify-nicely/` with the `create-verification-skill` skill, so an agent runs the real `ncly` the way a user does. The skill covers:

- **Launch:** a new recipe, `just verify`, runs the skill's script. The script builds `ncly` into a temporary folder, points `HOME` and the XDG folders at a scratch folder, and installs the Go fixture `hello` there as an external extension, with its manifest, catalogs, and skill beside it. It runs on macOS and on Linux, as D015 requires
- **Doctor:** `ncly --version` names the build, and `ncly doctor` passes
- **Drive:** `features/` holds one file per feature, for `describe`, `skill`, `doctor`, `completion`, and the extension host through `hello`. Each file names the commands that reach the feature and the end state that proves it
- **Evidence:** the transcripts go to a folder outside the checkout, which the cleanup keeps, and the pull request quotes them
- **Who runs it:** an agent that did not write the change, on a small model of the other family, at effort `xhigh`. GPT-6-Luna verifies the work of a Claude model, and Claude Haiku 5.5 verifies the work of a Codex model. The agent shares the implementer's checkout, so it sees uncommitted changes. The implementer updates the feature files first. Before a run that bills or writes to a real keychain, the implementer asks Pascal and passes the consent on, because a delegate cannot ask

In [Work on a card](../../AGENTS.md#work-on-a-card), add step 6, Verify, between Checks and Docs, in the checklist and in the numbered steps below it. Its line says that an agent that did not write the change runs `verify-nicely` on the feature files that the card updated, and that its evidence is in the pull request. A card that changes nothing a user sees marks the step `n/a`. Add the skill to the table of [Find what you need](../../AGENTS.md#find-what-you-need).

### M06-T2 Keychain and status

- **Read:** [Keys](../north-star/contract.md#keys), [ncly auth](../north-star/core-spec.md#ncly-auth), [Output](../north-star/contract.md#output), [Discovery](../north-star/contract.md#discovery), [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill)
- **Proves:** `testdata/script/auth_status.txtar`, `extensions/skill/testdata/script/skill_nicely.txtar`

Add `internal/keychain`, with the time limits of Keys, 60 seconds in interactive mode and 10 seconds otherwise, the rule that a command asks an unavailable keychain nothing more, and the test keychain of the Scope. Add `auth status` to core in `internal/cli`, as a report command whose findings are `services`, through the report outcome of the SDK. It lists the services that the `requires.keys` of the manifests that discovery finds declare, reads each key from its variable first, and asks the keychain once whether it answers, then only for a service whose variable is empty. Keep the verdict of each service in one function, which doctor reuses in M06-T5. Register `AUTH_MISSING` and `KEYRING_UNAVAILABLE` in the registry of `sdk/contract`. Add `auth status` to the `nicely` skill in `extensions/skill`, within its 60 lines: it lists the keys that installed extensions need and where each one comes from.

`auth_status.txtar` gives two commands of the shell fixture the same service and a third its own, so the first service appears once. It covers:

- No declared service, which answers an empty `services`
- A key from the variable, whose keychain entry the call log shows unread, and a key from the test keychain
- A key in neither, which fails with `AUTH_MISSING`, exit 78, the entry's hint and `next`, the findings on stdout, and no `errors`
- `unavailable` with every key in the environment, which stays `ok` with `keychain` set to `unavailable`, and `unavailable` with an empty variable, which fails with `KEYRING_UNAVAILABLE` and no `next`
- `hang`, which counts the keychain unavailable after the shortened limit, and asks it once, as the call log shows
- An invalid manifest, whose services never appear

No key appears on stdout or stderr. The scenario runs once under the pseudo-locale.

### M06-T3 Login

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [Modes](../north-star/contract.md#modes), [Global flags](../north-star/contract.md#global-flags), [Keys](../north-star/contract.md#keys), [D057](../north-star/decisions/D057-auth-edge-cases.md), [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill)
- **Proves:** `testdata/script/auth_login.txtar`, `testdata/script/auth_dry_run.txtar`, `extensions/skill/testdata/script/skill_nicely.txtar`

Add `auth login` to core in `internal/cli`, with `--stdin` and `--force`. It rejects an invalid service name, `keychain` included, with `USAGE_INVALID` before any other step, and asks the keychain before it reads a key or asks for a confirmation. Its key prompt follows the echo rule of [ncly auth](../north-star/core-spec.md#ncly-auth), and its `y/N` confirmation takes its words from core's catalog. Its declaration sets `dry_run` in `modes`, so help and `ncly describe` show `--dry-run` from this card on, as [Global flags](../north-star/contract.md#global-flags) requires of every command that writes. The dry run shares the preparation of execution, the name check included, as [ncly auth](../north-star/core-spec.md#ncly-auth) defines. Add the rule to the `nicely` skill: on `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `TERMINAL_REQUIRED`, the agent relays the hint and `next` to the human, and never types or pipes a key itself.

The test binary gains a `pty` command, which runs a command with stdin and stdout on a pseudo-terminal, writes each line of its input only after the prompt that it names appears, and prints what the terminal showed. `ttyin` leaves stdout a pipe, so [Modes](../north-star/contract.md#modes) would make the call non-interactive, and `script` writes its input before `ncly` turns echo off, so the key would show. Under `pty`, `auth_login.txtar` proves that the typed key never appears in the terminal's output while the test keychain holds it, that a declined replacement keeps the old key and exits 2 with `CONFIRMATION_REQUIRED`, and that Ctrl-C at the prompt exits 130 with echo restored. It also covers:

- `--stdin` with the final newline of `echo` and with spaces around the key, which store the trimmed key, and an empty or blank input, which stores nothing and exits 2 with `USAGE_INVALID`
- `--stdin` with a terminal on stdin, through `ttyin -stdin`, which exits 2 with `USAGE_INVALID`
- No terminal and no `--stdin`, which gives `TERMINAL_REQUIRED` with its hint, even with `NCLY_TEST_KEYCHAIN=unavailable`
- `unavailable` with `--stdin`, which gives `KEYRING_UNAVAILABLE`, whose hint names the variable
- A replacement in non-interactive mode, which gives `CONFIRMATION_REQUIRED` and keeps the old key without `--force`, and replaces it with `--force`
- `write-hang`, which exits 1 with `RUNTIME` and the `next` `ncly auth status` once the shortened limit passes, after which `ncly auth status`, in a new process, finds the key in the keychain

Each success answers `service` and `action`, never a key. The scenario runs once under the pseudo-locale.

As the first writing command, `login` proves dry run against the keychain and the files in `auth_dry_run.txtar`, with `snapshot` and `unchanged`, which see the test keychain because it lives under `$HOME`. A dry run needs neither a terminal nor `--stdin`, reads no stdin, prompts for nothing, and reports `auth.keychain` as pending. The call log of the test keychain stays empty, and under `hang` the dry run answers at once. An invalid service name still exits 2 with `USAGE_INVALID`.

### M06-T4 Logout and service completion

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [Modes](../north-star/contract.md#modes), [Global flags](../north-star/contract.md#global-flags), [ncly completion](../north-star/core-spec.md#ncly-completion)
- **Proves:** `testdata/script/auth_logout.txtar`, `testdata/script/auth_dry_run.txtar`, `testdata/script/completion.txtar`

Add `auth logout` to core in `internal/cli`, with `--force`, the confirmation of M06-T3, and the name check of login. Its declaration sets `dry_run` too, and `auth_dry_run.txtar` gains the proof of M06-T3 for `logout`. `auth_logout.txtar` covers removal with `--force`, a confirmed and a declined prompt under `pty`, `KEYRING_UNAVAILABLE`, whose hint asks to unlock the keychain, and a service without a stored key, which exits 2 with `NOT_FOUND` before any confirmation, with the `next` `ncly auth status`, and with a hint that names the variable when the variable holds a key. Under `write-hang`, logout exits 1 with `RUNTIME` and the `next` `ncly auth status`, after which `ncly auth status`, in a new process, no longer finds the key. The scenario runs once under the pseudo-locale.

Completion suggests the declared services after `ncly auth login` and `ncly auth logout`, from the manifests that discovery finds, as `status` lists them, without asking the keychain or starting an extension. Core knows these values itself, so they need no list command, and the argument completes through the dynamic completion that [M04-T5](M04-skill.md#m04-t5-skill-name-completion) brings to `internal/cli`. `completion.txtar` proves that a typed prefix narrows the services, and that a service of an invalid manifest never appears.

### M06-T5 Doctor auth

- **Read:** [Required keys](../north-star/core-spec.md#required-keys), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [ncly auth](../north-star/core-spec.md#ncly-auth), [Keys](../north-star/contract.md#keys)
- **Proves:** `testdata/script/doctor_auth.txtar`

Add the `auth` component to `ncly doctor` in `internal/cli`, built on the verdict of each service from M06-T2, so `ncly doctor auth` and `ncly auth status` never disagree. Register the warning `KEYRING_ENV_ONLY` in `sdk/contract`. Doctor runs its `auth` component last, after every other check, and gives each keychain call the smaller of the keychain's limit and the time left of `--timeout`.

`doctor_auth.txtar` follows the report of M05: the checks under `checks` on stdout, exit 78 when one fails, and no `errors`. It covers:

- No declared service, where the component holds `auth.keychain` alone
- A pass from the variable and one from the keychain, whose messages name the source and never the key
- `AUTH_MISSING`, with the hint and `next` `ncly auth login <service>`
- `unavailable` with every key in the environment, a `warn` with `KEYRING_ENV_ONLY` and exit 0, and `unavailable` with an empty variable, where `auth.<service>` fails with `KEYRING_UNAVAILABLE` and no `next`
- `hang` with a `--timeout` longer than the shortened limit, which warns, and `hang` with a shorter `--timeout`, where the checks of the shell fixture's component still pass, and only the `auth` checks fail with `CHECK_TIMEOUT`

`ncly doctor` without a component lists the `auth` checks beside the others, and the scenario compares the verdict of each service with `ncly auth status`. It runs once under the pseudo-locale.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal runs `verify-nicely` on the `auth` features on his Mac, where the macOS keychain may ask for his password
