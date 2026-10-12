# M03 Describe

Status: planned
Version: v0.3.0

## Demo

`ncly describe` describes core commands, a bundled extension, and an external extension in the same shape, without running any of them.

## Scope

Implement [ncly describe](../north-star/core-spec.md#ncly-describe). Grow the declarations of M00 with the fields that the commands of M01 and M02 need. One declaration supplies help, validation, doctor, and targeted discovery. Build no custom schema language or framework for future commands.

Describe comes before [M04](M04-skill.md), so that `skill` finds the skills of installed extensions through `ncly describe --json`, as an agent would, and the `nicely` skill teaches discovery from its first release.

## Open questions

Settle each one in [ncly describe](../north-star/core-spec.md#ncly-describe), then delete this section.

- What `since` means for an extension's command: the release of `ncly` or the version of the extension
- The key that gives an external extension's location, beside `source`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M03-T1 | Concise listing | agent | — | todo |
| M03-T2 | Command detail | agent | M03-T1 | todo |
| M03-T3 | Result schemas | agent | M03-T2 | todo |
| M03-T4 | Config sources | agent | M03-T2 | todo |

### M03-T1 Concise listing

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `testdata/script/describe_list.txtar`

`ncly describe` without a path, and with a group path, returns concise entries with `path`, `summary`, `since`, and `source`. An external extension's entries come from its manifest, and the scenario proves that it never starts. An unknown path answers `NOT_FOUND`.

### M03-T2 Command detail

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Command descriptions](../north-star/contract.md#command-descriptions)
- **Proves:** `testdata/script/describe_command.txtar`

A command path returns `args`, `flags` with the applicable write, output, and timeout flags, `flag_groups`, `effects`, `requires`, `modes`, and `examples`. Extend the M00 declarations where they lack a field, with requiredness, value types, and defaults.

### M03-T3 Result schemas

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `testdata/script/describe_schema.txtar`

`result` gives `kind`, `keys`, and a JSON Schema, draft 2020-12, with every reference resolved and no remote fetch, plus `dry_run_schema` where dry run is supported. Derive `keys` from the schema.

### M03-T4 Config sources

- **Read:** [ncly describe](../north-star/core-spec.md#ncly-describe), [Configuration](../north-star/contract.md#configuration)
- **Proves:** `testdata/script/describe_config.txtar`

`config` lists each key a command reads, with its value and source, and never a secret. With an invalid config file, discovery uses defaults and adds `CONFIG_DEFAULTS_USED`.
