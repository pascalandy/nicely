# M2 Agent

Status: planned
Release: v0.2.0

## Goal

`ncly agent` runs any task through a harness and a profile. Profiles live in one place, transcript uses them, and the `headless` skill's launcher retires.

## Depends on

M1.

## Scope

**Spec first.** Add an `ncly agent` section to [cli-spec.md](../north-star/cli-spec.md), with the error codes below, before writing Go.

**Commands.**

```
ncly agent run [task] [--profile <name>] [--input <file> | -] [--cwd <dir>] [--read-only] [--json]
ncly agent profiles [--json]
ncly agent profiles show <name> [--json]
```

- The argument is the task. Material to work on comes from stdin or `--input`.
- The answer goes to stdout. `--json` returns `answer`, `profile`, `harness`, `model`, and `duration_ms`.
- `--read-only` fails the run when a tracked or untracked file under `--cwd` changed. Ignored files do not count.

```bash
cat notes.md | ncly agent run "Extract the action items"
ncly agent run "Review this branch for risky changes" --profile second-opinion --cwd . --read-only
```

**Profiles.**

```toml
[agent]
default_profile = "everyday"
max_depth = 2

[agent.profiles.everyday]
harness = "claude"
model = "<model id>"
effort = "medium"

[agent.profiles.second-opinion]
harness = "codex"
model = "<model id>"
effort = "high"
```

**Harnesses.** Port the adapters for Claude Code, Codex, and Grok from the `headless` skill's launcher, `scripts/headless.py`, and its notes per harness. Write new adapters for Pi and OpenCode, which the launcher does not run yet.

**Depth limit.** Each child agent inherits `NCLY_AGENT_DEPTH` plus one. At `max_depth`, `ncly agent run` refuses to start.

**Transcript moves onto profiles.**

1. Go runs the summary step through the agent package. The Python component stops at the transcript.
2. `--profile` on `ncly transcript run` names an agent profile.
3. `ncly transcript list profiles` is removed in favor of `ncly agent profiles`.

**Doctor.** A new `agent` component checks that the harness of each profile is installed. It never starts a harness, because a run can bill.

**New error codes.**

| Code | Exit | When |
|---|---|---|
| `HARNESS_MISSING` | 78 | The harness of the profile is not installed |
| `DEPTH_LIMIT` | 2 | `NCLY_AGENT_DEPTH` reached `max_depth` |
| `READ_ONLY_VIOLATION` | 1 | Files changed during a `--read-only` run |

An unknown profile exits 2 with `NOT_FOUND`. After a harness starts, a failure exits 1 and must never be rerun automatically.

**Follow-up in `pascalandy/skills`.** The `headless` skill points to `ncly agent`. Pascal decides when.

## Out of scope

- The built-in code review modes of the harnesses: [M99](M99-parking-lot.md).
- Sandbox options beyond `--read-only`.
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
- [ ] `ncly transcript run` summarizes through an agent profile
- [ ] v0.2.0 is tagged and released

## Open questions

1. Is a default `max_depth` of 2 right? It allows one agent launched by `ncly` to launch one more. Recommendation: start with 2 and raise it only for a real need.
