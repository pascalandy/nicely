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
