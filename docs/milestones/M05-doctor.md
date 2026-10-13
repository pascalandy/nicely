# M05 Doctor

Status: planned
Version: v0.5.0

## Demo

`ncly doctor` reports core's readiness and the requirements of every installed extension, on stdout even when a check fails.

## Scope

Implement [ncly doctor](../north-star/core-spec.md#ncly-doctor) with the `core`, `extensions`, and `skill` components. The `auth` component arrives with keys in [M06](M06-auth.md), and the `transcript` component in [M16](M16-transcript-finish.md). Doctor adds bounded local `--version` probes to the program runner that M01 brought. It runs them one at a time, because the runner's cleanup signals every descendant of `ncly`, so a probe's timeout would also stop a probe that runs beside it. Probes run together only once [#49](https://github.com/pascalandy/nicely/issues/49) scopes that cleanup to one program.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M05-T1 | Report and core component | agent | — | todo |
| M05-T2 | Bounded probes | agent | M05-T1 | todo |
| M05-T3 | Extensions component | agent | M05-T2 | todo |
| M05-T4 | Skill component | agent | M05-T1 | todo |

### M05-T1 Report and core component

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Output](../north-star/contract.md#output), [Configuration](../north-star/contract.md#configuration)
- **Proves:** `testdata/script/doctor_core.txtar`

Add the path that keeps a failed report on stdout, because an ordinary failed answer goes to stderr. Add `doctor` to the `nicely` skill: an agent calls it during setup, or when a dry run reports a problem that the agent cannot explain. Invalid config is the failed `core.config` finding with `CONFIG_INVALID`, independent checks still run, and an unknown key is a `CONFIG_UNKNOWN_KEY` warning. Doctor exits 0 without a failed check, 78 with one, and 1 when doctor itself fails.

### M05-T2 Bounded probes

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs)
- **Proves:** `testdata/script/doctor_probe.txtar`

Add a bounded local `--version` probe of each required program to the program runner, run one at a time. `--timeout` bounds the checks, 30 seconds by default. A missing or too old program fails with `PREREQ_MISSING`, whose `next` holds the install command when one is known for this system, and a probe that hangs ends at the timeout, after which the remaining probes report that they did not run.

### M05-T3 Extensions component

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `testdata/script/doctor_extensions.txtar`

The `extensions` component lists every extension that core finds and where it came from, every extension that a core command, a bundled extension, or an earlier extension hides, and every incompatible manifest. It starts no extension.

### M05-T4 Skill component

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `testdata/script/doctor_skill.txtar`

The `skill` component warns with `SKILL_SOURCE_MISSING`, `SKILL_NAME_MISMATCH`, and `SKILL_INVALID`. A skill subfolder of `[skill] paths` whose name differs from its skill's `name` is a warning, and the report stays `ok`. A correctly named skill in a folder named `skills`, and an installed extension's skill in `ncly-<domain>.skill/`, named after its domain, get no `SKILL_NAME_MISMATCH`.
