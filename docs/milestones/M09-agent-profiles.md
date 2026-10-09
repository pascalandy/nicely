# M09 Agent profiles

Status: planned
Version: v0.9.0

## Demo

`ncly agent run --dry-run` shows the selected profile and the redacted claude or pi invocation, without starting a harness.

## Scope

`agent` is the first real external extension: Nicely needs no agent to operate, so it ships outside the binary. It lives in `extensions/agent/`, builds as `ncly-agent`, and owns profiles, harness adapters, tools-off mode, and the depth limit, as [extensions/agent/spec.md](../../extensions/agent/spec.md) defines. This milestone reads profiles and prepares runs. Execution and its records arrive in [M10](M10-agent-run.md), and the other harnesses in [M11](M11-harnesses.md).

Before taps exist, `just install` builds the first-party extensions and puts them where core finds them.

## Open questions

Settle each one in [extensions/agent/spec.md](../../extensions/agent/spec.md), then delete this section.

- Whether `ncly agent run` reads stdin without `-` when stdin is a pipe, given that some harnesses leave stdin open
- The answer keys of `ncly agent run` that transcript records: `run_id`, `answer`, `profile`, `harness`, the harness version, `model`, and `duration_ms`
- How a user without the repository installs a first-party external extension before taps arrive in M12

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M09-T1 | Extension and install | agent | — | todo |
| M09-T2 | Profiles | agent | M09-T1 | todo |
| M09-T3 | Claude adapter dry run | agent | M09-T2 | todo |
| M09-T4 | Pi adapter dry run | agent | M09-T3 | todo |
| M09-T5 | Depth limit | agent | M09-T3 | todo |

### M09-T1 Extension and install

- **Read:** [extensions/agent/spec.md](../../extensions/agent/spec.md), [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/agent_install.txtar`

Create `extensions/agent/` with its manifest, spec, and `SKILL.md`, built as `ncly-agent` on `sdk/`. `just install` builds it into the data folder, and `ncly agent --help` reaches it through core.

### M09-T2 Profiles

- **Read:** [Profiles](../../extensions/agent/spec.md#profiles)
- **Proves:** `testdata/script/agent_profile.txtar`

`[agent.profiles]` and `default_profile` in the config, with `ncly agent profile list` and `ncly agent profile view <name>`. `--harness`, `--model`, `--effort`, and `--provider` describe a profile without reading the config. An unknown profile exits 2 with `NOT_FOUND`. A profile's semantic constraint fails with `CONFIG_INVALID` only when a command selects that profile.

### M09-T3 Claude adapter dry run

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters), [ncly agent](../../extensions/agent/spec.md#ncly-agent)
- **Proves:** `testdata/script/agent_dry_run.txtar`

The `claude` adapter checks the harness version, 2.1.291 or later, and `--dry-run` shows the profile, the requested mode, and the redacted invocation with its tools-off flags. A stub harness in `PATH` proves that the harness never starts. A harness older than its minimum is `PREREQ_MISSING`. A profile for a harness without an adapter fails with `CAPABILITY_UNSUPPORTED`.

### M09-T4 Pi adapter dry run

- **Read:** [Adapters](../../extensions/agent/spec.md#adapters)
- **Proves:** `testdata/script/agent_dry_run_pi.txtar`

The `pi` adapter, version 1.0.4 or later, with its provider, its efforts, and its tools-off flags, under the same checks as `claude`.

### M09-T5 Depth limit

- **Read:** [Depth limit](../../extensions/agent/spec.md#depth-limit)
- **Proves:** `testdata/script/agent_depth.txtar`

At `[agent] max_depth`, 2 by default, `ncly agent run` exits 2 with `AGENT_DEPTH_LIMIT` before preparing a harness. A launched harness receives `NCLY_AGENT_DEPTH` plus one.
