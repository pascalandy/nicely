# M2 Agent

Status: planned
Release: v0.2.0

## Goal

`ncly agent` runs any task through a harness and a profile. The summary step of transcript moves to Go, and the `headless` skill's launcher retires.

## Depends on

M1, which brings the profiles in the config and `internal/run`.

## Scope

**Spec first.** Move the commands and error codes below into [cli-spec.md](../north-star/cli-spec.md) before writing Go.

**Commands.**

```
ncly agent run [task] [--profile <name>] [--input <file> | -] [--cwd <dir>] [--read-only] [--dry-run] [--json]
ncly agent profile list [--json]
ncly agent profile view <name> [--json]
```

- The argument is the task. Material to work on comes from `--input` or stdin.
- Material runs with the harness's tools turned off, as D028 requires. A task without material may use tools.
- The answer goes to stdout. `--json` returns `ok`, `answer`, `profile`, `harness`, `model`, and `duration_ms`.
- `--read-only` fails the run when a tracked or untracked file under `--cwd` changed. Ignored files do not count.
- `--dry-run` prints the harness command it would run, without starting the harness.

```bash
cat notes.md | ncly agent run "Extract the action items"
ncly agent run "Review this branch for risky changes" --profile second-opinion --cwd . --read-only
```

**Harnesses.** Port the adapters for Claude Code, Codex, and Grok from the `headless` skill's launcher, `scripts/headless.py`, and its notes per harness. Write new adapters for Pi and OpenCode, which the launcher does not run yet. Each adapter knows how to turn its harness's tools off.

**Depth limit.** `internal/run` already passes `NCLY_AGENT_DEPTH` plus one to each harness. At `max_depth` in `[agent]`, 2 by default, `ncly agent run` refuses to start. Depth 2 lets an agent launched by `ncly` launch one more.

**Transcript moves onto the agent package.** Go runs the summary step, and the Python program stops at the transcript. `--profile` keeps its meaning, because the profiles have lived in the config since M1.

**Doctor.** A new `agent` component checks that the harness of each profile is installed. It never starts a harness, because a run can bill.

**New error codes.**

| Code | Exit | When |
|---|---|---|
| `HARNESS_MISSING` | 78 | The harness of the profile is not installed |
| `DEPTH_LIMIT` | 2 | `NCLY_AGENT_DEPTH` reached `max_depth` |
| `READ_ONLY_VIOLATION` | 1 | Files changed during a `--read-only` run |

An unknown profile exits 2 with `NOT_FOUND`. After a harness starts, a failure exits 1 and must never be rerun automatically.

**Follow-up in `pascalandy/skills`.** The `headless` skill points to `ncly agent`. Pascal decides when.

## To define when M2 starts

- Whether `ncly agent run` reads stdin without `-` when stdin is a pipe, given that some harnesses leave stdin open.
- The flags that turn the tools off in each harness, starting from the `claude` flags of the transcript CLI.
- Whether `ncly agent model list` replaces the `list models` command of the transcript CLI.

## Out of scope

- The built-in code review modes of the harnesses: [M99](M99-parking-lot.md).
- Sandbox options beyond `--read-only` and the tool-off mode.
- Cost tracking.

## Acceptance

Each scenario puts stub executables in `PATH` for the harnesses it needs. A stub prints a fixed answer, so no scenario costs anything.

```
# A task runs through the default profile
exec ncly agent run "Say hello" --json
stdout '"harness":"claude"'

# Stdin becomes the material
stdin notes.txt
exec ncly agent run "Summarize"
stdout 'stub answer'

# The depth limit stops recursion
env NCLY_AGENT_DEPTH=2
exits 2 ncly agent run "Say hello" --json
stderr '"code":"DEPTH_LIMIT"'

# A missing harness needs a human. This archive ships a claude stub only.
exits 78 ncly agent run "Say hello" --profile second-opinion --json
stderr '"code":"HARNESS_MISSING"'
```

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] A real run succeeds with each of the five harnesses Pascal has installed
- [ ] `ncly transcript run` summarizes through the agent package
- [ ] v0.2.0 is tagged and released
