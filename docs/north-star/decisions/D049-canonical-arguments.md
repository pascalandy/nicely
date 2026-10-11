# D049 Pass an extension its validated call in one canonical form

Decided 2026-10-10. Once core has validated a call from the manifest, it starts the extension with the words of the command path after the domain, each flag that the caller set as one `--<name>=<value>` word, then `--` and the positional arguments. Global flags reach the extension only as variables. [Dispatch](../contract.md#dispatch) holds the rules.

**Why.** Core has already parsed the call, including a group of shorthands that mixes the extension's flags with global ones, such as `-vg Hi`, and the precedence rules of [Output](../contract.md#output). One form spares every extension a second parser of those rules: a shell extension reads its call with a `case` statement, and no global flag can pass for one of its own. git passes `git-<name>` the words after `git <name>` untouched, because git validates nothing first. Nicely validates first, so it can pass the simpler form.

**Rejected.** Passing the caller's words unchanged, because every extension would then need the global flags and the parser rules to read them. Passing the call as a JSON document on stdin, because stdin carries the extension's input, such as the material of `ncly agent run`.
