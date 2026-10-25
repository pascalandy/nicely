# Agent extension spec

`agent` is an external extension, built as `ncly-agent` from this folder. It runs one trusted task through Codex. [M09](../../docs/milestones/M09-agent-profiles.md) added preparation and [M10](../../docs/milestones/M10-agent-run.md) added execution and inspection. It follows the [contract](../../docs/north-star/contract.md).

[Deferred agent features](../../docs/backlog/XT-106-agent.md#later) include other harnesses, named profiles, providers, model discovery, a Codex API key, and post-run file comparison. They are not prerequisites of this extension.

## Install

`just install`, from a checkout, builds `ncly` into `~/.local/bin/`. It builds each first-party external extension, such as `ncly-agent`, into the `extensions` folder of Nicely's data folder, such as `~/.local/share/nicely/extensions/`. It writes the extension's manifest, catalogs, and skill beside it, as `ncly-agent.toml`, `ncly-agent.locales/`, and `ncly-agent.skill/`, where [Discovery](../../docs/north-star/contract/discovery.md) finds them. It honors an absolute `XDG_DATA_HOME`. It builds every file before it replaces an installed one, so a failed build keeps the previous install.

## ncly agent

```
ncly agent run (<task> | --task-file <file>) --tools [--model <id>] [--effort <level>] [--cwd <dir>] [--run-id <id>] [--timeout <duration>] [--dry-run] [--json]
```

- Exactly one task argument or `--task-file` is required: neither or both fail with `USAGE_INVALID`. The file contains the trusted task instructions, and a relative path reads from the current directory. It must be a regular file other than the input of `ncly`: a folder, a pipe, a device, or `/dev/stdin`, even when a file feeds it, fails with `USAGE_INVALID` before anything reads it. An unreadable file fails with `USAGE_INVALID` before launch, and so does a task that holds only whitespace. The task counts by its SHA-256, the same for both sources.
- `--cwd` selects an existing directory, or defaults to the current directory. A relative path resolves from the current directory, and Codex receives the absolute path. An invalid directory fails with `USAGE_INVALID` before launch.
- The adapter requires explicit `--tools`. Without it, even in a dry run, the command answers `CAPABILITY_UNSUPPORTED` before launching Codex. Never enable tools silently.
- External material is outside this first command. It accepts no `--input` flag. Do not put a transcript or web page into `--task-file` to bypass [D028](../../docs/north-star/decisions/D028-outside-text.md).
- Nicely reads no stdin for this command. It sends the prepared task to Codex on stdin and closes that pipe after the task.
- Dry run shares preparation with execution. Preparation checks, in this order, the `[agent]` settings, the depth, the task, the directory, and the tools mode, which the call states, then the model and the Codex version, which a new task needs, and answers the first problem. A call whose run ID a record holds decides on that record between the two, as [Records and replay](#records-and-replay) defines, so a replay needs neither a model nor Codex. The dry run then shows the resolved settings and the redacted invocation, as [Dry run](#dry-run) defines, starts no task, and records nothing. Login validity remains a pending execution check; a dry run makes no paid request. Without `--dry-run`, the call then runs the task under its run ID, as [Records and replay](#records-and-replay) defines.
- The default timeout is 5 minutes. `--timeout` overrides it under the common duration rules: a duration such as `30s`, `5m`, or `2h`, or a number of seconds, above 0, and any other value fails with `USAGE_INVALID`. The timeout covers the whole call, from the start of preparation, as [Stops and failures](#stops-and-failures) defines.
- Text mode prints the answer. JSON uses the common envelope and returns `run_id`, `answer`, `harness`, `model`, and `duration_ms`. `harness` is `codex`. The record's `tools` holds the harness version.
- Preparation errors use `USAGE_INVALID`, `CONFIG_INVALID`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, or `AGENT_DEPTH_LIMIT`. Failures after launch exit 1 with `RUNTIME`, whose hint and `next` hold `ncly run view <id>`. Signals retain exits 130 and 143. No failure after launch permits an automatic new task.

```bash
ncly agent run "Review this branch for risky changes" --tools --model '<model id>' --dry-run
ncly agent run --task-file task.md --tools --run-id "$RUN_ID" --json
ncly run view "$RUN_ID" --json
```

### Dry run

`--dry-run --json` answers with the keys of [Dry-run answer](../../docs/north-star/contract/dry-run-answer.md) and these keys:

| Key | Type | Meaning |
|---|---|---|
| `harness` | string | `codex` |
| `harness_version` | string | The version that `codex --version` printed, such as `0.163.0` |
| `model` | string | The model, from `--model` or `[agent] model` |
| `effort` | string | The effort, from `--effort`, `[agent] effort`, or `medium` |
| `cwd` | string | The absolute folder where Codex works |
| `task_sha256` | string | The SHA-256 of the task, in lowercase hexadecimal |
| `invocation` | array | The path of Codex, then the words that execution passes to it, as [Codex invocation](#codex-invocation) lists them. `<answer-file>` stands for the file that execution creates for the answer. The task never appears, because Codex reads it on stdin, and no key does |
| `environment` | object | The variables that Nicely sets for Codex, as strings: `NCLY_AGENT_DEPTH`, as [Depth limit](#depth-limit) defines. Codex inherits the rest of the environment of `ncly`, without any `*_API_KEY` variable, which this answer never shows |

`plan` holds one item, whose `key` is `task` and whose `output` is the `runs` folder of the state directory, such as `~/.local/state/nicely/runs`. Its one step is `codex`, with the action `run`, and its `effects` set `user_writes`, `network`, and `paid` to `true`. `pending_checks` holds `agent.login`, because Codex checks its own login only when it runs the task.

```json
{"ok":true,"contract_version":1,"cwd":"/Users/me/project","dry_run":true,"effort":"medium","environment":{"NCLY_AGENT_DEPTH":"1"},"harness":"codex","harness_version":"0.163.0","invocation":["/opt/homebrew/bin/codex","exec","--model","<model id>","-c","model_reasoning_effort=\"medium\"","--sandbox","workspace-write","--skip-git-repo-check","--cd","/Users/me/project","--json","--output-last-message","<answer-file>","-"],"model":"<model id>","pending_checks":[{"id":"agent.login","message":"Codex checks its login only when it runs the task."}],"plan":[{"key":"task","output":"/Users/me/.local/state/nicely/runs","steps":[{"id":"codex","action":"run"}],"effects":{"network":true,"paid":true,"user_writes":true}}],"task_sha256":"<sha256>"}
```

Without `--json`, the dry run prints the invocation as a POSIX shell reads it back, then the model, the effort, the folder, the environment, and each pending check, as plain text. A `reuse` plan prints that the call would answer from the saved run, then the model, the effort, and the folder.

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

Codex has no verified switch that removes all tools, so this adapter supports only explicit `--tools`. The flags above retain the help check recorded on 2026-10-16 in [PR #156](https://github.com/pascalandy/nicely/pull/156). A real task passed the independent installed-harness check of M10, recorded on 2026-10-17 in [PR #191](https://github.com/pascalandy/nicely/pull/191), with Codex 0.163.0 and `ncly` at commit 66d5496. It covered the dry run, the execution, the saved answer, and a replay that started no Codex. The adapter uses `sdk/runner` for execution, cancellation, and environment filtering.

The manifest declares `codex` in `programs` as a required program, not `conditional`, with `version = "0.162.1"` and the default probe `--version`. Its `install` holds `brew install --cask codex` and `sudo pacman -S openai-codex`. Existing `ncly doctor agent` checks that declaration, as [Required programs](../../docs/north-star/core-spec.md#required-programs) defines: the `agent.codex` check fails with `PREREQ_MISSING` for a missing or older Codex, or a probe that prints no version, and its hint and `next` hold the install command of this machine. Doctor only runs `codex --version`, so it starts no task. Preparation runs the same check through the program check of `sdk/runner`, with the environment that Codex receives and a limit of 10 seconds, so a missing, older, or silent Codex fails `ncly agent run` with the same `PREREQ_MISSING`, hint, and `next`. A probe that does not answer within the limit fails with `PREREQ_MISSING` too, whose hint asks to check that `codex --version` answers. A separate `agent status` command and a matrix of configured harnesses would duplicate that first check.

Codex uses its own existing login. `agent` declares no API key and receives none from Nicely. Keep the runner's removal of `*_API_KEY` variables. A user configures the native login outside Nicely; the extension never starts an interactive login. A Codex API key waits in [XT-106](../../docs/backlog/XT-106-agent.md#later).

## Records and replay

One run has one item, `task`, and one step, `codex`, whose effect is always `paid`. Persist its intent and finish the [durability barrier](../../docs/north-star/contract/record-format.md) before launch. A failed barrier starts no task. Once launch is possible, the step stays `unknown` until a verified answer is saved. A failure proven to precede launch may be `failed`.

The run's `inputs` hold the task's SHA-256, absolute working directory, `tools: true`, and any explicitly supplied model and effort. A value omitted from the call is omitted from these stated settings, so a configuration change does not change the call's identity. `resolved` holds `harness: "codex"`, the actual model, and the actual effort. Compare stated inputs, not newly resolved defaults, on replay.

After preparation, the call checks `--run-id` against [Run IDs](../../docs/north-star/contract/run-ids.md), in a dry run too. Without `--dry-run`, it generates an ID when the call gives none, then takes the run lock and reads the record while it holds the lock. For an ID that no record holds, it writes the initial record, with the step `codex` `pending` and the run `running`, then persists the intent: the step turns `unknown`, with its `started_at`. Only then does Codex start, in the run's own temporary folder, `ncly-run-<id>-` and a random suffix in the system temporary directory, which holds `<answer-file>` and is removed when the call ends. The call reads no event from the stdout of Codex. The stderr of Codex reaches the stderr of `ncly` only with `--verbose`.

A record that cannot be written answers `RUNTIME`, exit 1, before Codex starts, with a hint that names the `runs` folder. A failure after the launch, such as a nonzero exit or an answer file without text, answers `RUNTIME`, exit 1, and leaves the step and the run `unknown`. Every answer of a recorded call holds its `run_id`, failures included.

Save `answer.txt` under `runs/<run-id>/` in the state directory. Record its path, size, and SHA-256 as the step's artifact. Step evidence contains `format_version: 1` and `duration_ms`. Publish the answer and completed record before printing success. Both initial and repeated answers come from that committed evidence.

[Run IDs](../../docs/north-star/contract/run-ids.md) governs locking and replay. A call whose ID a record holds takes the first case that applies, before the model and the Codex check:

1. A record that cannot be read, that holds another `format_version`, or whose `resolved` lacks the `harness` `codex`, a model, or a known effort: `RESUME_UNSAFE`
2. A record of another command path, other `inputs`, or other item keys: `USAGE_INVALID`, because the ID names another call
3. A completed run: its verified saved answer, with the `model` of its `resolved` and the `duration_ms` of its evidence, without any call to Codex, even if the configuration or the installed Codex changed since. An `answer.txt` whose size or SHA-256 differs from its artifact answers `RESUME_UNSAFE`, and so does evidence without `format_version` 1 and a `duration_ms` of 0 or more
4. A run whose step `codex` is still `pending`, because its intent was never saved: the call runs the task again under the same ID, with the `model` and the `effort` of the record's `resolved`, after a new Codex check. It writes `attempts` plus one, `running`, and the version of Codex in `tools` before the intent
5. Any other run, such as one whose step is `unknown` or `failed`: `RESUME_UNSAFE`, because Codex may have done paid work

Each of these errors keeps the `run_id`, and its hint and `next` hold `ncly run view <id>`.

A dry run with `--run-id` tries the run lock without waiting and without creating its file, releases it at once, then reads the record. It answers what execution would answer: `TEMPORARY` while another process holds the run, the error of cases 1, 2, and 5, the plan of a new task for an ID that no record holds and for case 4, and, for case 3, a `reuse` plan once it verified the saved answer as execution does, or `RESUME_UNSAFE` for an answer that changed. A `reuse` plan gives its item the `output_dir` of the record instead of `output`, and its step `codex` the action `reuse`, its `effects` are all `false`, and `pending_checks` is empty. `harness_version`, `model`, and `effort` then come from the record, `invocation` is empty, and `environment` holds no variable, because no Codex starts.

### Stops and failures

Ctrl-C, SIGTERM, and the timeout stop Codex and every descendant that `ncly` can reach, with the process-tree cleanup of [Programs that ncly runs](../../docs/north-star/contract/programs-that-ncly-runs.md): the signal, or SIGTERM for the timeout, then SIGKILL 15 seconds later. They stop the Codex check at once, as they stop a version probe. Without `--dry-run`, what the call answers depends on where the stop comes. Each answer keeps the `run_id`, and each answer after the record has `ncly run view <id>` as its hint and `next`:

- Before the record, such as during the Codex check or the wait for the run lock, no record exists: a signal answers with exit 130 or 143, and the timeout answers `RUNTIME`, whose hint asks for a longer `--timeout`
- After the intent, before Codex starts, and for a start that fails, such as a Codex that is no longer executable or a temporary folder that cannot be made, no task ran: the step turns `failed` with its error, and the run `interrupted` after a signal, or `failed`. The call answers `INTERRUPTED` with exit 130, `TERMINATED` with exit 143, or `RUNTIME`
- Once Codex started, its paid work may have happened: the step and the run stay `unknown`, with the error of the stop, and the call answers `INTERRUPTED`, `TERMINATED`, or `RUNTIME` for the timeout. A repeated call under that ID answers `RESUME_UNSAFE`, and nothing starts the task again on its own
- Once Codex exited with its answer, the work is finished, and the record keeps finished work, as the timeout rule of Programs that ncly runs asks: the call still saves the verified answer and the completion. After the timeout, it answers that saved answer, and after a signal, `INTERRUPTED` or `TERMINATED` with the keys of that answer. A repeated call answers the saved answer

A stop during a dry run answers as a stop before the record, without `run_id`. A repeated call after a failed step answers `RESUME_UNSAFE` too, because the record saved the intent of a paid step, as case 5 of Records and replay says. Only a human decides whether a new run ID may start the task again.

This task supports no explicit resume. A record with no persisted effect intent can retry under the existing run-ID rules. An unknown paid step cannot. Use `ncly run view` to inspect it and let the human decide whether to start a distinct task. A record is not a daemon.

## Depth limit

`ncly agent run` reads `[agent] max_depth`, 2 by default, then `NCLY_AGENT_DEPTH`, the number of agents above the call, 0 when unset or empty. At or above the limit, the call answers `AGENT_DEPTH_LIMIT` once it has read the settings, before it checks the task, the model, the folder, Codex, or `--tools`, so it starts no Codex, not even the version probe. Its hint names the file that sets `max_depth`, or the shared config file: only a human decides whether a deeper chain of agents may run. A `max_depth` that is no whole number from 1 answers `CONFIG_INVALID`, and an `NCLY_AGENT_DEPTH` that is no whole number from 0 answers `USAGE_INVALID`.

Codex receives the depth plus one in `NCLY_AGENT_DEPTH`, for the version probe as for the task, so an `ncly agent run` that Codex starts counts one more agent above it. The dry run shows that value in `environment`.

## Error codes

| Code | Exit | When | Since |
|---|---|---|---|
| `AGENT_DEPTH_LIMIT` | 2 | Starting Codex would exceed the configured agent depth | M09 |
