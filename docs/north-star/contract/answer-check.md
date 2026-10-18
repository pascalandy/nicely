# Answer check

Core checks what an extension answers, as [D050](../decisions/D050-answer-check.md) decides. Without `--json`, the extension writes to the stdout and stderr of `ncly` directly, so a human sees its text, styles, and forms unchanged, and core checks only its exit code. A command whose result kind is `text` ignores `--json`, as [Output](output.md) allows, so core treats it the same way under `--json`.

With `--json`, core holds the extension's stdout until the extension exits, so a broken answer never reaches the caller. It reads stderr line by line and holds the latest line: when a new line arrives, it forwards the held one. When the answer belongs on stdout, core forwards the held line too. Otherwise the held line is the answer, which core checks. A last line without a newline counts as a line. An extension follows [Output](output.md), so it writes nothing else to stderr without `--verbose` or `NCLY_DEBUG`, but core forwards what it writes anyway, so a human still sees it.

Core accepts the answer only when every rule below holds:

1. The exit code is one of [Exit codes](exit-codes.md): 0, 1, 2, 75, 78, 130, or 143
2. The answer is one line that holds one JSON object, with `ok`, a boolean that is `true` exactly when the exit code is 0, and `contract_version`, the integer `1`, in exactly one place. A success puts it on stdout. A failure puts it on the last line of stderr, with stdout empty. A report puts its findings on stdout with exit 0 or 78, as [Output](output.md) says, and a report command that fails itself answers as any failure, on stderr
3. An answer on stderr has a non-empty `errors`, and an answer on stdout has none
4. `errors` and `warnings`, when present, are arrays of objects that each hold the strings `code`, `message`, and `hint`, and, when present, `next`, a non-empty array of strings. Each code is uppercase letters, digits, and underscores
5. `results`, when present, is an array of objects that each hold the boolean `ok`, which is `false` exactly when the item has a non-empty `errors`. A success has no failed item
6. An error code that starts with the extension's domain and an underscore is declared in its manifest. A warning code is checked for its form only
7. When core knows the leading error code, as its own or as a declared one, it maps to the exit code. No top-level error code that core knows maps to a more cautious exit code, in the order of [Output](output.md), whatever the codes that core does not know
8. When an item failed, the exit code is 1 led by `RUNTIME`, 75 led by `TEMPORARY`, 130 led by `INTERRUPTED`, or 143 led by `TERMINATED`, as [Output](output.md) defines for a batch. Exit 75 holds only when every error, at the top and in every item, is `TEMPORARY`

An extension that a signal ended counts as exiting with 128 plus the signal's number, as a shell reports it. Core passes a valid answer on unchanged. Anything else is a protocol failure, and so is an extension that core cannot start, such as a script whose interpreter is missing: core drops the output that it holds and answers `RUNTIME` with exit 1, because the extension may already have produced effects, and the message names the rule that broke. Without `--json`, an exit code outside the list is a protocol failure too, and core prints its error block after the extension's text.

When `ncly` receives Ctrl-C or SIGTERM before the extension exits, the first signal decides the verdict: core answers 130 with `INTERRUPTED` or 143 with `TERMINATED`, whatever the extension's exit code. A valid answer keeps its data keys, its `results`, and its other errors after the leading one. Without a valid answer, core answers alone, with no data key. Without `--json`, core prints its own error block only when the extension's exit code differs from the signal's. A signal that arrives after the extension exited changes nothing: core keeps its handler registered and publishes the answer that it has decided.
