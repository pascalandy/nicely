# M2 Agent

Status: planned
Release: v0.2.0

## Goal

`ncly agent` exposes the internal agent service already used by transcript in M1. It adds general tasks and the remaining harness adapters, and the `headless` skill's launcher retires after validation.

## Depends on

M1, which brings profiles, the internal agent service, run records, and `internal/run`.

## Scope

**Spec first.** Move the commands and error codes below into [cli-spec.md](../north-star/cli-spec.md) before writing Go.

**Commands.**

```
ncly agent run [task] [--profile <name>] [--input <file> | -] [--cwd <dir>] [--check-unchanged] [--dry-run] [--json]
ncly agent profile list [--json]
ncly agent profile view <name> [--json]
```

- The argument is the task. Material to work on comes from `--input` or stdin.
- Material runs with the harness's tools turned off, as D028 requires. A task without material may use tools.
- The answer goes to stdout. `--json` uses the common envelope and returns `run_id`, `answer`, `profile`, `harness`, `model`, and `duration_ms`
- `--check-unchanged` compares tracked and untracked files under `--cwd` after the run. Ignored files do not count. It reports a change but does not prevent writes, undo them, or inspect effects outside that folder. No `--read-only` flag promises protection that this check cannot provide
- `--dry-run` uses shared preparation and shows the profile, requested mode, and redacted harness invocation without starting the harness
- Profile inspection states the adapter's supported modes and any compatibility check still pending. An unsupported tools-off request fails before launching outside material
- Runs use the M1 record contract. General harness tasks are inspectable but not resumable unless their adapter and domain have evidence that supports safe continuation. A nonzero exit or missing answer never justifies launching the task again

```bash
cat notes.md | ncly agent run "Extract the action items"
ncly agent run "Review this branch for risky changes" --profile second-opinion --cwd . --check-unchanged
```

**Harnesses.** Extend M1's Claude Code and Pi adapters with Codex, Grok, and OpenCode. Use the `headless` launcher's existing adapters and notes where available. Each adapter declares supported versions, tools-off behavior, input delivery, cancellation, and result parsing. Validate its declared modes on the real installed harness before enabling them.

**Depth limit.** The public command uses the internal service's M1 depth check. At `max_depth` in `[agent]`, 2 by default, the service refuses to launch. `internal/run` passes `NCLY_AGENT_DEPTH` plus one to each harness.

**Transcript keeps the same owner.** Its summary already uses the internal agent service in M1. Adding the public command changes neither transcript's steps nor its profile and run-record contracts.

**Doctor.** A new `agent` component checks that the harness of each profile is installed. It never starts a harness, because a run can bill.

**New error codes.**

| Code | Exit | When |
|---|---|---|
| `FILES_CHANGED` | 1 | Files covered by `--check-unchanged` changed during the run |

`PREREQ_MISSING`, `DEPTH_LIMIT`, and `CAPABILITY_UNSUPPORTED` reuse the M1 registry entries.

An unknown profile exits 2 with `NOT_FOUND`. After a harness starts, a runtime failure exits 1 and must never be rerun automatically. Cancellation retains 130 or 143 and takes precedence over item failures.

**Follow-up in `pascalandy/skills`.** The `headless` skill points to `ncly agent`. Pascal decides when.

## To define when M2 starts

- Whether `ncly agent run` reads stdin without `-` when stdin is a pipe, given that some harnesses leave stdin open.
- The verified tools-off flags and supported versions of the newly added harnesses. An adapter's declaration must match its observed behavior
- Whether `ncly agent model list` replaces the `list models` command of the transcript CLI.
- The key variables that each harness reads, such as `ANTHROPIC_API_KEY`. Its adapter declares them, so the environment rule of cli-spec.md passes them to the harness.

## Out of scope

- The built-in code review modes of the harnesses: [M99](M99-parking-lot.md).
- Write-prevention sandboxes. `--check-unchanged` is an observation after execution, and tools-off mode remains required for outside material
- Cost tracking.

## Acceptance

Each scenario puts stub executables in `PATH` for the harnesses it needs. A stub prints a fixed answer, so no scenario costs anything.

Reuse the M0 contract suite. Add distinct scenarios for unsupported modes, malformed results, recorded uncertain exits, file changes detected after execution, and cancellation. A stub that accepts a flag does not prove that a real harness enforces it.

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
env NCLY_AGENT_DEPTH=0
exits 78 ncly agent run "Say hello" --profile second-opinion --json
stderr '"code":"PREREQ_MISSING"'
```

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] A real run succeeds with each of the five harnesses Pascal has installed, and the modes declared for each have been verified
- [ ] `ncly transcript run` summarizes through the agent package
- [ ] v0.2.0 is tagged and released
