---
name: nicely
description: Use before running any ncly command.
---

# Nicely

- Set `NCLY_NO_INPUT=1` and `NCLY_JSON=1`
- List commands with `ncly describe --json`, then read one with `ncly describe PATH --json`: its flags, effects, modes, and result keys
- Run a command that writes or bills with `--dry-run` first. The dry run checks what that call needs
- Branch on the exit code and `errors[].code`, never on `message` or `hint`. `next` holds the command to run as words
- With `--json`, stdout stays empty on command failure. Read the error object from stderr's last line
- Exit 2: fix the call. Exit 78: relay `hint` and `next` to the human. Exit 1: inspect before anything else. Only exit 75 permits repeating the same call
- An item error describes its item and never permits a retry
- Run `ncly doctor --json` to check the whole machine at setup or for an unexplained dry-run problem
- Select one extension with its slug, such as `ncly doctor video-convert --json`
- Doctor and each status command marked `doctor` list findings under `checks`, on stdout. On exit 78, relay the `hint` and `next` of each failed check. Another report uses the key that describe gives it. `errors` holds only a failure of the command
- Give the shell a timeout longer than the command's `--timeout` plus 15 seconds of cleanup
- Each extension ships a skill named after its slug, its root's words joined with `-`: `ncly skill view video-convert`. `ncly skill list` lists every skill, and `ncly skill status --json` checks them. The skill domain has none: this skill covers it

## Read the skill of a topic

- Before a command whose describe lists `requires.keys`, or on `AUTH_MISSING`, `KEY_NOT_GRANTED`, or `TERMINAL_REQUIRED`, read `ncly skill view nicely-auth`. Never type, pipe, or grant a key
- Before a command whose describe sets `modes.recorded`, or after a lost answer, read `ncly skill view nicely-runs`
- To install, build, or remove an extension, read `ncly skill view nicely-extensions`
