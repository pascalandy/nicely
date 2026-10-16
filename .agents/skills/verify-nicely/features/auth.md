# auth

`ncly auth` keeps the keys of extensions. `ncly auth status` lists each service that an installed extension declares, and says whether its key comes from its environment variable or from the OS keychain, without printing any key. `ncly auth login` stores a key in the keychain. This feature drives it against the real keychain, with the test service `verifynicely`, whose variable is `VERIFYNICELY_API_KEY`, and a fake key that is no human's key. It never touches the key of a real service.

## Sub-features

- `auth-keychain` reports whether the real keychain answers
- `auth-missing` fails the report with `AUTH_MISSING` for the test service when neither its variable nor the keychain holds its key
- `auth-env` reads the key from the variable, and never prints it
- `auth-login-dry-run` plans the login without asking the keychain anything
- `auth-login` stores the fake key in the real keychain from `--stdin`, and answers without it
- `auth-keychain-source` reads the stored key from the keychain
- `auth-replace` refuses to replace the stored key without `--force`

## How to get to it (user POV)

- `ncly auth status`
- `ncly auth status --json`
- `ncly auth login <service> --stdin`, `--force`, and `--dry-run`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed, on Linux. On macOS, only Pascal runs this feature, because the keychain may ask for his password
- The keychain may hold the fake key of an earlier run under `verifynicely`, because this release has no logout yet
- The fake key of this run: `fake=verify-nicely-fake-$(date +%s)`

- **Fixture.** Run `just verify prepare "$run" auth`. Exit 0, and stderr names the service `verifynicely` and its variable `VERIFYNICELY_API_KEY`.
- **Help.** Run `just verify step "$run" auth auth --help`. Exit 0, and `COMMANDS` lists `login` and `status`.
- **Missing key.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Without the fake key of an earlier run, exit 78, stderr is empty, and stdout is one line with `"ok":false`, `"keychain":"available"`, and the one entry `{"service":"verifynicely","configured":false,"source":"none","code":"AUTH_MISSING", ...}`, whose `hint` is `ncly auth login verifynicely` and whose `next` is `["ncly","auth","login","verifynicely"]`. With it, exit 0 and the entry `{"service":"verifynicely","configured":true,"source":"keychain"}`: report which one you saw.
- **Missing key as text.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status`. The first line is `keychain      The keychain answers.`. Without the fake key of an earlier run, exit 78, and the line of `verifynicely` says `No key for verifynicely in VERIFYNICELY_API_KEY or the keychain.`, with `hint: ncly auth login verifynicely` under it. With it, exit 0, and the line says `The key comes from the keychain.`
- **Key from the variable.** Run `VERIFYNICELY_API_KEY="$fake" just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"env"}]}`.
- **Login dry run.** Run `just verify step "$run" auth auth login verifynicely --dry-run --json`. Exit 0, and stdout holds `"dry_run":true`, the plan `{"key":"verifynicely","action":"login","effects":{"user_writes":true,"network":false,"paid":false}}`, and the pending check `auth.keychain`.
- **Login.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --force --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"action":"login","service":"verifynicely"}`.
- **Key from the keychain.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"keychain"}]}`.
- **Replacement.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --json`. Exit 2, and stderr ends with an answer whose first error is `CONFIRMATION_REQUIRED`, with the hint `ncly auth login verifynicely --stdin --force`.
- **Login as text.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --force`. Exit 0, and stdout is `Stored the key of verifynicely in the keychain.`
- **Proof.** Run `grep -c "$fake" "$run/evidence/auth.txt"` after the steps. It prints `0` and exits 1, grep's code for no match: no transcript holds the fake key.

## Gotchas

- Set `VERIFYNICELY_API_KEY` on the step itself, as written: the wrapper passes the caller's environment to `ncly`, so a value left in the shell changes the result of every later step
- `"keychain":"unavailable"` on Linux means that no Secret Service answers, such as gnome-keyring outside a desktop session: report it, because the steps that expect `available` then fail with `KEYRING_UNAVAILABLE`
- `prepare` refuses to run on macOS without `--mac-keychain`, which only a human passes
- The fake key stays in the keychain under the service `nicely` and the account `verifynicely` after the run, until a release with `ncly auth logout` cleans it up. It is no human's key
