# Completion values

Completion reads every manifest, as [Discovery](discovery.md) says, so it offers an extension's commands and flags without starting it. A value that only an extension knows, such as a skill name, comes from the list command that the argument's or the flag's `values_from` names, as [D055](../decisions/D055-completion-values-from-a-list-command.md) decides: a list command of the same extension or of another installed one, such as `skill list` for `skill view`'s name argument.

That command qualifies only when every condition below holds:

- It belongs to an installed extension, never to core
- Its verb is `list`
- It takes no argument
- It is not a group
- Its `result.kind` is `object`
- It declares `effects`, as [ncly describe](../core-spec.md#ncly-describe) defines them, with `user_writes` and `paid` both false. A command that omits `effects` never qualifies

Core checks these conditions on the command's own declaration each time it completes, and [Manifest](manifest.md) applies the same check to a command of the manifest that names it, so its author finds out before any Tab.

Core calls that command as [Commands that core calls](commands-that-core-calls.md) defines, so the program that starts is the one whose declaration core checked. Core forwards none of its output. Core reads no key for it, as for a dry run, so a Tab never waits on the keychain: `NCLY_MISSING_KEYS` names only the services that no human granted, as [Grants](grants.md) defines.

The limit of the call is 2 seconds, which bound the reading of its output too. The command therefore gets no drain limit, even when it exits within them, because a Tab must not wait. Ctrl-C or SIGTERM during completion ends the command and its descendants as the limit does, at once. `ncly` then exits with 130 or 143.

When the command answers within the limit, core suggests the `field` of each object in the array under `key`.

Any failure, such as a list command that is not installed or does not qualify, an error answer, a timeout, or a missing key or field, suggests nothing and prints nothing, so a Tab never shows an error. Completion writes the reason to Cobra's completion debug log, which Cobra writes only when `BASH_COMP_DEBUG_FILE` names a file. `values_from` only suggests: the command that declares it still accepts a value that completion did not offer, and answers an unknown one itself, as `ncly skill view` answers `NOT_FOUND`.
