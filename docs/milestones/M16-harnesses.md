# M16 Harnesses

Status: planned
Version: v0.16.0

## Demo

`ncly agent run` works through Codex, Grok, and OpenCode too, and `--check-unchanged` reports the files that a run changed.

## Scope

Extend the `claude` and `pi` adapters of M09 with `codex`, `grok`, and `opencode`. Use the `headless` skill's existing adapters and notes where available. Each adapter declares supported versions, tools-off behavior, input delivery, cancellation, and result parsing, and its declared modes are validated on the real installed harness before they are enabled.

`--check-unchanged` reports the files that a run changed, with the limits that [ncly agent](../../extensions/agent/spec.md#ncly-agent) states. Write-prevention sandboxes and cost tracking stay out of scope.

A new `agent` component of `ncly doctor` checks that the harness of each profile is installed. It never starts a harness, because a run can bill.

## Open questions

Settle each one in [extensions/agent/spec.md](../../extensions/agent/spec.md), then delete this section.

- The verified tools-off flags and supported versions of each new harness
- The key variables that each harness reads, which its adapter declares as optional keys
- Whether `ncly agent model list` replaces the `list models` command of the transcript CLI

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M16-T1 | Codex adapter | agent | — | todo |
| M16-T2 | Grok adapter | agent | — | todo |
| M16-T3 | OpenCode adapter | agent | — | todo |
| M16-T4 | Check unchanged | agent | — | todo |
| M16-T5 | Doctor agent | agent | — | todo |

### M16-T1 Codex adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `testdata/script/agent_codex.txtar`

The `codex` adapter, with its verified tools-off mode.

### M16-T2 Grok adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `testdata/script/agent_grok.txtar`

The `grok` adapter, with its verified tools-off mode.

### M16-T3 OpenCode adapter

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `testdata/script/agent_opencode.txtar`

The `opencode` adapter, with its provider and its verified tools-off mode.

### M16-T4 Check unchanged

- **Read:** [ncly agent](../../extensions/agent/spec.md#ncly-agent)
- **Proves:** `testdata/script/agent_check_unchanged.txtar`

`--check-unchanged` exits 1 with `AGENT_FILES_CHANGED` when a covered file changed during the run, and reports nothing about ignored files.

### M16-T5 Doctor agent

- **Read:** [ncly doctor](../north-star/cli-spec.md#ncly-doctor)
- **Proves:** `testdata/script/doctor_agent.txtar`

The `agent` component checks that the harness of each profile is installed, without starting one.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal runs a real task with each of the five harnesses he has installed
- The `headless` skill of `pascalandy/skills` points to `ncly agent`
