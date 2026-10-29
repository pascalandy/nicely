# Output

stdout carries data. stderr carries progress, warnings, and errors. Without `--json`, human output uses styles, such as bold headings and colored labels, and `--no-color` turns them off. Deciding whether to style reads only the environment and the stream, and never queries the terminal or tmux. Styles appear only on a terminal, unless `CLICOLOR_FORCE=1` asks for them in a pipe. `NO_COLOR` and `TERM=dumb` turn them off even then.

Styles change only the weight and the color of the text, as [D060](../decisions/D060-style-the-plain-text.md) decides, so the output on a terminal, without its escape codes, equals the output in a pipe. They use the 16 ANSI colors, which the terminal's theme sets, and a word or a mark always carries the meaning that a color shows, such as `fail` or a `●` that a legend explains. `ncly` never wraps a line itself, so a path or a command copies whole, and the terminal wraps a long line. The Markdown that [ncly skill view](../../../extensions/skill/spec.md#ncly-skill) renders in interactive mode is the one exception. A form, a prompt, and a spinner appear only in interactive mode, so no pipe gets them.

With `--json`, every answer is one JSON object on one line, never a bare array. `ok` is `true` or `false` and agrees with the exit code. `contract_version` is the integer version of this public protocol, initially `1`. `--help`, `--version`, and `ncly completion` print text and ignore `--json`.

Parser failures use this envelope too. Machine mode is resolved from `NCLY_JSON` and a valid `--json` flag before reporting an invalid command, unknown flag, or missing value, wherever that flag occurs before `--`. A flag that takes a value takes the next argument even when it starts with `-`, as the parser does, so `ncly --lang --json` sets the language and leaves machine mode off. When a switch repeats, its last valid value counts, so `--json --json=bad` keeps machine mode for the parser failure. Words after `--` are operands, never flags or command names, so `ncly -- help` is `USAGE_INVALID` too. Help retains its documented precedence, including inside a group of shorthands such as `-zh`. Text after `=` is a flag value, so `-z=h` is an unknown flag, not a request for help.

- On success, stdout holds the object. Progress is not shown, and stderr stays empty unless `--verbose` or `NCLY_DEBUG` adds lines.
- On failure, stdout stays empty and the object ends stderr. With `--verbose` or `NCLY_DEBUG`, diagnostic lines come before it.
- When the output cannot be written, such as to a full disk or to a pipe whose reader closed, the command exits 1 with `RUNTIME` on stderr, even for `--help` and `--version`. A lost answer never counts as a success. When stderr refuses the answer too, the exit code stays 1, and the answer is lost.

```json
{"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram","next":["ncly","auth","login","deepgram"]}]}
```

`errors` describes the command's failure. The command computes its overall verdict first, then puts an error with the matching exit code first. When the failures of the command itself map to different exit codes, the most cautious one leads: `130` or `143`, then `1`, `78`, `2`, and `75` only when every failure is temporary. When several failures share that exit code, the first one in this order leads. For exit 1, `RESUME_UNSAFE`, then the answering extension's own codes in the order of its manifest, then `RUNTIME`. For exit 78, `CONFIG_INVALID`, `EXTENSION_INVALID`, `TERMINAL_REQUIRED`, `PREREQ_MISSING`, `CHECK_TIMEOUT`, `CAPABILITY_UNSUPPORTED`, `KEYRING_UNAVAILABLE`, `AUTH_MISSING`, `KEY_NOT_GRANTED`, `AUTH_REJECTED`, then the extension's own codes, so the first fix that a human needs leads. For exit 2, `USAGE_INVALID`, `CONFIRMATION_REQUIRED`, `NOT_FOUND`, then the extension's own codes. Input order or the order in which failures arrive never determines retry safety. Item errors remain inside `results`. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](error-codes.md). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the problem. A command is never translated. |
| `next` | Optional. The command that the hint names, as an array of words, such as `["ncly","skill","list"]`, present only when that command runs as is, with no value left to fill in. Agents run it instead of copying a command out of the hint, as [D053](../decisions/D053-next-command.md) decides. |

A message that refuses a value from the command line, such as an unknown command or path, quotes only its first 80 characters, then `…`. A valid run ID, at most 64 characters, stays whole. The hint and `next` keep the whole value, so that they still run as is.

`next` never changes what the exit code allows. After exit 78, a human runs it. It never names a repeat of a call that the exit code forbids repeating. A `next` command that writes or bills still gets its dry run first.

A failure after partial work keeps the data keys that still apply, such as `output_dir`, next to `errors`. A recorded operation also returns `run_id`, including on failure. Large results are referenced by their artifact paths instead of copied into diagnostics or run listings.

A command that processes several items, such as several URLs or files, checks every input before it starts. A bad input stops the whole command with exit 2. Otherwise the answer holds `results`: one object per item, in input order, each with its own `ok` and either its data or its `errors`. Successful data remains available when another item fails. The command exits 0 when every item succeeded, 75 only when every failure is temporary and the whole invocation meets [Retry safety](retry-safety.md), and 1 otherwise. An incomplete batch uses the top-level `TEMPORARY` or `RUNTIME` code matching that verdict. On cancellation, 130 or 143 takes precedence and `results` contains only attempted items, in input order.

A report command, such as `ncly doctor` or `ncly auth status`, answers with a report. The report goes to stdout even when a check fails, and it lists its findings under its own key, such as `checks`, instead of `errors`. `ok` gives the verdict, and the exit code matches it. When the report command itself fails, it answers with `errors` like any other command.
