# Status command

An extension adds checks to [ncly doctor](../core-spec.md#ncly-doctor) through its status command, as [D056](../decisions/D056-extension-status-command.md) decides.
Its path follows [Extension identity](extension-identity.md), and its manifest sets `doctor = true`. Doctor runs it and lists its checks under the extension's component. An agent may also call it alone, as any other command.

A command may set `doctor` only when every condition below holds, and [Manifest](manifest.md) rejects a manifest otherwise:

- Its path is the extension root followed by `status`, such as `transcript status` or `video convert status`, and it is not a group
- Its `result.kind` is `report`
- It takes no argument and declares no flag group. Its only own flags are `live` and `timeout`, and it may declare either one or neither. Neither is required or has `values`, `live` is a switch whose default is false, and `timeout` takes a `string`, so each call that doctor builds passes validation
- It declares `effects`, as [ncly describe](../core-spec.md#ncly-describe) defines them, with `user_writes` and `paid` both false, and its `modes` sets none of `dry_run`, `recorded`, and `resume`. When its `effects` set `network`, it declares `live`

Doctor calls the status command as [Commands that core calls](commands-that-core-calls.md) defines. It passes `--live=true` when its own caller set `--live`, and `--timeout=<n>s`, where `n` is the number of seconds left of its own `--timeout`, rounded up, each only when the command declares that flag. The limit of the call is the end of that `--timeout`, as [ncly doctor](../core-spec.md#ncly-doctor) defines.

A status command uses the network only under `--live=true`, so without it, the command stays on the machine, as doctor does by default under [D011](../decisions/D011-local-checks.md). It checks the conditional programs that its config selects, as [Required programs](../core-spec.md#required-programs) checks the others: a missing or too old one fails with `PREREQ_MISSING`, whose hint and `next` hold its install command. A Go extension calls the program check of `sdk/runner` that doctor calls, with the program's entry in its own manifest, so it never copies the lookup, the probe, the version rule, or the choice of install command.

The status command answers as a report command, as [Output](output.md) defines: its findings under `checks`, on stdout, with exit 0 when no check fails and 78 when one fails. Each check is an object of [ncly doctor](../core-spec.md#ncly-doctor), whose `component` is the slug and whose `id` starts with the slug and a dot. A failed check's `code` maps to exit 78, from core or the manifest.
A warning uses a core warning code or the own-code prefix of [Extension identity](extension-identity.md). Doctor reads only `checks`.

Doctor checks the whole answer against these rules: every entry of `checks`, and the agreement of `ok` and the exit code with them. A single invalid entry rejects the whole answer, so doctor never drops an entry, and a report never turns green by accident. For a rejected answer, failed command, or unanswered call, doctor lists one failed check.
Its `id` is `<slug>.status`, with this code:

- `CHECK_TIMEOUT`, when the command has not answered at the limit
- The leading error when a valid failure answer maps to exit 78. For example, `CONFIG_INVALID` reports a broken config table for the slug
- `EXTENSION_INVALID` for an answer that doctor rejects, such as one that breaks the Answer check or holds one entry that breaks the rules of this section, and for a failure with exit 1, 2, or 75. Its message names the rule that broke

Doctor starts nothing for an extension whose manifest is invalid, because it cannot trust its declarations. The `extensions` component lists that manifest, and `ncly doctor <slug>` answers `EXTENSION_INVALID` with exit 78, as a call under its extension root does.
