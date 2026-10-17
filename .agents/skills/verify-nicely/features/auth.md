# auth

`ncly auth` keeps the keys of extensions. `ncly auth status` lists each service that an installed extension declares, and says whether its key comes from its environment variable or from the OS keychain, without printing any key. `ncly auth login` stores a key in the keychain, and `ncly auth logout` removes it. In a terminal, `login` asks for the key in a masked field, and a replacement or a logout asks for a typed `[y/N]` confirmation. `ncly auth grant` lets an extension receive a key once a human confirms it in a terminal, and `ncly auth revoke` takes the grant back. This feature drives it against the real keychain, with the test service `verifynicely`, whose variable is `VERIFYNICELY_API_KEY`, and a fake key that is no human's key. It never touches the key of a real service.

## Sub-features

- `auth-keychain` reports whether the real keychain answers
- `auth-grant-dry-run` plans the grant of the test service to the fixture, with the confirmation pending, without a terminal
- `auth-grant-terminal` refuses to grant without a terminal, and names the command that a human runs in one
- `auth-revoke-missing` answers `NOT_FOUND` for a grant that does not exist
- `auth-missing` fails the report with `AUTH_MISSING` for the test service when neither its variable nor the keychain holds its key
- `auth-env` reads the key from the variable, and never prints it
- `auth-login-dry-run` plans the login without asking the keychain anything
- `auth-login` stores the fake key in the real keychain from `--stdin`, and answers without it
- `auth-keychain-source` reads the stored key from the keychain
- `auth-replace` refuses to replace the stored key without `--force`
- `auth-logout-dry-run` plans the logout without asking the keychain anything
- `auth-logout` refuses to remove the key without `--force`, removes it with `--force`, and answers `NOT_FOUND` once it is gone
- `auth-terminal` asks for the key in the masked field on a real terminal, never shows it, and erases the field once answered. A declined replacement keeps the key, Ctrl-C at the field stores nothing, and a confirmed logout removes the key
- `auth-form` asks for a missing service in the form for missing values on a real terminal, then for the key in the masked field, and prints `Next time:` with the service, never the key. Outside a terminal, the call fails with the full command as its hint
- `auth-doctor` gives the test service the verdict of `ncly auth status` in the `auth` component of `ncly doctor`
- `auth-cleanup` logs the fake key out during the cleanup, after a failed run too, and checks that it is gone

## How to get to it (user POV)

- `ncly auth status`
- `ncly auth status --json`
- `ncly auth login <service> --stdin`, `--force`, and `--dry-run`
- `ncly auth logout <service>`, `--force`, and `--dry-run`
- `ncly auth grant <service> <domain>` and `--dry-run`
- `ncly auth revoke <service> <domain>` and `--force`
- `ncly auth login <service>` and `ncly auth logout <service>` in a terminal, through `just verify tmux`
- `ncly auth login` without a service, in a terminal and in a pipe
- `ncly doctor auth`, and `ncly doctor` without a component
- `just verify cleanup "$run"`, which logs the fake key out

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed, on Linux. On macOS, only Pascal runs this feature, because the keychain may ask for his password
- `tmux -V` answers, for the steps in a terminal, as [terminal.md](terminal.md) describes `just verify tmux`
- The keychain may hold the fake key of an earlier run under `verifynicely`, left by a release without logout or by a cleanup that could not finish
- The fake key of this run: `fake=verify-nicely-fake-$(date +%s)`

