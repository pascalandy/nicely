# Agent extension spec

`agent` is an external extension, built as `ncly-agent` from this folder. It runs one trusted task through Codex. [M09](../../docs/milestones/M09-agent-profiles.md) adds preparation and [M10](../../docs/milestones/M10-agent-run.md) adds execution and inspection. It follows the [contract](../../docs/north-star/contract.md).

[Deferred agent features](../../docs/milestones/M99-parking-lot.md#agent-extras) include other harnesses, named profiles, providers, model discovery, a Codex API key, and post-run file comparison. They are not prerequisites of this extension.

## ncly agent

```
ncly agent run (<task> | --task-file <file>) --tools [--model <id>] [--effort <level>] [--cwd <dir>] [--run-id <id>] [--timeout <duration>] [--dry-run] [--json]
```

- Exactly one task argument or `--task-file` is required. The file contains the trusted task instructions. An unreadable file fails with `USAGE_INVALID` before launch.
- `--cwd` selects an existing directory, or defaults to the current directory. An invalid directory fails with `USAGE_INVALID` before launch.
- The adapter requires explicit `--tools`. Without it, even in a dry run, the command answers `CAPABILITY_UNSUPPORTED` before launching Codex. Never enable tools silently.
- External material is outside this first command. It accepts no `--input` flag. Do not put a transcript or web page into `--task-file` to bypass [D028](../../docs/north-star/decisions/D028-outside-text.md).
- Nicely reads no stdin for this command. It sends the prepared task to Codex on stdin and closes that pipe after the task.
- Dry run shares preparation with execution. It checks settings, depth, directory, Codex version, and the tools mode. It shows the resolved settings and redacted invocation, starts no task, and records nothing. Login validity remains a pending execution check; a dry run makes no paid request.
- The default timeout is 5 minutes. `--timeout` overrides it under the common duration rules.
- Text mode prints the answer. JSON uses the common envelope and returns `run_id`, `answer`, `harness`, `model`, and `duration_ms`. `harness` is `codex`. The record's `tools` holds the harness version.
- Preparation errors use `USAGE_INVALID`, `CONFIG_INVALID`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, or `AGENT_DEPTH_LIMIT`. Failures after launch exit 1. Signals retain exits 130 and 143. No failure after launch permits an automatic new task.

```bash
ncly agent run "Review this branch for risky changes" --tools --model '<model id>' --dry-run
ncly agent run --task-file task.md --tools --run-id "$RUN_ID" --json
ncly run view "$RUN_ID" --json
```

## Settings

The model and effort come from the explicit flags, then `[agent]` in the normal configuration precedence. No model ID is built into Nicely. A missing or empty model is `USAGE_INVALID`, with a hint naming `--model`. Effort defaults to `medium`. Invalid configured values use `CONFIG_INVALID`; invalid explicit flags use `USAGE_INVALID`.

```toml
[agent]
model = "<model id>"
effort = "medium"
max_depth = 2
```

The manifest declares `agent.model`, `agent.effort`, and `agent.max_depth`. `max_depth` is a positive integer. There is no profile registry, harness selector, or provider selector. Future settings require their own demonstrated use.

## Codex invocation

Codex version 0.162.1 or later runs `exec` with `--model` and `-c model_reasoning_effort="<effort>"`. It also passes `--sandbox workspace-write`, `--skip-git-repo-check`, `--cd`, `--json`, and `--output-last-message <file>`. The final `-` sends the task through stdin. Keep Codex's sandbox enabled. No Nicely flag promises that Codex cannot write.

Stdout contains JSONL events. The answer comes from the output file and must contain text after a zero exit. A nonzero exit or missing answer after launch leaves the step unknown. Supported effort values are `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`, subject to the model. Preparation checks the known effort values without a model-catalog API; unsupported model choices fail at execution.

Codex has no verified switch that removes all tools, so this adapter supports only explicit `--tools`. The flags above retain the help check recorded on 2026-10-16 in [PR #156](https://github.com/pascalandy/nicely/pull/156). Real task compatibility still needs the independent installed-harness check in M10. The adapter uses `sdk/runner` for execution, cancellation, and environment filtering.

Declare Codex as a required program with its local version probe. Existing `ncly doctor agent` checks that declaration. A separate `agent status` command and a matrix of configured harnesses would duplicate that first check.

Codex uses its own existing login. `agent` declares no API key and receives none from Nicely. Keep the runner's removal of `*_API_KEY` variables. A user configures the native login outside Nicely; the extension never starts an interactive login. A Codex API key waits in [M99](../../docs/milestones/M99-parking-lot.md#agent-extras).

## Records and replay

One run has one item, `task`, and one step whose effect is always `paid`. Persist its intent and finish the [durability barrier](../../docs/north-star/contract.md#record-format) before launch. A failed barrier starts no task. Once launch is possible, the step stays `unknown` until a verified answer is saved. A failure proven to precede launch may be `failed`.

The run's `inputs` hold the task's SHA-256, absolute working directory, `tools: true`, and any explicitly supplied model and effort. A value omitted from the call is omitted from these stated settings, so a configuration change does not change the call's identity. `resolved` holds `harness: "codex"`, the actual model, and the actual effort. Compare stated inputs, not newly resolved defaults, on replay.

Save `answer.txt` under `runs/<run-id>/` in the state directory. Record its path, size, and SHA-256 as the step's artifact. Step evidence contains `format_version: 1` and `duration_ms`. Publish the answer and completed record before printing success. Both initial and repeated answers come from that committed evidence.

[Run IDs](../../docs/north-star/contract.md#run-ids) governs locking and replay. The same completed call returns the verified saved answer without starting Codex, even if configuration or the installed harness later changes. Changed stated inputs fail with `USAGE_INVALID`. Unreadable records, unsupported versions, changed artifacts, and uncertain paid effects fail with `RESUME_UNSAFE`.

This task supports no explicit resume. A record with no persisted effect intent can retry under the existing run-ID rules. An unknown paid step cannot. Use `ncly run view` to inspect it and let the human decide whether to start a distinct task. A record is not a background worker.

## Depth limit

`ncly agent run` checks `[agent] max_depth`, 2 by default, before preparation. `NCLY_AGENT_DEPTH` defaults to 0. At or above the limit, answer `AGENT_DEPTH_LIMIT`. A launch receives the depth plus one.

## Error codes

| Code | Exit | When | Since |
|---|---|---|---|
| `AGENT_DEPTH_LIMIT` | 2 | Starting Codex would exceed the configured agent depth | M09 |
