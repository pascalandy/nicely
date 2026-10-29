# skill extension spec

`skill` is the bundled extension: a Go package compiled into `ncly`, with the same manifest and contract as an external extension, as [Extensions](../../docs/north-star/contract/extensions.md) defines. It is bundled because an agent needs it to operate Nicely. It follows the [contract](../../docs/north-star/contract.md). [M04](../../docs/milestones/M04-skill.md) built it, [M05](../../docs/milestones/M05-doctor.md) added `status`. Taps and `skill link` remain in [COR-622](../../docs/backlog/COR-622-taps-offer.md#later).

## ncly skill

```
ncly skill list [--json]
ncly skill view <name> [--json]
ncly skill status [--json]
```

- Skills come from the folders in `[skill] paths`, then from the `nicely` skill that ships inside `ncly`, then from installed extensions. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter.
- `skill` finds the skills of installed extensions through `ncly describe --json`, whose extension root entries give their skill folders, as [D054](../../docs/north-star/decisions/D054-extension-skills-through-describe.md) decides. It never searches for extensions itself, so it lists the skill of the extension that core runs, never one that core hides. It runs `"$NCLY_BIN" describe --json` through the program runner, and when that call fails or gives an answer that it cannot read, `list` and `view` fail with `RUNTIME`, whose hint and `next` are `ncly describe --json`.
- A skill's subfolder in `[skill] paths` takes the skill's name, as the [Agent Skills specification](https://agentskills.io/specification) requires. When the names differ, the skill still works, and `ncly skill status` warns. For an extension, `status` derives the slug from its root entry's describe `path`, under [Extension identity](../../docs/north-star/contract/extension-identity.md#identity-and-root). It compares the skill's `name` with that slug.
- A subfolder may be a symbolic link to a folder elsewhere. Harness skill folders often hold such links. A broken link is not a skill, and `list` skips it.
- When two sources hold the same skill name, the first source wins and `list` adds a `SKILL_SHADOWED` warning. Two subfolders that resolve to the same folder are one skill, listed with the path of the first, without a warning.
- Inside one folder of `paths`, when two subfolders declare the same `name`, the subfolder whose name sorts first byte by byte wins, and `list` adds `SKILL_SHADOWED`.
- A `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description` is not a skill. `list` skips it and adds a `SKILL_INVALID` warning whose hint names the file to fix, and `view` answers `NOT_FOUND` for it.
- A folder of `paths` that is missing or cannot be read is skipped. `list` continues with the other folders and adds a `SKILL_SOURCE_MISSING` warning whose hint names the folder and the config file that lists it, and `view` answers `NOT_FOUND` for a skill that no other source holds.
- `list` adds no warning when no folder is configured.
- A config file that cannot be read or holds a syntax error, or a `paths` that is not a list of folders, each absolute or starting with `~/`, fails `list` with `CONFIG_INVALID`, whose hint names the file, as [Configuration](../../docs/north-star/contract/configuration.md) requires of every command that reads the config.
- `list` returns each skill's `name`, `description`, `path`, and `source`, sorted by name byte by byte, in `skills`, an empty array when no source holds a skill. Without `--json`, it prints one line per skill: its name, in bold, and a colon, then its full description, aligned as the help aligns its commands, so the terminal wraps a long description. `source` is `config` for a folder of `paths`, `bundled` for the `nicely` skill, and `extension` for an installed extension's skill. A reader accepts a future source value that it does not know. The bundled skill has no folder, so its `path` is `null`.
- `view` prints the `SKILL.md` of the skill. When the skill has a folder, the first line gives it, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, from M07, Glamour renders the Markdown in a style that matches the help: headings in bold uppercase, inline code in cyan, dashes for bullets, and ANSI colors only. `view` then removes the frontmatter, which Glamour would draw as a rule and a heading, and shows the folder as a first line `folder  <path>`, with `folder` in bright black, instead of the comment. Without color, inline code keeps its backticks, so it stays apart from the prose. A code block keeps its lines as written, in the margin of the prose, without color. A link or an image shows its destination as the skill writes it, so a relative path still resolves against the folder, and a heading keeps the case of code, links, and images, so they copy whole.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- `view` finds the skill as `list` does, and adds none of the warnings of `list`, which describe other skills.
- An unknown name exits 2 with `NOT_FOUND`, whose `next` is `ncly skill list`.
- Completion suggests skill names after `ncly skill view`: the `name` argument declares `values_from` on `skill list`, as [Completion values](../../docs/north-star/contract/completion-values.md) defines.

```json
{"ok":true,"contract_version":1,"skills":[{"name":"nicely","description":"Use before running any ncly command.","path":null,"source":"bundled"},{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video.","path":"/Users/me/code/skills/transcript","source":"config"}]}
```

`view --json` returns `{"ok":true,"contract_version":1,"name":"…","path":"…","source":"…","content":"…"}`, with the same `path` and `source` as `list`.

### The nicely skill

`ncly` ships the `nicely` skill, which teaches an agent to operate Nicely: the variables to set, discovery with `ncly describe`, dry run before execution, codes and `next` rather than messages, retry safety, run IDs, and recovery through `ncly run`. One line in an agent's instructions points to it: `When asked to use nicely, run ncly skill view nicely`. The first example of `ncly --help` names it too. A skill named `nicely` in `[skill] paths` wins over it, with `SKILL_SHADOWED`.

The skill teaches only what the release that ships it can do. M04 shipped the loop with `describe`, dry run, codes, and exit codes. Each item that adds a core or skill command adds its lines. M05 added `doctor` and `skill status`. M06 added `auth status` and the rule that an agent relays a missing key to the human and never pipes one itself. M08 added `auth grant`, `auth revoke`, and `KEY_NOT_GRANTED` under the same rule. M10 added `run list`, `run view`, and run IDs. M18 adds the command that creates an extension repository from the template, with a link to [Extensions](../../docs/north-star/contract/extensions.md). #340 adds the install of an extension at a tag, which quotes its rule there. Explicit resume remains in [COR-621](../../docs/backlog/COR-621-resume-sdk.md#later). The skill writes each command in a code span, such as `ncly skill list`, so that a test reads it. A test fails when the skill names an `ncly` command, or an error or warning code, that this `ncly` lacks. The skill holds at most 60 lines, so an agent reads it in one call.

Every external extension ships a skill named after its slug.
`ncly skill list` grows with installed extensions.
An agent reads `ncly skill view <slug>` for an extension root that it finds through describe. The `skill` extension ships no skill named `skill`: the `nicely` skill, which it serves as its bundled source, teaches core and the skill domain together.

## Config

```toml
[skill]
paths = ["~/code/skills"]
```

| Key | Default | Meaning |
|---|---|---|
| `paths` | none | Folders of skills, searched in order. Each one is absolute or starts with `~/`, as [Configuration](../../docs/north-star/contract/configuration.md) requires. A higher config source replaces the whole list |

## ncly skill status

`status` is the status command of the skill domain, as [Status command](../../docs/north-star/contract/status-command.md) defines, so its checks form the `skill` component of `ncly doctor`. Its declaration sets `doctor = true`, has no flag, and declares effects that neither write nor bill. It reports each kind of problem under one `id`, with one `warn` entry per problem, whose hint names the folder or the file to fix, and one `pass` entry when it finds none:

| Id | Code | When it warns |
|---|---|---|
| `skill.sources` | `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read |
| `skill.sources` | `SKILL_NO_SOURCE` | No folder of `[skill] paths` is configured |
| `skill.names` | `SKILL_NAME_MISMATCH` | A configured subfolder differs from its skill name, or an extension's skill differs from the slug derived from its describe `path` |
| `skill.files` | `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` |

A warning leaves the report `ok`, so `status` exits 0. A broken config fails the command with `CONFIG_INVALID`, as `list` does.

## Warning codes

| Code | When | Since |
|---|---|---|
| `SKILL_SHADOWED` | Two skills share a name, in two sources or inside one folder of `[skill] paths` | M04 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M04 |
| `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` | M04 |
| `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read | M04 |
| `SKILL_NAME_MISMATCH` | A subfolder of `[skill] paths` is not named after its skill, or an extension's skill is not named after its slug | M05 |
