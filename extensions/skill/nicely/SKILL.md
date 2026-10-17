---
name: nicely
description: Use before running any ncly command.
---

# Nicely

`ncly` is a command line that agents operate as well as humans. Follow these rules on every call.

- Set `NCLY_NO_INPUT=1` and `NCLY_JSON=1` for every ncly call
- List commands with `ncly describe --json`, then read one with `ncly describe PATH --json`: its flags, effects, modes, and result keys
- Run a command that writes or bills with `--dry-run` first. The dry run checks what that call needs
- Check the whole machine with `ncly doctor --json` during setup, or when a dry run reports a problem you cannot explain
- List the keys that installed extensions need, and where each one comes from, with `ncly auth status --json`. Its findings are under `services`, and it never prints a key
- On `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, `KEY_NOT_GRANTED`, or `TERMINAL_REQUIRED`, relay `hint` and `next` to the human, who stores the key with `ncly auth login` and grants it to an extension with `ncly auth grant`, or takes a grant back with `ncly auth revoke`. Never type or pipe a key yourself, and never grant one
- Branch on the exit code and `errors[].code`, never on `message` or `hint`. `next` holds the command to run as words
- Exit 2: fix the call. Exit 78: relay `hint` and `next` to the human. Exit 1: inspect before anything else. Only exit 75 permits repeating the same call
- `ncly doctor` and each status command marked `doctor` list their findings under `checks`, on stdout. On exit 78, relay the `hint` and `next` of each failed check to the human. Another report command lists its findings under the key that `ncly describe` gives it. `errors` holds only a failure of the command itself
- An item error describes its item and never permits a retry
- Give the shell a timeout longer than the command's `--timeout` plus 15 seconds of cleanup
- Each extension ships a skill named after its domain: `ncly skill view DOMAIN`. The skill domain has none: this skill covers it and core. `ncly skill list` lists every skill, and `ncly skill status --json` checks their folders, names, and files
