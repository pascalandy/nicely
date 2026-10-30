# Valid manifest

A manifest is invalid unless every rule below holds. Core rejects it before work, as [Compatibility](compatibility.md) asks of a format version that it does not support:

- `format_version` and `contract_version` are both `1`. Identity and command ownership follow [Extension identity](extension-identity.md)
- Each word of a path is lowercase letters, digits, and hyphens, starting with a letter, and no path appears twice. Each parent down to the extension root is declared, so that help lists each one with its summary. A path with children and no arguments groups them: alone, it prints its help, as [Usage](usage.md) says
- A command's `result.kind` is one of the four kinds of the [table](manifest.md)
- A flag's name, an argument's name, and a value's name follow the rule of a path word. A shorthand is one ASCII letter. No flag takes the name or the shorthand of a global flag, `--dry-run`, or `--run-id`, and no name or shorthand appears twice in one command. Core checks every declaration before it builds a command from it, so a broken one never reaches the parser
- A type is one of the four of the [table](manifest.md), and a switch's is `boolean`. A switch is neither required nor variadic. No required argument follows an optional one, and only the last argument may be variadic. A `default` has the flag's type, an integer for `integer` and a finite number for `number`, a switch's is `false`, and a required or variadic flag has none. A `values` holds at least one value, and each one is nonempty, given once, and read by its type. A switch has none
- Each flag group names at least two of the command's own flags, each once, with `0 <= min <= max <=` the number of its flags, and `max` at least 1
- A `values_from` sits on an argument or on a flag that takes a value, and its `command`, `key`, and `field` are not empty. When `command` is at or beneath this extension root, it names a qualifying command of this manifest. [Completion values](completion-values.md) defines qualification. A sibling extension in the same domain is another provider
- Each declared error or warning uses the uppercase prefix of [Extension identity](extension-identity.md), such as `HELLO_GRUMPY` or `VIDEO_CONVERT_FAILED`. The code contains only uppercase letters, digits, and underscores. It differs from every core code and appears once across errors and warnings
- Each declared error maps to `1`, `2`, or `78`. A warning has no `exit`. Exit 75 belongs to `TEMPORARY` alone, because [Retry safety](retry-safety.md) judges the whole invocation. Exits 130 and 143 belong to signals
- Each explicit code scope contains at least one entry, with no duplicates. Each `commands` entry names a declared command that runs. Each `modes` entry is `run` or `dry_run`. An explicit `dry_run` requires every selected command to offer it
- A declared `untested` contains a nonempty reason and an HTTP or HTTPS issue URL whose path ends with `/issues/<positive integer>`
- A command that sets `doctor` is a status command that qualifies, as [Status command](status-command.md) defines
- Each key of `programs` is a file name, without `/`, and at least one command names it. Each name in the `requires.bins` of a command is a key of `programs`, and appears once there
- Each name in the `requires.keys` of a command is a service name, as [Keys](keys.md) defines, and appears once there. Each key of `keys` is a service that the `requires.keys` of at least one command names, and its `optional` is a boolean
- A schema is valid JSON Schema draft 2020-12, declares no other `$schema`, and is a table whose `type` is `object`, and whose top level, once its references are resolved, holds only `type`, `properties`, `required`, `title`, `description`, and `$schema`, which describe the answer's keys. The item of a batch's `results`, when declared, follows the same rule. Each `$ref` in it is local and resolves within it, without a cycle, and it holds no `$id`, `$anchor`, `$dynamicRef`, or `$dynamicAnchor`. A `text` result has no schema, and only a command whose `modes` set `dry_run` has `dry_run_schema`
- Each declared `config` key starts with the slug and appears once. It holds at least two words joined by dots. Each word starts with a lowercase letter and contains only lowercase letters, digits, hyphens, or underscores
- A `version` is one or more numbers joined by dots. A `probe` is a non-empty array of strings, and so is each command of `install`
- The `en` catalog holds every catalog ID that the manifest names

An invalid code scope identifies its TOML line and column. Positions come from parsed TOML, including quoted keys and inline tables.
