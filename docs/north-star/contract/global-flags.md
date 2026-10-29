# Global flags

Each global flag other than `--help` and `--version` has an environment variable with the same effect. A flag wins over its variable, so `ncly` checks the variable only when the flag is absent. An `NCLY_` variable of a switch turns it on when it equals `1`, and leaves it off when it is unset, empty, or `0`. Any other value is `USAGE_INVALID` with exit 2, and a refused `NCLY_JSON` leaves machine output off, as [D066](../decisions/D066-refuse-unreadable-variables.md) decides. [Configuration](configuration.md) defines which languages `--lang` and `NCLY_LANG` refuse. `--help`, `--version`, and the completion requests of the shell scripts still work with a refused value, as they do with an invalid flag value. `NO_COLOR` turns color off when it is set and not empty, as [no-color.org](https://no-color.org/) defines it. [Programs that ncly runs](programs-that-ncly-runs.md) receive the resolved values.

| Flag | Variable | Default | Effect |
|---|---|---|---|
| `-h`, `--help` | | | Shows help on stdout and exits 0. It wins over every other argument, including unknown flags, but still honors `--lang` and `--no-color`. |
| `--version` | | | Prints `ncly vX.Y.Z` on stdout and exits 0. It works after any command and wins over every other argument except `--help`. |
| `--json` | `NCLY_JSON=1` | off | Machine output, as described in [Output](output.md). |
| `--no-input` | `NCLY_NO_INPUT=1` | off | Never prompts, even in a terminal. |
| `--no-color` | `NO_COLOR`, or `TERM=dumb` | off | Plain text. |
| `--lang <tag>` | `NCLY_LANG` | detected | Language of human text, such as `en` or `fr-CA`. |
| `-v`, `--verbose` | `NCLY_VERBOSE=1` | off | Adds step details on stderr. |

`NCLY_DEBUG=1` adds internals, timings, and child commands on stderr. It has no flag. Keys never appear in any output.

Every command that writes accepts `--dry-run`. The other flags below apply when the command confirms an action or produces files.

| Flag | Effect |
|---|---|
| `-n`, `--dry-run` | Uses the same preparation as execution and shows the planned effects and checks still pending. It changes no user file, config, key, grant, or persistent run record. It reads no key and sends no paid request. It may fill only Nicely's cache, including Python dependencies. |
| `-f`, `--force` | Answers yes to every confirmation, such as an overwrite or a deletion. It never grants a key, because a grant needs a human in a terminal. |
| `-o`, `--output <path>` | Writes the result to this file, or to this folder when the command writes several files. |
| `--run-id <id>` | Names the run of a recorded command, so that a repeated call continues that run, as [Run IDs](run-ids.md) defines. |

Commands that call the network or run for minutes also accept `--timeout <duration>`. A duration is `30s`, `5m`, `2h`, or a number of seconds. Each command states its default in its help.
