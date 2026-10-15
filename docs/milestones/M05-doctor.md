# M05 Doctor

Status: done
Version: v0.5.0

## Demo

`ncly doctor` reports core's config, every installed extension with the programs that it requires, and the checks that `skill` adds through `ncly skill status`, on stdout even when a check fails.

## Scope

Implement [ncly doctor](../north-star/core-spec.md#ncly-doctor) with the `core` and `extensions` components, a component for each installed extension, and the [status command](../north-star/contract.md#status-command) through which an extension adds its own checks, which `ncly skill status` proves. The `auth` component arrives with keys in [M06](M06-auth.md), and transcript's status command in [M16](M16-transcript-finish.md). Doctor never installs anything, as [D011](../north-star/decisions/D011-local-checks.md) decides, so `ncly auth` in M06 stays the first command that writes.

Doctor runs the version probes and the status commands through the program runner that M01 brought, and kills one at the limit with the stop that [M04-T5](M04-skill.md#m04-t5-skill-name-completion) adds for completion. It runs them one at a time, because the runner's cleanup signals every descendant of `ncly`, so the timeout of one would also stop another that runs beside it. They run together only once [#49](https://github.com/pascalandy/nicely/issues/49) scopes that cleanup to one program.

The scenarios use the shell fixture of M01, which gains required programs, config keys, and a status command, and stub programs in `PATH` that stand in for real tools and package managers.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M05-T1 | Report and core component | agent | — | done |
| M05-T2 | Required programs | agent | M05-T3 | done |
| M05-T3 | Extensions component | agent | M05-T1 | done |
| M05-T4 | Status commands and skill status | agent | M05-T2 | done |

### M05-T1 Report and core component

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Output](../north-star/contract.md#output), [Configuration](../north-star/contract.md#configuration), [SDK](../north-star/contract.md#sdk)
- **Proves:** `testdata/script/doctor_core.txtar`, `testdata/script/doctor_config.txtar`, `extensions/skill/testdata/script/skill_nicely.txtar`

Add `doctor` to core in `internal/cli`, with the `core` component, the check object, and the `component` argument, which answers `NOT_FOUND` with exit 2 for an unknown component. The report goes through the report outcome of the SDK, which keeps a failed report on stdout, as [SDK](../north-star/contract.md#sdk) says of core's doctor. Invalid config is the failed `core.config` finding with `CONFIG_INVALID`, and independent checks still run. Doctor exits 0 without a failed check and 78 with one, and `doctor_core.txtar` proves a failed report: exit 78, the findings under `checks` on stdout, a failed check with its `hint`, and no `errors`.

Add `doctor` to the `nicely` skill in `extensions/skill`: an agent calls it during setup, or when a dry run reports a problem that the agent cannot explain. Update the skill's rule for reading a failure: doctor and each status command marked `doctor` put their findings under `checks`, and on exit 78 the agent relays the `hint` and `next` of each failed check to the human. Another report command, such as `ncly auth status`, puts its findings under the data key that `ncly describe` gives it, and `errors` stays the path for a failure of the command itself, so the skill never teaches `checks` as a key of every report. The skill stays within its 60 lines, as `skill_nicely.txtar` checks.

In `doctor_config.txtar`, each unknown key is a `warn` entry of `core.config` with `CONFIG_UNKNOWN_KEY`. A key that the shell fixture's manifest declares under `[hello]` gets none, an undeclared key of `[hello]` gets one, and a table named after a domain that no installed extension has gets none, in the shared and in the local config. Doctor reads the declared keys from the manifests that the discovery of `internal/cli` finds, as [M03-T4](M03-describe.md#m03-t4-config-sources) declares them, and the unknown keys from `sdk/config`, which collects them since M00.

### M05-T2 Required programs

- **Read:** [Required programs](../north-star/core-spec.md#required-programs), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Command tree](../north-star/core-spec.md#command-tree), [Manifest](../north-star/contract.md#manifest), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [Compatibility](../north-star/contract.md#compatibility)
- **Proves:** `testdata/script/doctor_probe.txtar`, `testdata/script/extension_manifest.txtar`, `sdk/contract/compat_test.go`

Add `conditional`, `version`, `probe`, and `install` to the entries of `programs` in the manifest of `sdk/contract`, with their rules, and to the entries of `requires.bins` that `ncly describe` returns, with a consumer case in `compat_test.go` that reads an entry with and without them. Add the program check to `sdk/runner`, public so that a Go extension's status command calls it too: it looks a program up, probes it with stdin closed when it declares `version`, reads and compares the version, and picks the install command. Each installed extension becomes a component of doctor in `internal/cli`, which runs that check for each program that is not conditional, one at a time, and stops a probe at the limit with the stop of M04-T5. Register `CHECK_TIMEOUT` in the registry of `sdk/contract`. Because each installed extension now names a component, add `core` and `extensions` to `reservedNames` in `internal/cli`, and to the extra names that `TestReservedNamesAgreeWithTheDomainTable` in `internal/cli/spec_test.go` expects beside the domain table's rows, next to `help`, `version`, and `config`.

`doctor_probe.txtar` puts stub programs in `PATH`: one missing, one too old, one recent enough, one that prints its version on stderr, one that exits 1, and one that hangs. A conditional program that is absent, and one that is too old, get no check. A failed check holds, in its hint and its `next`, the install command of the first package manager that the stub `PATH` holds, in the order of the running system, and a program without a command for that manager has no `next`. With a short `--timeout`, the hanging probe's check fails with `CHECK_TIMEOUT`, and so does each check that had not started. `extension_manifest.txtar` proves that the manifest rejects a malformed `version`, an empty `probe`, and an empty install command. In `doctor_probe.txtar`, an `ncly-extensions` in `PATH` is skipped: `ncly doctor extensions` answers a valid report of the built-in `extensions` component of M05-T3, and an execution marker that never appears proves that the impostor never starts.

### M05-T3 Extensions component

- **Read:** [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Extensions](../north-star/contract.md#extensions), [Discovery](../north-star/contract.md#discovery), [Compatibility](../north-star/contract.md#compatibility)
- **Proves:** `testdata/script/doctor_extensions.txtar`, `sdk/contract/compat_test.go`

The `extensions` component in `internal/cli` holds one `extensions.<domain>` entry for each extension that core finds, with `domain`, `source`, and `path`: a `pass` for the one that runs, a `pass` with `hidden_by` for each one that a core command, a bundled extension, or an earlier extension hides, and a `fail` with `EXTENSION_INVALID` for an incompatible manifest. It starts no extension, as an execution marker that never appears proves. `doctor_extensions.txtar` asserts the keys of each entry, including the `null` path of the bundled `skill`, and a consumer case in `compat_test.go` reads a check with and without the keys of an extension.

### M05-T4 Status commands and skill status

- **Read:** [Status command](../north-star/contract.md#status-command), [ncly skill status](../../extensions/skill/spec.md#ncly-skill-status), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Manifest](../north-star/contract.md#manifest), [Dispatch](../north-star/contract.md#dispatch), [Answer check](../north-star/contract.md#answer-check), [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill)
- **Proves:** `extensions/skill/testdata/script/skill_status.txtar`, `testdata/script/doctor_status.txtar`, `testdata/script/extension_manifest.txtar`

Add `doctor` to the manifest of `sdk/contract`, with the conditions of a status command, which `Check` reports for a Go extension too. Doctor in `internal/cli` dispatches each status command through the extension that discovery found, one at a time after the probes, checks the whole answer, and lists its checks under the extension's component.

The first consumer is `ncly skill status` in `extensions/skill`, which reaches core only through its environment and the CLI: it reads `[skill] paths` through `sdk/config`, and the skills of installed extensions through `ncly describe --json`, as [M04-T3](M04-skill.md#m04-t3-skills-from-extensions) does. Its declaration sets `doctor = true`. Add it to the `nicely` skill, within its 60 lines. `skill_status.txtar` runs `ncly skill status` alone and `ncly doctor skill`, and checks that both report the same checks, one entry per problem. `skill.sources` warns with `SKILL_SOURCE_MISSING`, `skill.names` with `SKILL_NAME_MISMATCH` for a subfolder of `[skill] paths` whose name differs from its skill's `name` and for an extension's skill that is not named after its domain, and `skill.files` with `SKILL_INVALID`, and the report stays `ok`. A correctly named skill in a folder named `skills`, and an installed extension's skill in `ncly-<domain>.skill/`, named after its domain, get no `SKILL_NAME_MISMATCH`.

`doctor_status.txtar` covers the edge cases with a `hello status` in the shell fixture, which sets `doctor = true`, declares `live` and `timeout`, and records the canonical arguments that it receives, so the scenario proves `--live=true` and `--timeout=<n>s`. A variable steers its answer: a warning, a failed check with `next`, a valid `pass` beside an entry whose `component` is another domain, a failure with `CONFIG_INVALID`, a failure with `RUNTIME`, and an answer that hangs. Doctor lists the two valid answers as they are. In place of each other answer, it lists one `hello.status` check with `EXTENSION_INVALID`, `CONFIG_INVALID`, or `CHECK_TIMEOUT`, and never the `pass` of the mixed answer. With a broken manifest, `ncly doctor hello` answers `EXTENSION_INVALID` with exit 78 and starts nothing. `extension_manifest.txtar` proves that the manifest rejects `doctor` on a path other than `hello status`, on a group, on a command whose kind is not `report`, with an argument or another own flag, with a required own flag, a `live` whose default is true, a `timeout` that is not a `string`, or a flag group, with effects that write or bill or without effects, with `network` effects and no `live`, and with a mode.
