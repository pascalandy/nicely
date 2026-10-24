# Grants

A key reaches an extension only after a human grants that key to that extension in a terminal, through [ncly auth](../core-spec.md#ncly-auth) `grant`, as [D030](../decisions/D030-trusted-extensions.md) and [D058](../decisions/D058-grants-by-path.md) decide. `--force` never grants a key. Grants are consent and protection against accidental leaks, not isolation.

A grant binds a service to an extension's slug and source. The source of an external extension is the absolute path where [Discovery](discovery.md) found its executable, without resolving a link, as the `path` of [ncly doctor](../core-spec.md#ncly-doctor) reports it: `null` for a bundled extension. An extension from another source inherits no grant through its name, such as another `ncly-transcript` earlier in `PATH`. A new version at the same path keeps its grants, and a key that it declares for the first time reaches it only after a human grants it.

Grants stay on one machine, in `grants.json` in the state folder, such as `~/.local/state/nicely/grants.json`. The file is one JSON object with `format_version`, `1`, and `grants`, an array.
Each grant holds `service`, `domain`, `path`, and `granted_at`, an ISO 8601 time.
The file holds one grant per service and slug.
The existing `domain` field holds that slug, as [Extension identity](extension-identity.md) defines. Core writes it as [Record format](record-format.md) writes a record, under the lock `grants.lock` beside it, and revalidates it under that lock. A missing file holds no grant. A file that core cannot read, or that holds another `format_version`, grants nothing, and `ncly auth grant`, `revoke`, and `status` fail with `CONFIG_INVALID`, whose hint names the file.

When core starts a command of an extension, it passes the key of each service that the command's `requires.keys` declares and that a human granted to this extension at its source, from the service's variable or, when the variable is empty, from the keychain, as [Keys](keys.md) reads it. It passes no other key. `NCLY_MISSING_KEYS` names each declared service whose key it does not pass, as `<service>=<code>`, separated by commas, in the order of `requires.keys`, such as `deepgram=KEY_NOT_GRANTED`:

- `KEY_NOT_GRANTED` when no human granted the key to this extension at its source
- `KEYRING_UNAVAILABLE` when the variable is empty and the keychain does not answer
- `AUTH_MISSING` when neither the variable nor the keychain holds the key

Core leaves `NCLY_MISSING_KEYS` out when it passes every declared key. A dry run, and a [status command](status-command.md) without `--live=true`, receive no key, because neither sends a request: core reads no key for them, and `NCLY_MISSING_KEYS` names only the services that no human granted, so a dry run reports a missing grant without a keychain read.

Core never refuses to start an extension over a key, and never asks for a grant when it starts one, because only the extension knows whether this invocation needs the key: returning a verified saved transcript needs no new Deepgram request. For a key it needs, the extension answers the code that `NCLY_MISSING_KEYS` names.
That code's hint and `next` follow these rules:

- `KEY_NOT_GRANTED` names `ncly auth grant <service> <slug>`
- `AUTH_MISSING` names `ncly auth login <service>`
- `KEYRING_UNAVAILABLE` names the variable to export and has no `next`, as [Keys](keys.md) defines

 A Go extension gets this problem from the SDK, as [SDK](sdk.md) defines.

An optional key, which a manifest marks in its `keys` table, as [Manifest](manifest.md) defines, follows the same rules. The command runs without it, such as a harness that uses its own login, so a missing grant, a missing key, and an unavailable keychain never block that fallback, and [ncly auth](../core-spec.md#ncly-auth) `status` and [ncly doctor](../core-spec.md#ncly-doctor) report it without failing.

In a nested call, such as an extension that runs another `ncly` command, the inner `ncly` applies these rules to the inner extension, with the variables that reached it and the keychain, and sets `NCLY_MISSING_KEYS` anew, so the outer extension's value never reaches the inner one. A key that exists only in the outer shell's environment does not cross an extension that was not granted it.
