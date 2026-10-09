# skill extension spec

`skill` is the bundled extension: a Go package compiled into `ncly`, with the same manifest and contract as an external extension, as [Extensions](../../docs/north-star/contract.md#extensions) defines. It is bundled because an agent needs it to operate Nicely. It follows the agent contract of [contract.md](../../docs/north-star/contract.md). [M03](../../docs/milestones/M03-skill.md) builds it, and [M12](../../docs/milestones/M12-taps.md) adds taps and `skill link`.

## ncly skill

```
ncly skill list [--json]
ncly skill view <name> [--json]
ncly skill <name>
```

- Skills come from the folders in `[skill] paths`, then from the `nicely` skill that ships inside `ncly`, then from installed extensions. Taps add more sources in M12. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter.
- A skill's subfolder takes the skill's name, as the [Agent Skills specification](https://agentskills.io/specification) requires. When the names differ, the skill still works and `ncly doctor skill` warns.
- A subfolder may be a symbolic link to a folder elsewhere. Harness skill folders often hold such links, and `skill link` creates them from M12. A broken link is not a skill, and `list` skips it.
- When two sources hold the same skill name, the first source wins and `list` adds a `SKILL_SHADOWED` warning. Two subfolders that resolve to the same folder are one skill, listed with the path of the first, without a warning.
- Inside one folder of `paths`, when two subfolders declare the same `name`, the subfolder whose name sorts first byte by byte wins, and `list` adds `SKILL_SHADOWED`.
- A `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description` is not a skill. `list` skips it and adds a `SKILL_INVALID` warning whose hint names the file to fix, and `view` answers `NOT_FOUND` for it.
- A folder of `paths` that is missing or cannot be read is skipped. `list` continues with the other folders and adds a `SKILL_SOURCE_MISSING` warning whose hint names the folder, and `view` answers `NOT_FOUND` for a skill that no other source holds.
- When no folder is configured, `list` returns the skills of the other sources and adds a `SKILL_NO_SOURCE` warning whose hint names `[skill] paths`.
- `list` returns names, descriptions, and paths, sorted by name.
- `view` prints the `SKILL.md` of the skill. Its first line gives the folder, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, Glamour renders the Markdown, from M07.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- `ncly skill <name>` is a shortcut for `view`, for humans. Agents use `view`. No skill may take the name of a `skill` verb.
- An unknown name exits 2 with `NOT_FOUND`.

```json
{"ok":true,"contract_version":1,"skills":[{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/transcript"}]}
```

`view --json` returns `{"ok":true,"contract_version":1,"name":"…","path":"…","content":"…"}`.

### The nicely skill

`ncly` ships the `nicely` skill, which teaches an agent to operate Nicely: the variables to set, discovery with `ncly describe`, readiness with `ncly doctor`, dry run before execution, codes rather than messages, retry safety, and recovery through `ncly run`. One line in an agent's instructions points to it: run `ncly skill view nicely` before using `ncly`. A skill named `nicely` in `[skill] paths` wins over it, with `SKILL_SHADOWED`.

Every extension ships a `SKILL.md` too, which teaches an agent to use that extension, so `ncly skill list` grows with every extension installed.

## Config

```toml
[skill]
paths = ["~/code/skills"]
```

| Key | Default | Meaning |
|---|---|---|
| `paths` | none | Folders of skills, searched in order. A higher config source replaces the whole list |

## Doctor checks

The `skill` component of `ncly doctor` warns with `SKILL_SOURCE_MISSING` for a folder of `[skill] paths` that is missing or cannot be read. It warns with `SKILL_NAME_MISMATCH` when a subfolder's name differs from its `name`, and with `SKILL_INVALID` for a `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description`. A warning leaves the report `ok`.

## Warning codes

| Code | When | Since |
|---|---|---|
| `SKILL_SHADOWED` | Two skills share a name, in two sources or inside one folder of `[skill] paths` | M03 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M03 |
| `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` | M03 |
| `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read | M03 |
| `SKILL_NAME_MISMATCH` | A skill's subfolder name differs from the `name` in its `SKILL.md` | M05 |
