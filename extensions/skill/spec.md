# skill extension spec

`skill` is the bundled extension: a Go package compiled into `ncly`, with the same manifest and contract as an external extension, as [Extensions](../../docs/north-star/contract.md#extensions) defines. It is bundled because an agent needs it to operate Nicely. It follows the [contract](../../docs/north-star/contract.md). [M04](../../docs/milestones/M04-skill.md) builds it, [M05](../../docs/milestones/M05-doctor.md) adds `status`, and [M12](../../docs/milestones/M12-taps.md) adds taps and `skill link`.

## ncly skill

```
ncly skill list [--json]
ncly skill view <name> [--json]
ncly skill status [--json]
```

- Skills come from the folders in `[skill] paths`, then taps in effective config order from M12, then bundled `nicely`, then installed extensions. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter.
- `skill` finds the skills of installed extensions through `ncly describe --json`, whose entry for each extension's domain gives the folder of its skill, as [D054](../../docs/north-star/decisions/D054-extension-skills-through-describe.md) decides. It never searches for extensions itself, so it lists the skill of the extension that core runs, never one that core hides. It runs `"$NCLY_BIN" describe --json` through the program runner, and when that call fails or gives an answer that it cannot read, `list` and `view` fail with `RUNTIME`, whose hint and `next` are `ncly describe --json`.
- A skill's subfolder in `[skill] paths`, or in a tap from M12, takes the skill's name, as the [Agent Skills specification](https://agentskills.io/specification) requires. When the names differ, the skill still works, and `ncly skill status` warns. An extension's skill keeps the folder `ncly-<domain>.skill/`, so `status` compares its `name` with the domain instead.
- A subfolder may be a symbolic link to a folder elsewhere. Harness skill folders often hold such links, and `skill link` creates them from M12. A broken link is not a skill, and `list` skips it.
- When two sources hold the same skill name, the first source wins and `list` adds a `SKILL_SHADOWED` warning. Two subfolders that resolve to the same folder are one skill, listed with the path of the first, without a warning.
- Inside one folder of `paths`, when two subfolders declare the same `name`, the subfolder whose name sorts first byte by byte wins, and `list` adds `SKILL_SHADOWED`.
- A `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description` is not a skill. `list` skips it and adds a `SKILL_INVALID` warning whose hint names the file to fix, and `view` answers `NOT_FOUND` for it.
- A folder of `paths` that is missing or cannot be read is skipped. `list` continues with the other folders and adds a `SKILL_SOURCE_MISSING` warning whose hint names the folder and the config file that lists it, and `view` answers `NOT_FOUND` for a skill that no other source holds.
- When no folder is configured, `list` returns the skills of the other sources and adds a `SKILL_NO_SOURCE` warning whose hint names `[skill] paths` in the shared config file.
- A config file that cannot be read or holds a syntax error, or a `paths` that is not a list of folders, each absolute or starting with `~/`, fails `list` with `CONFIG_INVALID`, whose hint names the file, as [Configuration](../../docs/north-star/contract.md#configuration) requires of every command that reads the config.
- `list` returns each skill's `name`, `description`, `path`, and `source`, sorted by name byte by byte, in `skills`, an empty array when no source holds a skill. Without `--json`, it prints one line per skill: its name, in bold, and a colon, then its full description, aligned as the help aligns its commands, so the terminal wraps a long description. `source` is `config` for a folder of `paths`, `bundled` for the `nicely` skill, and `extension` for an installed extension's skill. M12 adds `tap`, so a reader accepts a value that it does not know. The bundled skill has no folder, so its `path` is `null`.
- `view` prints the `SKILL.md` of the skill. When the skill has a folder, the first line gives it, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, from M07, Glamour renders the Markdown in a style that matches the help: headings in bold uppercase, code in cyan, dashes for bullets, and ANSI colors only. `view` then removes the frontmatter, which Glamour would draw as a rule and a heading, and shows the folder as a first line `folder  <path>` instead of the comment. Without color, code keeps its backticks, so it stays apart from the prose.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- `view` finds the skill as `list` does, and adds none of the warnings of `list`, which describe other skills.
- An unknown name exits 2 with `NOT_FOUND`, whose `next` is `ncly skill list`.
- Completion suggests skill names after `ncly skill view`: the `name` argument declares `values_from` on `skill list`, as [Completion values](../../docs/north-star/contract.md#completion-values) defines.

```json
{"ok":true,"contract_version":1,"skills":[{"name":"nicely","description":"Use before running any ncly command.","path":null,"source":"bundled"},{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/transcript","source":"config"}]}
```

`view --json` returns `{"ok":true,"contract_version":1,"name":"…","path":"…","source":"…","content":"…"}`, with the same `path` and `source` as `list`.

From M12, `ncly skill view <tap>/<name>` selects a standalone skill from that tap, even when another source hides it.
`<tap>` is its canonical tap ID, as [Tap identity and storage](../../docs/north-star/core-spec.md#tap-identity-and-storage) defines.
The tap's publication supplies the skill path. `list` adds `tap` and `revision` to a standalone tap skill's entry.
`view --json` adds those fields and `shadowed`, an array of hidden sources with `source`, `path`, and, for taps, `tap` and `revision`.
It also adds `links`, each with `harness`, `path`, and `target`, for the selected skill's managed harness links.
These fields describe files, never what a running harness has loaded. `list` keeps one line per selected skill.
Installed extension skills still come only from `ncly describe`, with its selected tap origin, as [Discovery and grants](../../docs/north-star/core-spec.md#discovery-and-grants) defines.
A standalone tap skill's `nicely.toml` uses the [manifest's](../../docs/north-star/contract.md#manifest) `programs`, `keys`, and `requires` formats for its requirements.

### The nicely skill

`ncly` ships the `nicely` skill, which teaches an agent to operate Nicely: the variables to set, discovery with `ncly describe`, dry run before execution, codes and `next` rather than messages, retry safety, run IDs, and recovery through `ncly run`. One line in an agent's instructions points to it: run `ncly skill view nicely` before using `ncly`. The first example of `ncly --help` names it too. A skill named `nicely` in `[skill] paths` wins over it, with `SKILL_SHADOWED`.

The skill teaches only what the release that ships it can do. M04 ships the loop with `describe`, dry run, codes, and exit codes, and each milestone that adds a command of core or of the skill domain adds its lines: `doctor` and `skill status` in M05, `auth status` and the rule that an agent relays a missing key to the human and never pipes one itself in M06, `KEY_NOT_GRANTED` under the same rule in M08, `run` and run IDs in M10, and `run resume` in M15. The skill writes each command in a code span, such as `ncly skill list`, so that a test reads it. A test fails when the skill names an `ncly` command or an error code that this `ncly` lacks. The skill holds at most 60 lines, so an agent reads it in one call.

M12 adds the tap commands and `skill link`. It invites users to install the recommended first-party extensions with `ncly tap add pascalandy/nicely`.

Every extension ships a skill named after its domain, which teaches an agent to use that extension, so `ncly skill list` grows with every extension installed, and an agent that meets a domain in `ncly describe` reads `ncly skill view <domain>`. The `skill` extension ships no skill named `skill`: the `nicely` skill, which it serves as its bundled source, teaches core and the skill domain together.

## Config

```toml
[skill]
paths = ["~/code/skills"]
```

| Key | Default | Meaning |
|---|---|---|
| `paths` | none | Folders of skills, searched in order. Each one is absolute or starts with `~/`, as [Configuration](../../docs/north-star/contract.md#configuration) requires. A higher config source replaces the whole list |

## ncly skill link

```text
ncly skill link [<name>] [--harness <name>] [--dry-run]
```

From M12, `link` creates a symlink for the selected skill, or every selected skill when `<name>` is absent.
Without `--harness`, it targets every configured harness. An unknown skill or harness answers `NOT_FOUND`.
The harness names and destination folders come from `[skill.link]`:

```toml
[skill.link]
claude = "~/.claude/skills"
```

Each value is absolute or starts with `~/`, as [Configuration](../../docs/north-star/contract.md#configuration) requires. No configured harness is `CONFIG_INVALID` with a hint naming that table.
Each link is `<harness-folder>/<skill-name>` and targets the selected skill folder. Tap targets use that tap's `current` path, so sync selects the new files.
`link` skips the bundled `nicely` skill, which agents read with `ncly skill view nicely`.
A link is managed when its target is a skill folder that Nicely selects or a path under a tap's `current` link.
Any other file or link at the destination is a conflict. The command answers `RUNTIME` and preserves it.
Repeating `link` leaves a matching managed link alone. A source switch replaces only a managed link.
Link updates follow [Publication and recovery](../../docs/north-star/core-spec.md#publication-and-recovery), including locks, revalidation, and whole-invocation retry checks.
Tap removal removes the links in the `[skill.link]` folders whose targets lie inside that tap's storage.
Core inspects those targets without importing the skill extension. It leaves a user-modified link to another target alone, so that link does not block removal.

The answer has `links`, sorted by harness and skill name. Each entry has `name`, `harness`, `path`, `target`, and `action`: `created`, `updated`, or `unchanged`.
A partial failure keeps the completed entries beside `errors`. Dry run returns `dry_run: true`, `plan`, and `pending_checks`, which is empty.
Each plan entry names the same link fields, its intended action, and `effects`: `user_writes: true`, `network: false`, and `paid: false`.
Dry run checks conflicts and changes no files or links.

## ncly skill status

`status` is the status command of the skill domain, as [Status command](../../docs/north-star/contract.md#status-command) defines, so its checks form the `skill` component of `ncly doctor`. Its declaration sets `doctor = true`, has no flag, and declares effects that neither write nor bill. It reports each kind of problem under one `id`, with one `warn` entry per problem, whose hint names the folder or the file to fix, and one `pass` entry when it finds none:

| Id | Code | When it warns |
|---|---|---|
| `skill.sources` | `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read |
| `skill.names` | `SKILL_NAME_MISMATCH` | A subfolder of `[skill] paths`, or of a tap from M12, is not named after its skill, or an installed extension's skill is not named after the domain that `ncly describe` gives |
| `skill.files` | `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` |

A warning leaves the report `ok`, so `status` exits 0. A broken config fails the command with `CONFIG_INVALID`, as `list` does.

## Warning codes

| Code | When | Since |
|---|---|---|
| `SKILL_SHADOWED` | Two skills share a name, in two sources or inside one folder of `[skill] paths` | M04 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M04 |
| `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` | M04 |
| `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read | M04 |
| `SKILL_NAME_MISMATCH` | A subfolder of `[skill] paths`, or of a tap from M12, is not named after its skill, or an extension's skill is not named after its domain | M05 |
