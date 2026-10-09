# M05 Doctor

Status: planned
Version: v0.5.0

## Demo

`ncly doctor` reports core's readiness and the requirements of every installed extension, on stdout even when a check fails.

## Scope

Implement [ncly doctor](../north-star/cli-spec.md#ncly-doctor) with the `core`, `extensions`, and `skill` components. The `auth` component arrives with keys in [M06](M06-auth.md), and the `transcript` component in [M14](M14-transcript-finish.md). Doctor adds bounded local `--version` probes to the program runner that M01 brought.

## Open questions

Settle each one in [ncly doctor](../north-star/cli-spec.md#ncly-doctor) and [Extensions](../north-star/cli-spec.md#extensions), then delete this section.

- How an extension adds checks beyond its manifest's `requires`, such as the Python version or the browser cookies of transcript: a declared doctor command that core runs, or checks that the manifest declares
- How doctor tells an unknown config key from a key in an extension's table, so `CONFIG_UNKNOWN_KEY` never fires on a key that an installed extension owns
- Where the install command of each prerequisite lives for each system, such as in the manifest's `requires`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M05-T1 | Report and core component | agent | — | todo |
| M05-T2 | Bounded probes | agent | M05-T1 | todo |
| M05-T3 | Extensions component | agent | M05-T2 | todo |
| M05-T4 | Skill component | agent | M05-T1 | todo |
| M05-T5 | Install offers | agent | M05-T2 | todo |

### M05-T1 Report and core component

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor), [Output](../north-star/cli-spec.md#output), [Configuration](../north-star/cli-spec.md#configuration)
- **Proves:** `testdata/script/doctor_core.txtar`

Add the path that keeps a failed report on stdout, because an ordinary failed answer goes to stderr. Invalid config is the failed `core.config` finding with `CONFIG_INVALID`, independent checks still run, and an unknown key is a `CONFIG_UNKNOWN_KEY` warning. Doctor exits 0 without a failed check, 78 with one, and 1 when doctor itself fails.

### M05-T2 Bounded probes

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor), [Programs that ncly runs](../north-star/cli-spec.md#programs-that-ncly-runs)
- **Proves:** `testdata/script/doctor_probe.txtar`

Add a bounded local `--version` probe of each required program to the program runner. `--timeout` bounds the checks, 30 seconds by default. A missing or too old program fails with `PREREQ_MISSING`, and a probe that hangs ends at the timeout.

### M05-T3 Extensions component

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor), [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/doctor_extensions.txtar`

The `extensions` component lists every extension that core finds and where it came from, every extension that a core command, a bundled extension, or an earlier extension hides, every incompatible manifest, and the programs that each manifest requires. It starts no extension.

### M05-T4 Skill component

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `testdata/script/doctor_skill.txtar`

The `skill` component warns with `SKILL_SOURCE_MISSING`, `SKILL_NAME_MISMATCH`, and `SKILL_INVALID`. A folder whose name differs from its skill's `name` is a warning, and the report stays `ok`.

### M05-T5 Install offers

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor), [Modes](../north-star/cli-spec.md#modes)
- **Proves:** `testdata/script/doctor_install.txtar`

In interactive mode, doctor shows a missing tool's install command and runs it only after a yes. In non-interactive mode, it never installs anything, and the hint of the failed check holds the command.
