---
name: agent
description: Use when running a trusted task through Codex with ncly agent.
---

# agent

`ncly agent` works through the login that Codex already has. Nicely gives Codex no key.

- Check that Codex is installed and recent enough with `ncly doctor agent`. When its check `agent.codex` fails with `PREREQ_MISSING`, relay the hint: a human installs or updates Codex
- Give the task as one argument, or in a file with `--task-file`. Use only a task you trust: Codex may run commands and write in its folder, which `--cwd` selects
- Always add `--tools`. Codex cannot run a task with its tools turned off, so a call without it answers `CAPABILITY_UNSUPPORTED`
- Set the model with `--model`, or with `model` in the `[agent]` table of the config. The effort comes from `--effort` or `effort` in `[agent]`, and is `medium` by default
- On `AGENT_DEPTH_LIMIT`, the call already runs inside as many agents as `max_depth` in `[agent]` allows: do the task yourself, or relay the hint to the human. Never raise `max_depth` or change `NCLY_AGENT_DEPTH` yourself
- Check the task with `ncly agent run "<task>" --tools --model <id> --dry-run --json` first: it checks Codex and the settings, shows the invocation, and starts no task
- Then run it with `ncly agent run "<task>" --tools --model <id> --run-id <id> --json`, with a new ID of your choice, such as `review-1`: the answer holds `answer` and `run_id`, and the record of the run keeps that ID even when the answer is lost
- To get a lost answer back, first make the same call with the same `--run-id` and `--dry-run`. When the step `codex` of its plan has the action `reuse`, the run completed: the same call without `--dry-run` answers from the record, and Codex does not run again. When the action is `run`, no task started under that ID, and the call starts one, which Codex may bill. `USAGE_INVALID` with that ID means that the ID names another call
- The call stops Codex after `--timeout`, `5m` by default. Give a long task a longer one, such as `--timeout 30m`
- After `INTERRUPTED`, `TERMINATED`, or `RUNTIME`, inspect the run with the hint, `ncly run view <id>`. A step `failed` ran no task. A step `unknown` means that Codex may have billed without a saved answer. Either way, the same call answers `RESUME_UNSAFE`. Never start the task again under a new ID on your own: relay the hint to the human
