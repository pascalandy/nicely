# M11 Harnesses

Status: planned
Version: v0.11.0

## Demo

`ncly agent run` works through Codex, Grok, and OpenCode too, and `--check-unchanged` reports the files that a run changed.

## Scope

Extend the `claude` and `pi` adapters of M09 with `codex`, `grok`, and `opencode`. Use the `headless` skill's existing adapters and notes where available. Each adapter declares supported versions, tools-off behavior, input delivery, cancellation, and result parsing, and its declared modes are validated on the real installed harness before they are enabled.

`--check-unchanged` reports the files that a run changed, with the limits that [ncly agent](../../extensions/agent/spec.md#ncly-agent) states. Write-prevention sandboxes and cost tracking stay out of scope.

`ncly agent status`, the status command of the agent domain, adds checks to the `agent` component of `ncly doctor`: that the harness of each profile is installed, at or above its minimum version. It runs only the harness's declared local version probe, and never starts a harness task, because a run can bill.

## Open questions

Settle each one in [extensions/agent/spec.md](../../extensions/agent/spec.md), then delete this section.

- The verified tools-off flags and supported versions of each new harness
- The key variables that each harness reads, which its adapter declares as optional keys
- Whether `ncly agent model list` replaces the `list models` command of the transcript CLI

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M11-T1 | Codex adapter | agent | — | todo |
| M11-T2 | Grok adapter | agent | — | todo |
| M11-T3 | OpenCode adapter | agent | — | todo |
| M11-T4 | Check unchanged | agent | — | todo |
| M11-T5 | Doctor agent | agent | — | todo |

### M11-T1 Codex adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `extensions/agent/testdata/script/agent_codex.txtar`

The `codex` adapter, with its verified tools-off mode.

### M11-T2 Grok adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `extensions/agent/testdata/script/agent_grok.txtar`

The `grok` adapter, with its verified tools-off mode.

### M11-T3 OpenCode adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `extensions/agent/testdata/script/agent_opencode.txtar`

The `opencode` adapter, with its provider and its verified tools-off mode.

### M11-T4 Check unchanged

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent)
- **Proves:** `extensions/agent/testdata/script/agent_check_unchanged.txtar`

`--check-unchanged` exits 1 with `AGENT_FILES_CHANGED` when a covered file changed during the run, and reports nothing about ignored files. A completed run's ID with the flag toggled answers `USAGE_INVALID`, and the counted stub harness starts no second time.

### M11-T5 Doctor agent

- **Read:** [Status command](../north-star/contract.md#status-command), [Required programs](../north-star/core-spec.md#required-programs), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `extensions/agent/testdata/script/agent_status.txtar`

`ncly agent status` checks the harness of each profile through the program check of `sdk/runner`, which runs only its declared local version probe, never a harness task, and fails a missing or too old one with `PREREQ_MISSING` and its install command. `ncly doctor agent` lists the same checks. In `agent_status.txtar`, the harness of the configured profile is present, another supported harness is absent, and `ncly doctor agent` stays `ok`. A selected harness below its minimum version then fails with `PREREQ_MISSING`, the unused absent one still fails nothing, and the counted stub harnesses prove that no task starts.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal runs a real task with each of the five harnesses he has installed
- The `headless` skill of `pascalandy/skills` points to `ncly agent`