- **Fixture.** Run `just verify prepare "$run" auth`. Exit 0, and stderr names the service `verifynicely` and its variable `VERIFYNICELY_API_KEY`.
- **Help.** Run `just verify step "$run" auth auth --help`. Exit 0, and `COMMANDS` lists `grant`, `login`, `logout`, `revoke`, and `status`.
- **Grant dry run.** Run `just verify step "$run" auth auth grant verifynicely verifykeys --dry-run --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"dry_run":true,"pending_checks":[{"id":"auth.confirmation","message":"Whether a human at a terminal confirms the grant of the key of verifynicely to verifykeys."}],"plan":[{"key":"verifynicely","action":"grant","domain":"verifykeys","effects":{"user_writes":true,"network":false,"paid":false}}]}`.
- **Grant without a terminal.** Run `just verify step "$run" auth auth grant verifynicely verifykeys --json`. Exit 78, stdout is empty, and stderr ends with an answer whose first error is `TERMINAL_REQUIRED`, with the hint ``Run `ncly auth grant verifynicely verifykeys` in a terminal`` and the `next` `["ncly","auth","grant","verifynicely","verifykeys"]`.
- **No grant to revoke.** Run `just verify step "$run" auth auth revoke verifynicely verifykeys --force --json`. Exit 2, and stderr ends with an answer whose first error is `NOT_FOUND`, with the hint `ncly auth status` and the same command in `next`.
- **Missing key.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Without the fake key of an earlier run, exit 78, stderr is empty, and stdout is one line with `"ok":false`, `"keychain":"available"`, and the one entry `{"service":"verifynicely","configured":false,"source":"none","code":"AUTH_MISSING", ...}`, whose `hint` is `ncly auth login verifynicely` and whose `next` is `["ncly","auth","login","verifynicely"]`. With it, exit 0 and the entry `{"service":"verifynicely","configured":true,"source":"keychain"}`: report which one you saw.
- **Missing key as text.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status`. The first line is `keychain      available`, then an empty line. Without the fake key of an earlier run, exit 78, and the line of `verifynicely` is `verifynicely  none`. Two lines follow it, in the column of `none`: `fail[AUTH_MISSING] No key for verifynicely in VERIFYNICELY_API_KEY or the keychain.`, then `hint: ncly auth login verifynicely`. With the key, exit 0, and the line is `verifynicely  keychain`.
- **Key from the variable.** Run `VERIFYNICELY_API_KEY="$fake" just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"env"}]}`.
- **Login dry run.** Run `just verify step "$run" auth auth login verifynicely --dry-run --json`. Exit 0, and stdout holds `"dry_run":true`, the plan `{"key":"verifynicely","action":"login","effects":{"user_writes":true,"network":false,"paid":false}}`, and the pending check `auth.keychain`.
- **Login.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --force --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"action":"login","service":"verifynicely"}`.
- **Key from the keychain.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"keychain"}]}`.
- **Doctor with the key.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth doctor auth --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"checks":[{"id":"auth.keychain","component":"auth","status":"pass","message":"The keychain answers."},{"id":"auth.verifynicely","component":"auth","status":"pass","message":"The key comes from the keychain."}]}`.
- **Replacement.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --json`. Exit 2, and stderr ends with an answer whose first error is `CONFIRMATION_REQUIRED`, with the hint `ncly auth login verifynicely --stdin --force`.
- **Login as text.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --force`. Exit 0, and stdout is `Stored the key of verifynicely in the keychain.`
- **Logout dry run.** Run `just verify step "$run" auth auth logout verifynicely --dry-run --json`. Exit 0, and stdout holds `"dry_run":true`, the plan `{"key":"verifynicely","action":"logout","effects":{"user_writes":true,"network":false,"paid":false}}`, and the pending check `auth.keychain`.
- **Logout without --force.** Run `just verify step "$run" auth auth logout verifynicely --json`. Exit 2, and stderr ends with an answer whose first error is `CONFIRMATION_REQUIRED`, with the hint `ncly auth logout verifynicely --force` and the same command in `next`.
- **Logout.** Run `just verify step "$run" auth auth logout verifynicely --force --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"action":"logout","service":"verifynicely"}`.
- **Key gone.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Exit 78, and the entry of `verifynicely` has `"source":"none"` and `"code":"AUTH_MISSING"`.
- **Login in a terminal.** Run `just verify tmux "$run" auth --answer 'verifynicely key' "$fake" auth login verifynicely`. Exit 0. Before Enter, the terminal shows the masked field: `┃ verifynicely key`, `┃ It goes to the keychain and never shows.`, `┃ > ` followed by one `*` per character of the fake key, and `enter submit`. After Enter, the field is gone, and the terminal shows only `Stored the key of verifynicely in the keychain.`, with no `Next time:` line.
- **Key from the terminal.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth auth status --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"verifynicely","configured":true,"source":"keychain"}]}`.
- **Replacement in a terminal.** Run `just verify tmux "$run" auth --answer '[y/N]' n auth login verifynicely`. Exit 2. The terminal shows `? The keychain already holds a key for verifynicely. Replace it? [y/N] n`, then `error[CONFIRMATION_REQUIRED]: The key of verifynicely stays as it was.`, and no masked field.
- **Ctrl-C at the field.** Run `just verify tmux "$run" auth --answer 'verifynicely key' '^C' auth login verifynicely --force`. Exit 130. The terminal shows only `error[INTERRUPTED]: Stopped by Ctrl-C. Finished work is kept.` and its hint, because the field erased itself.
- **Logout in a terminal.** Run `just verify tmux "$run" auth --answer '[y/N]' y auth logout verifynicely`. Exit 0. The terminal shows `? Remove the key of verifynicely from the keychain? [y/N] y`, then `Removed the key of verifynicely from the keychain.`
- **Doctor without the key.** Run `VERIFYNICELY_API_KEY= just verify step "$run" auth doctor --json`. Exit 78, stderr is empty, and stdout lists `core.config`, then `auth.keychain` as a `pass`, then `{"id":"auth.verifynicely","component":"auth","status":"fail","code":"AUTH_MISSING", ...}` with the hint and `next` of the status entry, then the `extensions` checks, each a `pass`.
- **Nothing to remove.** Run `just verify step "$run" auth auth logout verifynicely --force --json`. Exit 2, and stderr ends with an answer whose first error is `NOT_FOUND`, with the hint and `next` `ncly auth status`.
- **Missing service.** Run `just verify step "$run" auth auth login`. Exit 2, stdout is empty, and stderr starts with `error[USAGE_INVALID]: The argument <service> is missing.`, with the hint `ncly auth login <service>`.
- **Login from the form.** Run `just verify tmux "$run" auth --answer 'Service of the key' verifynicely --answer 'verifynicely key' "$fake" auth login`. Exit 0. Before the first Enter, the terminal shows the form for missing values: `┃ service`, `┃ Service of the key, such as deepgram`, and `┃ > verifynicely`. Before the second, it shows the masked field of the step Login in a terminal. After it, both are gone, and the terminal shows `Stored the key of verifynicely in the keychain.`, then `Next time:` and `  ncly auth login verifynicely`, never the key.
- **Login for the cleanup.** Run `printf '%s\n' "$fake" | just verify step "$run" auth auth login verifynicely --stdin --force --json`. Exit 0, so the cleanup has a key to log out.
- **Proof.** Run `grep -c "$fake" "$run/evidence/auth.txt"` after the steps. It prints `0` and exits 1, grep's code for no match: no transcript holds the fake key, not even the one that the terminal steps typed.
- **Cleanup.** Run `just verify cleanup "$run"`, as SKILL.md says. Exit 0, stderr says `logged the fake key of verifynicely out of the keychain`, and `$run/evidence/cleanup.txt` holds an `ncly auth logout verifynicely --force --json` that exits 0, then an `ncly auth status --json` with `"keychain":"available"`, whose entry of `verifynicely` has `"source":"none"` and `"code":"AUTH_MISSING"`.

## Gotchas

- Set `VERIFYNICELY_API_KEY` on the step itself, as written: the wrapper passes the caller's environment to `ncly`, so a value left in the shell changes the result of every later step
- Quote each `--answer` as written: the shell would expand the brackets of `[y/N]` otherwise
- `"keychain":"unavailable"` on Linux means that no Secret Service answers, such as gnome-keyring outside a desktop session: report it, because the steps that expect `available` then fail with `KEYRING_UNAVAILABLE`
- `prepare` refuses to run on macOS without `--mac-keychain`, which only a human passes
- The cleanup logs the fake key out after a failed run too. When it cannot, because the keychain does not answer, it keeps the scratch folder and exits 1: run `just verify cleanup "$run"` again once the keychain answers, and report it if it still fails
