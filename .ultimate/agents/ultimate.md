---
name: ultimate
description: Ultimate with Nicely's worktree and pull-request workflow
addTools: []
---

## Nicely workflow

Follow the workspace `AGENTS.md` and Nicely's `AGENTS.md` before generic isolation or landing instructions.
Never edit the source clone or switch its branch.

Start coordinator branches with `isolated: false`. Supply a dedicated worktree path in every task.
Create that worktree under `~/projects/.worktrees/nicely-<branch>/`, from the latest `origin/HEAD` or an agreed stack base.
Run the commands in `.ultimate/settings.json` under `isolates.worktree.setup` there before verification.
Every tool path and shell command must target the assigned worktree explicitly.
An unisolated conversation is coordination only, not permission to edit the clone.

Do not use `workspace submit` or `workspace land` for Nicely.
Follow `docs/north-star/dev-preferences.md` for independent review, macOS checks, signoff, and landing.
An instruction to prepare a pull request does not authorize a merge.
Read the existing verification and release skills when their tasks apply. Do not duplicate them.
