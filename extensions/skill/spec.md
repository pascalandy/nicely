# skill extension spec

`skill` is the bundled extension: a Go package compiled into `ncly`, with the same manifest and contract as an external extension, as [Extensions](../../docs/north-star/contract.md#extensions) defines. It is bundled because an agent needs it to operate Nicely. It follows the [contract](../../docs/north-star/contract.md). [M04](../../docs/milestones/M04-skill.md) builds it, and [M12](../../docs/milestones/M12-taps.md) adds taps and `skill link`.

## ncly skill

```
ncly skill list [--json]
ncly skill view <name> [--json]
```

- Skills come from the folders in `[skill] paths`, then from the `nicely` skill that ships inside `ncly`, then from installed extensions. Taps add more sources in M12. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter.
- `skill` finds the skills of installed extensions through `ncly describe --json`, whose entry for each extension's domain gives the folder of its skill, as [D054](../../docs/north-star/decisions/D054-extension-skills-through-describe.md) decides. It never searches for extensions itself, so it lists the skill of the extension that core runs, never one that core hides.
- A skill's subfolder in `[skill] paths`, or in a tap from M12, takes the skill's name, as the [Agent Skills specification](https://agentskills.io/specification) requires. When the names differ, the skill still works and `ncly doctor skill` warns. An extension's skill keeps the folder `ncly-<domain>.skill/`, so doctor compares its `name` with the domain instead.
- A subfolder may be a symbolic link to a folder elsewhere. Harness skill folders often hold such links, and `skill link` creates them from M12. A broken link is not a skill, and `list` skips it.
- When two sources hold the same skill name, the first source wins and `list` adds a `SKILL_SHADOWED` warning. Two subfolders that resolve to the same folder are one skill, listed with the path of the first, without a warning.
- Inside one folder of `paths`, when two subfolders declare the same `name`, the subfolder whose name sorts first byte by byte wins, and `list` adds `SKILL_SHADOWED`.
- A `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description` is not a skill. `list` skips it and adds a `SKILL_INVALID` warning whose hint names the file to fix, and `view` answers `NOT_FOUND` for it.
- A folder of `paths` that is missing or cannot be read is skipped. `list` continues with the other folders and adds a `SKILL_SOURCE_MISSING` warning whose hint names the folder and the config file that lists it, and `view` answers `NOT_FOUND` for a skill that no other source holds.
- When no folder is configured, `list` returns the skills of the other sources and adds a `SKILL_NO_SOURCE` warning whose hint names `[skill] paths` in the shared config file.
- `list` returns each skill's `name`, `description`, `path`, and `source`, sorted by name. `source` is `config` for a folder of `paths`, `bundled` for the `nicely` skill, and `extension` for an installed extension's skill. M12 adds `tap`, so a reader accepts a value that it does not know. The bundled skill has no folder, so its `path` is `null`.
- `view` prints the `SKILL.md` of the skill. When the skill has a folder, the first line gives it, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, Glamour renders the Markdown, from M07.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- An unknown name exits 2 with `NOT_FOUND`, whose `next` is `ncly skill list`.
- Completion suggests skill names after `ncly skill view`: the `name` argument declares `values_from` on `skill list`, as [Completion values](../../docs/north-star/contract.md#completion-values) defines.

```json
{"ok":true,"contract_version":1,"skills":[{"name":"nicely","description":"Use before running any ncly command.","path":null,"source":"bundled"},{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/transcript","source":"config"}]}
```

`view --json` returns `{"ok":true,"contract_version":1,"name":"…","path":"…","source":"…","content":"…"}`, with the same `path` and `source` as `list`.

### The nicely skill

`ncly` ships the `nicely` skill, which teaches an agent to operate Nicely: the variables to set, discovery with `ncly describe`, dry run before execution, codes and `next` rather than messages, retry safety, run IDs, and recovery through `ncly run`. One line in an agent's instructions points to it: run `ncly skill view nicely` before using `ncly`. The first example of `ncly --help` names it too. A skill named `nicely` in `[skill] paths` wins over it, with `SKILL_SHADOWED`.

The skill teaches only what the release that ships it can do. M04 ships the loop with `describe`, dry run, codes, and exit codes, and each milestone that adds a core command adds its lines: `doctor` in M05, `run` and run IDs in M10, and `run resume` in M15. A test fails when the skill names an `ncly` command or an error code that this `ncly` lacks. The skill holds at most 60 lines, so an agent reads it in one call.

Every extension ships a skill named after its domain, which teaches an agent to use that extension, so `ncly skill list` grows with every extension installed, and an agent that meets a domain in `ncly describe` reads `ncly skill view <domain>`. The `skill` extension ships no skill named `skill`: the `nicely` skill, which it serves as its bundled source, teaches core and the skill domain together.

## Config

```toml
[skill]
paths = ["~/code/skills"]
```

| Key | Default | Meaning |
|---|---|---|
| `paths` | none | Folders of skills, searched in order. Each one is absolute or starts with `~/`, as [Configuration](../../docs/north-star/contract.md#configuration) requires. A higher config source replaces the whole list |

## Doctor checks

The `skill` component of `ncly doctor` warns with `SKILL_SOURCE_MISSING` for a folder of `[skill] paths` that is missing or cannot be read. It warns with `SKILL_NAME_MISMATCH` when a subfolder of `[skill] paths`, or of a tap from M12, is not named after its skill, or an installed extension's skill is not named after the domain that `ncly describe` gives, and with `SKILL_INVALID` for a `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description`. A warning leaves the report `ok`.

## Warning codes

| Code | When | Since |
|---|---|---|
| `SKILL_SHADOWED` | Two skills share a name, in two sources or inside one folder of `[skill] paths` | M04 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M04 |
| `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` | M04 |
| `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read | M04 |
| `SKILL_NAME_MISMATCH` | A subfolder of `[skill] paths`, or of a tap from M12, is not named after its skill, or an extension's skill is not named after its domain | M05 |
