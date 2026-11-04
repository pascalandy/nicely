# auth

`ncly auth` keeps the keys of extensions. `ncly auth status` lists each service that an installed extension declares, says whether its key comes from its environment variable or from the OS keychain, without printing any key, and which extensions hold its grant. `ncly auth login` stores a key in the keychain, and `ncly auth logout` removes it. In a terminal, `login` asks for the key in a masked field, and a replacement or a logout asks for a typed `[y/N]` confirmation. `ncly auth grant` lets an extension receive a key once a human confirms it in a terminal, and `ncly auth revoke` takes the grant back. On Linux, this feature uses the real Secret Service in the run's private session.
On macOS, this feature uses the login keychain.
This feature drives the keychain with the test service `verifynicely`, whose variable is `VERIFYNICELY_API_KEY`, and a fake key that is no human's key. Its fixture `verifykeys` declares that service, and `ncly verifykeys digest` answers the SHA-256 of the key that it receives, never the key. It never touches the key of a real service.

## Sub-features

- `auth-keychain` reports whether the real keychain answers
- `auth-grant-dry-run` plans the grant of the test service to the fixture, with the confirmation pending, without a terminal
- `auth-grant-terminal` refuses to grant without a terminal, and names the command that a human runs in one
- `auth-revoke-missing` answers `NOT_FOUND` for a grant that does not exist
- `auth-not-granted` keeps the stored key from the fixture, which answers `KEY_NOT_GRANTED` with the grant command as its `next`. Before the grant, `ncly auth status` and `ncly doctor auth` fail the key with `KEY_NOT_GRANTED` too
- `auth-grant` grants the key to the fixture after a typed `y` on a real terminal, which shows the path of the fixture as a fact. A second grant answers `unchanged`
- `auth-key-reaches` passes the key from the keychain to the fixture once granted, which answers its SHA-256, and stops once revoked
- `auth-grant-status` lists the grant in `ncly auth status`, and passes the key in it and in `ncly doctor auth` once granted
- `auth-missing` fails the report with `AUTH_MISSING` alone for the test service when neither its variable nor the keychain holds its key, even without its grant
- `auth-env` reads the key from the variable, and never prints it
- `auth-login-dry-run` plans the login without asking the keychain anything
- `auth-login` stores the fake key in the real keychain from `--stdin`, and answers without it
- `auth-keychain-source` reads the stored key from the keychain
- `auth-replace` refuses to replace the stored key without `--force`
- `auth-logout-dry-run` plans the logout without asking the keychain anything
- `auth-logout` refuses to remove the key without `--force`, removes it with `--force`, and answers `NOT_FOUND` once it is gone
- `auth-terminal` asks for the key in the masked field on a real terminal, never shows it, and erases the field once answered. A declined replacement keeps the key, Ctrl-C at the field stores nothing, and a confirmed logout removes the key
- `auth-form` asks for a missing service in the form for missing values on a real terminal, then for the key in the masked field, and prints `Next time:` with the service, never the key. Outside a terminal, the call fails with the full command as its hint
- `auth-doctor` gives the test service the verdict of `ncly auth status` in the `auth` component of `ncly doctor`, after `auth.grants`, which finds every grant in use
- `auth-cleanup` logs the fake key out during the cleanup, after a failed run too, and checks that it is gone

## How to get to it (user POV)

- `ncly auth status`
- `ncly auth status --json`
- `ncly auth login <service> --stdin`, `--force`, and `--dry-run`
- `ncly auth logout <service>`, `--force`, and `--dry-run`
- `ncly auth grant <service> <domain>` and `--dry-run`
- `ncly auth revoke <service> <domain>` and `--force`
- `ncly auth grant <service> <domain>` in a terminal, through `just verify tmux`
- `ncly verifykeys digest`, the fixture's command that needs the key
- `ncly auth login <service>` and `ncly auth logout <service>` in a terminal, through `just verify tmux`
- `ncly auth login` without a service, in a terminal and in a pipe
- `ncly doctor auth`, and `ncly doctor` without a component
- `just verify cleanup "$run"`, which logs the fake key out

## Driving it with verify.sh

Follow these guides in order:

1. [Prepare and inspect the key](auth/prepare.md)
2. [Grant and remove the key](auth/grant.md)
3. [Terminal steps and cleanup](auth/terminal.md)

## Gotchas

- Set `VERIFYNICELY_API_KEY` on the step itself, as written: the wrapper passes the caller's environment to `ncly`, so a value left in the shell changes the result of every later step
- Quote each `--answer` as written: the shell would expand the brackets of `[y/N]` otherwise
- `"keychain":"unavailable"` on Linux means the run's private Secret Service failed: stop and report it. Never use the user's session instead
- On macOS, `prepare` needs `--mac-keychain`. Run the steps in the desktop session, because the login keychain refuses an SSH session. [Desktop session](../../../../docs/qa/protocol.md#desktop-session) in docs/qa says how, and when to call a human
- `ncly auth status` and `ncly doctor auth` fail the stored key with `KEY_NOT_GRANTED` until the step Grant in a terminal, and again after the step Revoke: a key that no extension may receive is a problem to fix, and exit 78 is the expected result there
- `just verify step` has no terminal, and `grant` takes no `--force`, so only the `tmux` step grants the key. Never print the fake key to compare digests: pipe it straight into `sha256sum`
- Cleanup removes the fake key after a failed run too. If logout fails, cleanup keeps scratch and exits 1. On macOS, retry when the login keychain answers. On Linux, retry to discard the stopped private keyring. Report a repeated failure
