---
name: nicely-runs
description: Use before a recorded ncly command, or after the answer of one is lost.
---

# Nicely runs

A recorded command saves a record of its run, and its describe sets `modes.recorded`. Name the run before the call, so that a lost answer never causes an automatic second charge.

- Give a recorded command your own new `--run-id`, such as `review-1`, and keep it. After a lost answer, inspect that run with `ncly run view ID --json`. The same call with the same ID answers from its record without new work only when `run.status` is `completed`. For any other status, the same call can start paid work again or refuse: dry-run it first, and follow the skill of its extension
- Without an ID, list the runs of the command with `ncly run list --path 'PATH' --key KEY --json`, newest first. Check `more`, confirm a candidate with `ncly run view ID --json`, its inputs and every item, and report several matches to the human instead of picking the newest one. A run under `unreadable` may still have billed
- `ncly run view` exits 0 for a run that failed: read `run.status`. While `run.active` is true, a process still holds the run: never start the same work beside it
