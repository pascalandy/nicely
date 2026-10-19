# Error codes

Each code maps to exactly one exit code. This table holds core's codes, and a milestone adds core's new codes here. Every extension may answer with these codes too. An extension's own codes start with its domain, such as `TRANSCRIPT_DOWNLOAD_FAILED` or `FLEET_HOST_DOWN`, live in its spec and its [manifest](manifest.md), and map to one exit code each, with the meaning that this file defines.

The set of codes is open: later milestones add core codes, and every extension adds its own. A caller that meets an error code it does not know acts on the exit code, whose meaning never changes, and relays the hint. An unknown warning code is still a warning.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M00 |
| `CONFIRMATION_REQUIRED` | 2 | A step needs a confirmation in non-interactive mode, and `--force` is absent, or a human declined it in interactive mode | M00 |
| `CONFIG_INVALID` | 78 | A config file, or the grants file of [Grants](grants.md), has a syntax error or a value of the wrong type, cannot be read, or holds a `format_version` that this `ncly` does not support | M00 |
| `EXTENSION_INVALID` | 78 | The extension of the named domain cannot run: its manifest or its `en` catalog is missing, unreadable, or invalid, or speaks another format or contract version. In `ncly doctor`, its status command gave an answer that doctor rejects, or failed with an exit code other than 78, as [Status command](status-command.md) defines | M01 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M00 |
| `RUNTIME` | 1 | Any other failure during work | M00 |
| `TEMPORARY` | 75 | A temporary failure for which the whole invocation meets Retry safety | M00 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M00 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M00 |
| `NOT_FOUND` | 2 | The named command, skill, service, profile, component, run, prompt, or recording does not exist | M03 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old, or its version probe fails or prints no version | M05 |
| `CHECK_TIMEOUT` | 78 | A check of `ncly doctor` did not finish within its `--timeout` | M05 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M06 |
| `KEYRING_UNAVAILABLE` | 78 | The OS keychain does not answer when a command needs it: to read a key missing from the environment, for `ncly auth`, or for an `auth.<service>` check of `ncly doctor` | M06 |
| `KEY_NOT_GRANTED` | 78 | An extension needs a key that no human granted to it at its source, as [Grants](grants.md) defines | M08 |
| `CAPABILITY_UNSUPPORTED` | 78 | The selected program cannot enforce a required mode or protocol | M09 |
| `RESUME_UNSAFE` | 1 | A run cannot safely continue because evidence is missing, changed, uncertain, or unsupported | M10 |
| `AUTH_REJECTED` | 78 | A service refused a key, such as Deepgram answering 401, or the key holds a character that no request can carry. A human replaces the key | M13 |

Warnings use codes from this table, and an extension's own warnings start with its domain, such as `SKILL_SHADOWED`.

| Code | When | Since |
|---|---|---|
| `CONFIG_UNKNOWN_KEY` | A config file holds a key that this version of `ncly` does not know | M00 |
| `CONFIG_DEFAULTS_USED` | Discovery used defaults because a config file could not be read | M03 |
| `KEYRING_ENV_ONLY` | The OS keychain does not answer, so keys come only from the environment. The `auth.keychain` check of `ncly doctor` warns with it | M06 |
| `GRANT_UNUSED` | A grant names a service that the extension of its domain no longer declares, or a source that no longer holds that extension. The `auth.grants` check of `ncly doctor` warns with it | M08 |
