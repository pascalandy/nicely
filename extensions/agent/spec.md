# Agent extension spec

`agent` is an external extension, built as `ncly-agent` from this folder. It runs one trusted task through Codex. [M09](../../docs/milestones/M09-agent-profiles.md) adds preparation and [M10](../../docs/milestones/M10-agent-run.md) adds execution and inspection. It follows the [contract](../../docs/north-star/contract.md).

[Deferred agent features](../../docs/milestones/M99-parking-lot.md#agent-extras) include other harnesses, named profiles, providers, model discovery, a Codex API key, and post-run file comparison. They are not prerequisites of this extension.

## Install

`just install`, from a checkout, builds `ncly` into `~/.local/bin/`. It builds each first-party external extension, such as `ncly-agent`, into the `extensions` folder of Nicely's data folder, such as `~/.local/share/nicely/extensions/`. It writes the extension's manifest, catalogs, and skill beside it, as `ncly-agent.toml`, `ncly-agent.locales/`, and `ncly-agent.skill/`, where [Discovery](../../docs/north-star/contract.md#discovery) finds them. It honors an absolute `XDG_DATA_HOME`. It builds every file before it replaces an installed one, so a failed build keeps the previous install.

## ncly agent

```
ncly agent run (<task> | --task-file <file>) --tools [--model <id>] [--effort <level>] [--cwd <dir>] [--run-id <id>] [--timeout <duration>] [--dry-run] [--json]
```

- Exactly one task argument or `--task-file` is required: neither or both fail with `USAGE_INVALID`. The file contains the trusted task instructions, and a relative path reads from the current directory. It must be a regular file: a folder, a pipe, a device, or `/dev/stdin` fails with `USAGE_INVALID` before anything reads it. An unreadable file fails with `USAGE_INVALID` before launch, and so does a task that holds only whitespace. The task counts by its SHA-256, the same for both sources.
- `--cwd` selects an existing directory, or defaults to the current directory. A relative path resolves from the current directory, and Codex receives the absolute path. An invalid directory fails with `USAGE_INVALID` before launch.
- The adapter requires explicit `--tools`. Without it, even in a dry run, the command answers `CAPABILITY_UNSUPPORTED` before launching Codex. Never enable tools silently.
- External material is outside this first command. It accepts no `--input` flag. Do not put a transcript or web page into `--task-file` to bypass [D028](../../docs/north-star/decisions/D028-outside-text.md).
- Nicely reads no stdin for this command. It sends the prepared task to Codex on stdin and closes that pipe after the task.
- Dry run shares preparation with execution. Preparation checks, in this order, the `[agent]` settings, the depth, the task, the model, the directory, the Codex version, and the tools mode, and answers the first problem. The dry run then shows the resolved settings and the redacted invocation, as [Dry run](#dry-run) defines, starts no task, and records nothing. Login validity remains a pending execution check; a dry run makes no paid request. Until [M10](../../docs/milestones/M10-agent-run.md) adds execution, a call without `--dry-run` answers `CAPABILITY_UNSUPPORTED` once preparation passes, and starts no task.
- The default timeout is 5 minutes. `--timeout` overrides it under the common duration rules.
- Text mode prints the answer. JSON uses the common envelope and returns `run_id`, `answer`, `harness`, `model`, and `duration_ms`. `harness` is `codex`. The record's `tools` holds the harness version.
- Preparation errors use `USAGE_INVALID`, `CONFIG_INVALID`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, or `AGENT_DEPTH_LIMIT`. Failures after launch exit 1. Signals retain exits 130 and 143. No failure after launch permits an automatic new task.

```bash
ncly agent run "Review this branch for risky changes" --tools --model '<model id>' --dry-run
ncly agent run --task-file task.md --tools --run-id "$RUN_ID" --json
ncly run view "$RUN_ID" --json
```

### Dry run

`--dry-run --json` answers with the keys of [Dry-run answer](../../docs/north-star/contract.md#dry-run-answer) and these keys:

| Key | Type | Meaning |
|---|---|---|
| `harness` | string | `codex` |
| `harness_version` | string | The version that `codex --version` printed, such as `0.163.0` |
| `model` | string | The model, from `--model` or `[agent] model` |
| `effort` | string | The effort, from `--effort`, `[agent] effort`, or `medium` |
| `cwd` | string | The absolute folder where Codex works |
| `task_sha256` | string | The SHA-256 of the task, in lowercase hexadecimal |
| `invocation` | array | The path of Codex, then the words that execution passes to it, as [Codex invocation](#codex-invocation) lists them. `<answer-file>` stands for the file that execution creates for the answer. The task never appears, because Codex reads it on stdin, and no key does |

`plan` holds one item, whose `key` is `task` and whose `output` is the `runs` folder of the state directory, such as `~/.local/state/nicely/runs`. Its one step is `codex`, with the action `run`, and its `effects` set `user_writes`, `network`, and `paid` to `true`. `pending_checks` holds `agent.login`, because Codex checks its own login only when it runs the task.

```json
{"ok":true,"contract_version":1,"cwd":"/Users/me/project","dry_run":true,"effort":"medium","harness":"codex","harness_version":"0.163.0","invocation":["/opt/homebrew/bin/codex","exec","--model","<model id>","-c","model_reasoning_effort=\"medium\"","--sandbox","workspace-write","--skip-git-repo-check","--cd","/Users/me/project","--json","--output-last-message","<answer-file>","-"],"model":"<model id>","pending_checks":[{"id":"agent.login","message":"Codex checks its login only when it runs the task."}],"plan":[{"key":"task","output":"/Users/me/.local/state/nicely/runs","steps":[{"id":"codex","action":"run"}],"effects":{"network":true,"paid":true,"user_writes":true}}],"task_sha256":"<sha256>"}
```

Without `--json`, the dry run prints the invocation as a POSIX shell reads it back, then the model, the effort, the folder, and each pending check, as plain text.

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

Codex version 0.162.1 or later runs these words, in this order:

```
codex exec --model <model> -c model_reasoning_effort="<effort>" --sandbox workspace-write --skip-git-repo-check --cd <dir> --json --output-last-message <answer-file> -
```

Each value is one word, so no shell reads it. The final `-` sends the task through stdin. Keep Codex's sandbox enabled. No Nicely flag promises that Codex cannot write.

Stdout contains JSONL events. The answer comes from the output file and must contain text after a zero exit. A nonzero exit or missing answer after launch leaves the step unknown. Supported effort values are `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`, subject to the model. Preparation checks the known effort values without a model-catalog API; unsupported model choices fail at execution.

Codex has no verified switch that removes all tools, so this adapter supports only explicit `--tools`. The flags above retain the help check recorded on 2026-10-16 in [PR #156](https://github.com/pascalandy/nicely/pull/156). Real task compatibility still needs the independent installed-harness check in M10. The adapter uses `sdk/runner` for execution, cancellation, and environment filtering.

The manifest declares `codex` in `programs` as a required program, not `conditional`, with `version = "0.162.1"` and the default probe `--version`. Its `install` holds `brew install --cask codex` and `sudo pacman -S openai-codex`. Existing `ncly doctor agent` checks that declaration, as [Required programs](../../docs/north-star/core-spec.md#required-programs) defines: the `agent.codex` check fails with `PREREQ_MISSING` for a missing or older Codex, or a probe that prints no version, and its hint and `next` hold the install command of this machine. Doctor only runs `codex --version`, so it starts no task. Preparation runs the same check through the program check of `sdk/runner`, with the environment that Codex receives and a limit of 10 seconds, so a missing, older, or silent Codex fails `ncly agent run` with the same `PREREQ_MISSING`, hint, and `next`. A probe that does not answer within the limit fails with `PREREQ_MISSING` too, whose hint asks to check that `codex --version` answers. A separate `agent status` command and a matrix of configured harnesses would duplicate that first check.

Codex uses its own existing login. `agent` declares no API key and receives none from Nicely. Keep the runner's removal of `*_API_KEY` variables. A user configures the native login outside Nicely; the extension never starts an interactive login. A Codex API key waits in [M99](../../docs/milestones/M99-parking-lot.md#agent-extras).

## Records and replay

One run has one item, `task`, and one step, `codex`, whose effect is always `paid`. Persist its intent and finish the [durability barrier](../../docs/north-star/contract.md#record-format) before launch. A failed barrier starts no task. Once launch is possible, the step stays `unknown` until a verified answer is saved. A failure proven to precede launch may be `failed`.

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
