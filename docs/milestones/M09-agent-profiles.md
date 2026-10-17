# M09 Codex preparation

Status: ready
Version: v0.9.0

## Demo

`ncly agent run "Review this branch" --tools --dry-run` checks Codex and shows its invocation without starting a task.

## Scope

`agent` is an external extension in `extensions/agent/`, built as `ncly-agent` on the existing SDK. It prepares one trusted task for Codex. [Its spec](../../extensions/agent/spec.md) defines the command, settings, and tools policy. Execution and records arrive in [M10](M10-agent-run.md).

Use Codex's existing login. The extension declares no API key and forwards none. [M08](M08-grants.md) still serves extensions that need a key, such as transcript. A Codex API key and additional harnesses stay in [M99](M99-parking-lot.md#agent-extras).

One model and effort can come from flags or `[agent]` defaults. Named profiles, providers, a model catalog, and a generic adapter registry stay in M99. There is one Codex invocation to implement.

`just install` puts `ncly` in `~/.local/bin/` and the extension files in Nicely's data folder. Taps wait in [M99](M99-parking-lot.md#taps).

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M09-T1 | Extension and install | agent | — | todo |
| M09-T2 | Codex dry run | agent | M09-T1 | todo |
| M09-T3 | Depth limit | agent | M09-T2 | todo |

### M09-T1 Extension and install

- **Read:** [agent spec](../../extensions/agent/spec.md), [Extensions](../north-star/contract.md#extensions), [Required programs](../north-star/core-spec.md#required-programs)
- **Proves:** `extensions/agent/testdata/script/agent_install.txtar`

Add the manifest, catalogs, skill, and binary in `extensions/agent/`. Add `just install`, which puts `ncly` in `~/.local/bin/` and the extension files in Nicely's data folder. Declare Codex as a required program, with its minimum version and local version probe. Existing doctor behavior then checks `agent.codex`; no separate profile scanner or status command is needed.

Prove that `ncly agent --help` reaches the installed extension. A missing or old Codex fails `ncly doctor agent` with `PREREQ_MISSING` and its install hint. A counted stub proves that doctor starts no task.

### M09-T2 Codex dry run

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent), [Settings](../../extensions/agent/spec.md#settings), [Codex invocation](../../extensions/agent/spec.md#codex-invocation), [Operations](../north-star/contract.md#operations)
- **Proves:** `extensions/agent/testdata/script/agent_dry_run.txtar`

Resolve the model, effort, trusted task, and working directory. Check Codex's version and the requested tools mode through `sdk/runner`. Show the redacted invocation without launching a task. Missing model, unreadable task file, and invalid settings fail before launch.

Require explicit `--tools`. Without it, including in a dry run, answer `CAPABILITY_UNSUPPORTED`. The MVP accepts no `--input`, harness selector, or provider selector. Preserve the filtering of API keys. A stub verifies stdin delivery preparation, the sandbox flags, and that dry run never starts a task or reads stdin.

### M09-T3 Depth limit

- **Read:** [Depth limit](../../extensions/agent/spec.md#depth-limit)
- **Proves:** `extensions/agent/testdata/script/agent_depth.txtar`

At `[agent] max_depth`, 2 by default, preparation fails with `AGENT_DEPTH_LIMIT`. A launched Codex receives `NCLY_AGENT_DEPTH` plus one. Check this bound before preparing a child invocation.
