# Manifest

The manifest is a TOML file that declares the extension's commands, in the same declarations as core's commands, as [Command descriptions](command-descriptions.md) defines, and its own error codes. Its text fields hold catalog IDs, never text.

```toml
format_version = 1
contract_version = 1
domain = "hello"

[[commands]]
path = "hello"
summary = "hello.summary"

[[commands]]
path = "hello run"
summary = "hello.run.summary"
examples = ["ncly hello run Ada", "ncly hello run --greeting Hi Ada --json"]

[[commands.args]]
name = "name"
summary = "hello.run.name"

[[commands.flags]]
name = "greeting"
shorthand = "g"
value = "text"
summary = "hello.run.greeting"

[[commands]]
path = "hello check"
summary = "hello.check.summary"
result = { kind = "report" }

[[errors]]
code = "HELLO_GRUMPY"
exit = 1
```

| Key | Type | Meaning |
|---|---|---|
| `format_version` | integer | The manifest format, `1` |
| `contract_version` | integer | The contract version that the extension speaks, `1`, as `NCLY_CONTRACT_VERSION` carries it |
| `domain` | string | The domain, equal to the name of the executable after `ncly-` |
| `commands` | array | One declaration per command. Help lists them in alphabetical order, as it lists core's |
| `commands[].path` | string | The words after `ncly`, separated by one space, such as `hello run`. The first word is the domain, and the command whose path is the domain alone gives the domain's line in `ncly --help` |
| `commands[].summary` | string | The catalog ID of its one-line summary |
| `commands[].examples` | array | Full command lines, as strings that are never translated |
| `commands[].args` | array | The positional arguments in order, each with `name` and `summary`, a catalog ID |
| `commands[].flags` | array | The command's own flags, each with `name`, the long name without `--`, `summary`, a catalog ID, and the optional `shorthand`, one ASCII letter, and `value`, the name of the value, such as `text`. A flag without `value` is a switch |
| `commands[].args[].type`, `commands[].flags[].type` | string | Optional, from M03. The type of the value: `string`, `boolean`, `integer`, or `number`. An argument and a flag that takes a value default to `string`, and a switch is a `boolean` |
| `commands[].args[].required`, `commands[].flags[].required` | boolean | Optional, from M03. `true` when every call gives the value. Validation answers `USAGE_INVALID` before the command runs when a call lacks it |
| `commands[].args[].variadic`, `commands[].flags[].variadic` | boolean | Optional, from M03. `true` when it takes several values: the last argument takes every remaining word, and a flag takes one value each time it appears, which [Dispatch](dispatch.md) passes in the caller's order |
| `commands[].args[].values`, `commands[].flags[].values` | array | Optional, from M07. The only values that it accepts, as strings that its type reads, such as `["small", "large"]`. Validation answers `USAGE_INVALID` for any other value, and the form for missing values of [Modes](modes.md) offers them as a choice |
| `commands[].flags[].default` | typed value | Optional, from M03. The default of a flag that takes a value, of the flag's type, which [ncly describe](../core-spec.md#ncly-describe) reports. Core passes only the flags that the caller set, so the command applies its own default. A switch defaults to `false` |
| `commands[].flag_groups` | array | Optional, from M03. Each group has `flags`, the long names of some of the command's own flags, and the integers `min`, `0` when absent, and `max`, the number of its flags when absent: how many of them a call may set. Validation answers `USAGE_INVALID` before the command runs when a call sets fewer or more |
| `commands[].effects` | table | Optional, from M03. The booleans `user_writes`, `network`, and `paid`, each `false` when absent. A command that omits the table permits no assumption, so [ncly describe](../core-spec.md#ncly-describe) reports each effect `true` |
| `commands[].modes` | table | Optional, from M03. The booleans `dry_run`, `recorded`, and `resume`, each `false` when absent. `dry_run` gives the command `-n, --dry-run`, and `recorded` gives it `--run-id <id>`, after its own flags |
| `commands[].args[].values_from`, `commands[].flags[].values_from` | table | Optional, from M04. Where completion finds the values that only an extension knows, such as skill names: `command`, the path of a list command of this extension or of another one, such as `skill list` for a skill name, `key`, the key of its answer that holds an array of objects, and `field`, the string key of each object to suggest, as [Completion values](completion-values.md) defines |
| `commands[].result` | table | The command's result, the same object as `result` in [ncly describe](../core-spec.md#ncly-describe), whose fields join with their first consumer. M01 reads `kind`: `object`, the default, `batch`, `report`, or `text`, which tells core how to check the answer |
| `commands[].result.schema`, `commands[].result.dry_run_schema` | table | Optional, from M03. The JSON Schema, draft 2020-12, of the data keys of a normal answer and of a dry-run answer: an object schema, whose `required` keys a success holds. For a batch, its `results` may give the schema of an item's keys. A `$ref` points inside the same schema, such as `#/$defs/item`. [ncly describe](../core-spec.md#ncly-describe) returns the schema of the whole answer from it |
| `commands[].doctor` | boolean | Optional, from M05. `true` marks the domain's status command, which doctor runs to collect the extension's own checks, as [Status command](status-command.md) defines |
| `commands[].config` | array | Optional, from M03. The config keys that the command reads, as dotted paths under its domain, such as `agent.model`. A table key covers every key beneath it. [ncly describe](../core-spec.md#ncly-describe) reports their values and sources, and doctor checks the extension's table against them from M05, as [Configuration](configuration.md) defines |
| `commands[].requires.bins` | array | From M03. The names of the programs that the command may run, each a key of `programs` |
| `commands[].requires.keys` | array | From M03. The services whose keys the command may need, each a service name, as [Keys](keys.md) defines. `ncly auth status` and `ncly doctor` read them from M06, and core passes their granted keys from M08, as [Grants](grants.md) defines |
| `errors` | array | The extension's own error codes, each with `code` and `exit`, in the order that decides which one leads an answer, as [Output](output.md) defines |
| `programs` | table | From M03. One table per program that a command may run, keyed by the name of the executable that core looks for in `PATH`, such as `[programs.ffmpeg]`, so that each program has one definition. M03 reads the names, and M05 the keys below. [Required programs](../core-spec.md#required-programs) defines how doctor checks them |
| `programs.<name>.conditional` | boolean | Optional, from M05. `true` when a command needs the program only for some configs or invocations, and no command needs it on every call. Doctor never checks a conditional program |
| `programs.<name>.version` | string | Optional, from M05. The oldest version that works, as numbers joined by dots, such as `6.0`. Without it, finding the program is enough, and doctor starts no probe |
| `programs.<name>.probe` | array | Optional, from M05. The arguments that make the program print its version, `["--version"]` by default, such as `["-version"]` for `ffmpeg` |
| `programs.<name>.install` | table | Optional, from M05. One install command per package manager, each an array of words under the name of its package manager, such as `pacman = ["sudo", "pacman", "-S", "ffmpeg"]`. [Required programs](../core-spec.md#required-programs) lists the names that doctor knows |
| `keys` | table | Optional, from M08. One table per service that needs more than its name, keyed by the service's name, such as `[keys.anthropic]`, as `programs` holds one table per program |
| `keys.<service>.optional` | boolean | Optional, from M08. `true` when each command that declares the service runs without its key, such as a harness that falls back to its own login, as [Grants](grants.md) defines |

Core reads only the keys of this table and ignores any other, as [Compatibility](compatibility.md) asks of readers, so a later milestone adds a field with its first consumer, such as the `version` of a required program, which [M05](../../milestones/M05-doctor.md) checks. A field that an older core must not ignore needs a new `format_version`, which that core rejects instead of misreading.

A manifest must also follow the rules in [Valid manifest](valid-manifest.md).
