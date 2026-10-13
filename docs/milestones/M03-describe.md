# M03 Describe

Status: ready
Version: v0.3.0

## Demo

`ncly describe` describes core commands, a bundled extension, and an external extension in the same shape, without running any of them.

## Scope

Implement [ncly describe](../north-star/core-spec.md#ncly-describe). Grow the declarations of M00, and the manifest of M01, with the fields that the commands of M01 and M02 need. One declaration supplies help, validation, doctor, and targeted discovery, so help shows a command's required arguments, flag groups, and effects from the same fields that `describe` returns, as [Usage](../north-star/contract.md#usage) says. Build no custom schema language or framework for future commands.

Describe comes before [M04](M04-skill.md), so that `skill` finds the skills of installed extensions through `ncly describe --json`, as an agent would, and the `nicely` skill teaches discovery from its first release.

The scenarios reach a bundled extension through `ncly-bundled`, the test build of M02 whose bundled list holds the fixture `hello`, and an external one through the shell fixture of M01. Calls to describe and help, and calls with invalid usage, start neither one, as an execution marker that never appears proves. `describe_schema.txtar` starts the fixture in separate calls to collect its real answers, and `usage_declarations.txtar` starts it for its valid calls.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M03-T1 | Concise listing | agent | — | todo |
| M03-T2 | Command detail | agent | M03-T1 | todo |
| M03-T3 | Result schemas | agent | M03-T2 | todo |
| M03-T4 | Config sources | agent | M03-T2 | todo |
| M03-T5 | Help from declarations | agent | M03-T2 | todo |

### M03-T1 Concise listing

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Extensions](../north-star/contract.md#extensions), [Output](../north-star/contract.md#output), [Compatibility](../north-star/contract.md#compatibility)
- **Proves:** `testdata/script/describe_list.txtar`, `sdk/contract/compat_test.go`

`ncly describe` without a path, and with a group path, returns concise entries with `path`, `summary`, and `source`, in the order of their paths. An external extension's entries come from its manifest, and the scenario proves that it never starts. The entry of an extension's domain adds `skill`, the folder `ncly-<domain>.skill/` beside the manifest when it exists, and `null` otherwise or for a bundled extension. A broken extension's commands are absent, as from `ncly --help`, and a path inside its domain answers `EXTENSION_INVALID`, as `ncly <domain>` does.

An unknown path answers `NOT_FOUND`, whose `next` is `["ncly","describe"]`. As the first problem with `next`, it brings the key to the SDK's problem and its printer, and a consumer case in `compat_test.go` reads an answer whose problem has `next` and one without it.

### M03-T2 Command detail

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Command descriptions](../north-star/contract.md#command-descriptions), [Manifest](../north-star/contract.md#manifest), [Dispatch](../north-star/contract.md#dispatch), [SDK](../north-star/contract.md#sdk)
- **Proves:** `testdata/script/describe_command.txtar`, `testdata/script/usage_declarations.txtar`

A command path returns `args`, `flags`, `flag_groups`, `effects`, `requires`, `modes`, and `examples`, where `flags` holds the declared flags, plus `--dry-run` and `--run-id` when `modes` sets `dry_run` or `recorded`. Extend the M00 declarations, and add each field to [Manifest](../north-star/contract.md#manifest) in the same pull request, where they lack one: requiredness, value types, defaults, variadic flags, flag groups, effects, modes, and the programs and keys that a command requires. A fixture command that declares the `recorded` mode lists `run-id` among its flags.

Validation uses the same fields. A missing required argument or flag, a value that its type rejects, and a call that breaks a flag group each answer `USAGE_INVALID` before the command runs, so no fixture's execution marker appears: through `ncly` for the core commands that declare them and for the external shell fixture, and through the SDK's entry point, `ncly-hello` run directly, for the Go fixture. A fixture flag that the manifest declares variadic receives two distinct values, one of them holding a comma, and keeps them in the caller's order through core's parsing, dispatch, and the SDK's parsing: the shell fixture receives one `--<name>=<value>` word per value, and the Go fixture's handler reads them from `Lists`, through `ncly` and run directly. A fixture command whose `modes` sets `dry_run` and `recorded` receives `--dry-run=true` and `--run-id=<id>` after its own flags, the Go fixture's handler reads both from `Flags`, and a manifest that declares either as an own flag is invalid.

### M03-T3 Result schemas

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `testdata/script/describe_schema.txtar`

`result` gives `kind`, `keys`, and a JSON Schema, draft 2020-12, with every reference resolved and no remote fetch, plus `dry_run_schema` where dry run is supported. Derive `keys` from the schema. The scenario checks the fixture's real answers against the schema that its invocation selects, the normal one or the dry-run one, so a schema cannot drift from what the command answers.

### M03-T4 Config sources

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Configuration](../north-star/contract.md#configuration)
- **Proves:** `testdata/script/describe_config.txtar`, `sdk/config/config_test.go`

`config` lists each key a command reads, with its value and source, and never a secret. An extension's command names the keys that it reads in its manifest. `sdk/config` keeps the source of each value as it merges the files: the flag, the variable, the file that set it, or the default. A table merges key by key, so each of its keys keeps its own source, and a list keeps the source of the file that replaced it, even when that list is empty. With an invalid config file, discovery uses defaults and adds `CONFIG_DEFAULTS_USED`.

### M03-T5 Help from declarations

- **Read:** [Usage](../north-star/contract.md#usage), [Command descriptions](../north-star/contract.md#command-descriptions)
- **Proves:** `testdata/script/help.txtar`

The help of the fixture commands of M03-T2 shows a required argument as `<name>`, an optional one as `[name]`, a flag group as `(--a | --b)`, and one line for a command that writes, uses the network, or may bill, with `--dry-run` named when it is supported, in the help style of M00. The scenario compares each page with the answer of `ncly describe` for the same command, so help and discovery never disagree, and runs once under the pseudo-locale.
