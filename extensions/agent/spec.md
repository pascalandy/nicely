# agent extension spec

`agent` is a first-party external extension, built as `ncly-agent` from this folder. It runs a task through a harness with a profile from the config, and it owns every harness behavior: profiles, adapters, tools-off mode, and the depth limit. It stays out of core because Nicely needs no agent to operate. Other extensions use it through the CLI, as transcript does for its summary. It follows the [contract](../../docs/north-star/contract.md). [M09](../../docs/milestones/M09-agent-profiles.md) builds profiles and dry run, [M10](../../docs/milestones/M10-agent-run.md) execution, and [M11](../../docs/milestones/M11-harnesses.md) the other harnesses.

## ncly agent

```
ncly agent run [task] [--profile <name> | --harness <name> --model <id> --effort <level> [--provider <name>]] [--input <file> | -] [--cwd <dir>] [--check-unchanged] [--timeout <duration>] [--dry-run] [--json]
ncly agent profile list [--json]
ncly agent profile view <name> [--json]
```

- The argument is the task. Material to work on comes from `--input` or stdin
- Material runs with the harness's tools turned off, as [D028](../../docs/north-star/decisions/D028-outside-text.md) requires. A task without material may use tools
- The answer goes to stdout. `--json` uses the common envelope and returns `run_id`, `answer`, `profile`, `harness`, `model`, and `duration_ms`
- `--harness`, `--model`, `--effort`, and `--provider` describe a profile on the command line, without reading the config, and exclude `--profile`. A caller that recorded a resolved profile, such as transcript for a resume, passes it this way, so a later change to the config never changes the run
- `--dry-run` uses the shared preparation and shows the resolved profile, the requested mode, and the redacted harness invocation without starting the harness. It resolves the profile and checks the depth, the harness version, and tools-off support, so another extension can check a run and record its profile before any paid work, as transcript does before it transcribes
- `--check-unchanged` compares tracked and untracked files under `--cwd` after the run. Ignored files do not count. It reports a change but does not prevent writes, undo them, or inspect effects outside that folder. No `--read-only` flag promises protection that this check cannot provide
- Profile inspection states the adapter's supported modes and any compatibility check still pending. An unsupported tools-off request fails before launching outside material
- A run records its execution before the harness starts, as [Records and evidence](../../docs/north-star/contract.md#records-and-evidence) defines. Its one step declares the `paid` effect. It turns `unknown` before the harness starts, and `failed` only when the harness never started. A general harness task is inspectable but not resumable, and a nonzero exit or a missing answer never justifies launching the task again
- Every failure before the harness starts exits 2 or 78: `NOT_FOUND` for an unknown profile, `PREREQ_MISSING` for a missing or too old harness, `CAPABILITY_UNSUPPORTED` for a mode that the adapter cannot enforce, and `AGENT_DEPTH_LIMIT`. After a harness starts, a failure exits 1 and must never be rerun automatically. Cancellation keeps 130 or 143 and takes precedence over item failures. A caller, such as transcript, relies on this split to tell a harness that never started from one that may have billed

```bash
cat notes.md | ncly agent run "Extract the action items"
ncly agent run "Review this branch for risky changes" --profile second-opinion --cwd . --check-unchanged
```

## Profiles

A profile names a harness, a model, and an effort, plus a provider when the harness serves several. Profiles live in the config, and every command that runs an agent reads them there, except a resume, which uses the profile that its run recorded without reading the config again.

```toml
[agent]
default_profile = "everyday"

[agent.profiles.everyday]
harness = "claude"
model = "<model id>"
effort = "medium"

[agent.profiles.second-opinion]
harness = "pi"
provider = "openai-codex"
model = "<model id>"
effort = "high"
```

| Key | Meaning |
|---|---|
| `harness` | `claude`, `codex`, `grok`, `pi`, or `opencode` |
| `provider` | The model provider, for a harness that serves several, such as `pi` or `opencode` |
| `model` | A model ID that the harness accepts |
| `effort` | A reasoning effort that the harness accepts |

`--profile <name>` picks a profile, and `default_profile` applies without it. An unknown profile exits 2 with `NOT_FOUND`. Without `--profile` and without `default_profile`, no profile applies, and each command that runs an agent states what it does then.

A profile with an effort that its harness rejects, or with a `provider` for `claude`, is `CONFIG_INVALID` when that profile is selected. Syntax errors and values of the wrong declared type still fail as Configuration defines. An unused profile's semantic constraint blocks no other command.

## Adapters

M09 builds the adapters for `claude` and `pi`, and M11 adds `codex`, `grok`, and `opencode`. Until then, a profile that names another harness fails with `CAPABILITY_UNSUPPORTED`. Each adapter turns the tools off, sends the task on stdin, and checks the answer:

- `claude`, version 2.1.291 or later, runs with `--print`, `--model`, `--effort`, `--system-prompt`, and `--output-format json`. It turns tools off with `--tools ""`, `--strict-mcp-config --mcp-config '{"mcpServers":{}}'`, `--setting-sources ""`, `--disable-slash-commands`, `--no-session-persistence`, and `--permission-mode dontAsk`. `--settings` turns hooks off and sets `CLAUDE_CODE_EFFORT_LEVEL`, which would otherwise override `--effort`. The answer is `result`, and `is_error` must be `false`. The effort is `low`, `medium`, `high`, `xhigh`, or `max`, and a profile for `claude` has no `provider`
- `pi`, version 1.0.4 or later, runs with `--print`, `--model <provider>/<model>`, `--thinking`, and `--system-prompt`. It turns tools off with `--no-tools`, `--no-session`, `--no-skills`, `--no-prompt-templates`, `--no-context-files`, `--no-extensions`, and `--no-approve`. The answer is stdout, which must not be empty. The effort is `off`, `minimal`, `low`, `medium`, `high`, `xhigh`, or `max`

A harness older than its listed version is `PREREQ_MISSING`. These invocations come from the transcript CLI, and their flags were checked against `--help` on 2026-10-07. M10 rechecks the real adapters and observes that material cannot invoke tools, hooks, MCP servers, or loaded extensions. A stub accepting flags or a real harness returning text is not that proof.

Each adapter declares the key variables of its harness as optional keys: `claude` reads `ANTHROPIC_API_KEY`, and `pi` uses its adapter's verified provider-to-variable mapping, such as `OPENROUTER_API_KEY`. An adapter never derives a key variable by parsing translated help at runtime. The harness receives a key only when a human granted it to `agent`, and otherwise uses its own login, as [Programs that ncly runs](../../docs/north-star/contract.md#programs-that-ncly-runs) defines.

## Depth limit

`ncly agent run` checks `[agent] max_depth`, 2 by default, before every launch. `NCLY_AGENT_DEPTH` defaults to 0. A launch at or above the limit exits 2 with `AGENT_DEPTH_LIMIT`. The harness receives the depth plus one, so agents cannot launch agents without end.

## Error codes

| Code | Exit | When | Since |
|---|---|---|---|
| `AGENT_DEPTH_LIMIT` | 2 | Starting a harness would exceed the configured agent depth | M09 |
| `AGENT_FILES_CHANGED` | 1 | Files covered by `--check-unchanged` changed during the run | M11 |
