# Configuration

`config.toml` is the shared config: the setup the user wants, which can travel between machines, for example through dotfiles. `config.local.toml`, in the same folder, is the local config: what belongs to this machine only.

Core owns the top-level keys, such as `lang`. Each extension owns the table named after its slug, such as `[agent]` or `[video-convert]`.
[Extension identity](extension-identity.md) defines the slug, and the extension spec defines its table. The keys of an installed extension's table are the config keys that its manifest declares, as dotted paths that start with its slug. A declared table key covers every key beneath it; a scalar declaration such as `agent.model` covers that value.

Precedence, highest first: flags, environment variables, `config.local.toml`, `config.toml`, then defaults. A table, such as `[transcript.youtube]`, merges key by key across the files: an empty table adds no key over the keys that a lower file sets beneath it, and otherwise is the value of its key, which differs from an absent key. A list, such as `[skill] paths`, from a higher source replaces the whole list below it. Each resolved value keeps its source: the flag, the variable, the file that set it, or the default. [ncly describe](../core-spec.md#ncly-describe) reports it, and a hint about a configured value names it, such as the file that lists a missing folder.

A path in a config file is absolute or starts with `~/`, which stands for the home folder. Any other path is `CONFIG_INVALID`, because its meaning would depend on the folder where `ncly` runs.

| Purpose | Path |
|---|---|
| Shared config | `~/.config/nicely/config.toml` |
| Local config | `~/.config/nicely/config.local.toml` |
| Data: taps and extensions | `~/.local/share/nicely/` |
| State: logs, locks, run history, and grants | `~/.local/state/nicely/` |
| Cache: extracted programs, such as transcript's Python program | `~/.cache/nicely/` |

Each path honors `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, or `XDG_CACHE_HOME` when it holds an absolute path, as the XDG specification asks, and ignores a relative one. macOS uses the same paths as Linux. `NCLY_CONFIG` names another shared config file, and the local config is then read from the same folder.

- A missing config file is not an error, and the defaults apply.
- A syntax error, a value of the wrong type, including a number that JSON cannot hold, such as `nan` or `inf`, or a file that cannot be read exits 78 with `CONFIG_INVALID`, naming the file and, when the parser knows it, the line. `--help`, `--version`, `ncly completion`, and discovery still work with defaults. Discovery reports its fallback. Doctor reports this failure on stdout as a failed `core.config` check and uses defaults only for checks independent of the invalid configuration
- An unknown key is a `CONFIG_UNKNOWN_KEY` warning in `ncly doctor`, never an error, so an older `ncly` reads a config written for a newer one. Doctor checks a top-level key against core's keys, and a key of an installed extension's table against the keys that its manifest declares, so the warning never names a key that an installed extension owns. Doctor skips a table for an uninstalled extension when its name is a valid, unreserved slug under [Extension identity](extension-identity.md).
  The shared config may carry another machine's setup.
- `ncly` writes a config file only when the job of a command is to change the setup, such as a future config-editing command. It edits the file in place, keeps comments and formatting, and names the file it changed. Under `--dry-run`, it shows the change instead. Each such command states which file it writes.

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. `ncly` picks the closest catalog, so `fr_CA.UTF-8` and `fr` both select `fr-CA` once that catalog exists. A language with no close catalog, including `C` and `POSIX`, falls back to `en` without an error. `--lang` or `NCLY_LANG` that names no language is `USAGE_INVALID`, after `ncly` drops a POSIX encoding and modifier, such as `.UTF-8`. Such a value is not a BCP 47 tag, such as `123`, or holds an unknown language, such as `xx-YY`. `ncly` never checks `lang` in the config, `LC_ALL`, `LC_MESSAGES`, or `LANG`, which other machines or programs set. The pseudo-locale `en-XA` exists for tests: it marks every catalog string between `⟦` and `⟧`, and only an exact request selects it, so `en-GB` gets English. An extension picks among its own catalogs with the same rule, as [Text of an extension](text-of-an-extension.md) defines.

Numbers follow the language, such as `1,234.5` in English and `1 234,5` in Canadian French. Sizes count 1000 bytes per kilobyte, as macOS does, and name their unit in the language, such as `1.5 MB` or `1,5 Mo`. Dates use ISO 8601, such as `2026-10-06`, in every language: English readers parse it, and it is the Canadian French standard.

```toml
lang = "en"

[skill]
paths = ["~/code/skills"]
```
