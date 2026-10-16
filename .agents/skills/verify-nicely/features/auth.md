# auth

`ncly auth` keeps the keys of extensions. `ncly auth status` lists each service that an installed extension declares, and says whether its key comes from its environment variable or from the OS keychain, without printing any key. This feature drives it against the real keychain, with the test service `verifynicely`, whose variable is `VERIFYNICELY_API_KEY`, and a fake key that is no human's key. It never touches the key of a real service.

## Sub-features

- `auth-keychain` reports whether the real keychain answers
- `auth-missing` fails the report with `AUTH_MISSING` for the test service when neither its variable nor the keychain holds its key
- `auth-env` reads the key from the variable, and never prints it

## How to get to it (user POV)

- `ncly auth status`
- `ncly auth status --json`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed, on Linux. On macOS, only Pascal runs this feature, because the keychain may ask for his password
- The keychain holds no entry for `verifynicely`, since no command of this release stores a key
- The fake key of this run: `fake=verify-nicely-fake-$(date +%s)`

- **Fixture.** Run `just verify prepare "$run" auth`. Exit 0, and stderr names the service `verifynicely` and its variable `VERIFYNICELY_API_KEY`.
- **Help.** Run `just verify step "$run" auth auth --help`. Exit 0, and `COMMANDS` lists `status`.
- **Missing key.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Exit 78, stderr is empty, and stdout is one line with `"ok":false`, `"keychain":"available"`, and the one entry `{"service":"verifynicely","configured":false,"source":"none","code":"AUTH_MISSING", ...}`, whose `hint` is `ncly auth login verifynicely` and whose `next` is `["ncly","auth","login","verifynicely"]`.
- **Missing key as text.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status`. Exit 78, the first line is `keychain      The keychain answers.`, and the line of `verifynicely` says `No key for verifynicely in VERIFYNICELY_API_KEY or the keychain.`, with `hint: ncly auth login verifynicely` under it.
- **Key from the variable.** Run `VERIFYNICELY_API_KEY="$fake" just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"env"}]}`.
- **Proof.** Run `grep -c "$fake" "$run/evidence/auth.txt"` after the steps. It prints `0` and exits 1, grep's code for no match: no transcript holds the fake key.

## Gotchas

- Set `VERIFYNICELY_API_KEY` on the step itself, as written: the wrapper passes the caller's environment to `ncly`, so a value left in the shell changes the result of every later step
- `"keychain":"unavailable"` on Linux means that no Secret Service answers, such as gnome-keyring outside a desktop session: report it, because the steps that expect `available` then fail with `KEYRING_UNAVAILABLE`
- `prepare` refuses to run on macOS without `--mac-keychain`, which only a human passes
