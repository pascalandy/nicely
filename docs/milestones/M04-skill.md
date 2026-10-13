# M04 Skill

Status: planned
Version: v0.4.0

## Demo

`ncly skill list` and `ncly skill view` run as the first bundled extension, over the configured folders, the installed extensions, and the bundled `nicely` skill, which teaches an agent to use Nicely.

## Scope

`skill` is bundled because an agent needs it to operate Nicely: one line in an agent's instructions, `ncly skill view nicely`, replaces the long text that README.md used to plan. Its spec is [extensions/skill/spec.md](../../extensions/skill/spec.md). Output stays plain in this milestone. The styled list and the Markdown rendering of `view` arrive in [M07](M07-forms.md).

[M03](M03-describe.md) comes first, so `skill` reads the skills of installed extensions through `ncly describe --json`, as an agent would, and needs no discovery code of its own.

The `nicely` skill teaches only what this release can do, as [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill) says, and [M05](M05-doctor.md), [M10](M10-agent-run.md), and [M15](M15-resume.md) add their lines. Its draft for this milestone:

```text
Set NCLY_NO_INPUT=1 and NCLY_JSON=1 for every ncly call
List commands with ncly describe --json, then read one with ncly describe PATH --json: its flags, effects, modes, and result keys
Run a command that writes or bills with --dry-run first. The dry run checks what that call needs
Branch on the exit code and errors[].code, never on message or hint. next holds the command to run as words
Exit 2: fix the call. Exit 78: relay hint and next to the human. Exit 1: inspect before anything else. Only exit 75 permits repeating the same call
An item error describes its item and never permits a retry
Give the shell a timeout longer than the command's --timeout plus 15 seconds of cleanup
Each extension ships a skill named after its domain: ncly skill view DOMAIN. The skill domain has none: this skill covers it and core. ncly skill list lists every skill
```

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M04-T1 | Skill list | agent | — | todo |
| M04-T2 | Skill view | agent | M04-T1 | todo |
| M04-T3 | Skills from extensions | agent | M04-T1 | todo |
| M04-T4 | The nicely skill | agent | M04-T2 | todo |
| M04-T5 | Skill name completion | agent | M04-T1 | todo |

### M04-T1 Skill list

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Configuration](../north-star/contract.md#configuration)
- **Proves:** `extensions/skill/testdata/script/skill_list.txtar`, `extensions/skill/testdata/script/skill_shadowed.txtar`, `extensions/skill/testdata/script/skill_invalid.txtar`, `extensions/skill/testdata/script/skill_links.txtar`

`ncly skill list` over `[skill] paths`, as the first bundled extension, with `source` and `SKILL_SHADOWED`, `SKILL_NO_SOURCE`, `SKILL_INVALID`, and `SKILL_SOURCE_MISSING`. As the first command that fails on a broken config, where discovery falls back to defaults, it maps an unreadable, malformed, or wrongly typed config to `CONFIG_INVALID`, including a relative folder in `paths`. Its `[skill]` table decodes through `sdk/config`, which keeps each value's source from [M03-T4](M03-describe.md#m03-t4-config-sources). Add `skill` to the bundled list of `cmd/ncly`, and update that list's comment. The scenarios cover these edge cases:

- Two subfolders of one folder declare the same `name`, and the first byte by byte wins with `SKILL_SHADOWED`
- A `SKILL.md` lacks `description`, so `list` omits it, and the `SKILL_INVALID` hint names its path
- A valid folder comes before a missing one in `paths`, so `list` returns the valid skills and adds `SKILL_SOURCE_MISSING`, whose hint names the config file that lists the missing folder
- A linked skill folder is listed once with the path of the link, and a broken link is skipped

### M04-T2 Skill view

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `extensions/skill/testdata/script/skill_view.txtar`

`ncly skill view <name>`, with or without `--json`. The text view adds the folder line only when the skill has a folder, and `--json` returns the fields that [ncly skill](../../extensions/skill/spec.md#ncly-skill) specifies. An unknown name, or a skill that `list` omits as invalid, answers `NOT_FOUND` with exit 2 and the `next` `ncly skill list`. As the first answer with `next` that crosses the answer check, it proves that core passes `next` on from an extension and rejects a `next` that is not an array of strings.

### M04-T3 Skills from extensions

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Extensions](../north-star/contract.md#extensions), [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `extensions/skill/testdata/script/skill_extension.txtar`

The skill of an installed extension appears in `list` with the source `extension`, and `view` reads it, so every extension can teach agents how to use it. `skill` finds it through the `skill` key of `ncly describe --json`, never by searching for extensions itself, so the skill of an extension that core hides never appears. The shell fixture of M01 gets a `ncly-hello.skill/` folder, the Go fixture embeds its skill, and `Check` refuses a skill whose `name` is not the domain.

### M04-T4 The nicely skill

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Agents in both directions](../north-star/vision.md#agents-in-both-directions)
- **Proves:** `extensions/skill/testdata/script/skill_nicely.txtar`, and the session's transcript, attached to the pull request

The skill extension embeds the `nicely` skill as the data of its bundled source, not as a skill of its own domain, and `list` shows it with the source `bundled` and a `null` path. A skill named `nicely` in `[skill] paths` wins with `SKILL_SHADOWED`. The first example of `ncly --help` becomes `ncly skill view nicely`. README.md tells users to add one line to their agent instructions: run `ncly skill view nicely` before using `ncly`. The scenario fails when the skill names an `ncly` command or an error code that this `ncly` lacks, or holds more than 60 lines.

Then an agent in a fresh harness session, given only that line, finds and reads a skill using only `ncly`. Fix the `nicely` skill until it does.

### M04-T5 Skill name completion

- **Read:** [Completion values](../north-star/contract.md#completion-values), [Manifest](../north-star/contract.md#manifest), [ncly completion](../north-star/core-spec.md#ncly-completion)
- **Proves:** `extensions/skill/testdata/script/skill_completion.txtar`, `testdata/script/completion.txtar`, `testdata/script/extension_manifest.txtar`

Add `values_from` to the manifest in `sdk/contract` and to core's completion in `internal/cli`. The `name` argument of `skill view` declares it on `skill list`, so `ncly skill view <Tab>` suggests the names that `ncly skill list` returns. With the shell fixture, `completion.txtar` proves that a list command that fails, that answers without the key, or that outlasts the 2 seconds suggests nothing and prints nothing. `extension_manifest.txtar` proves that a `values_from` that names an undeclared command, a group, or a command whose kind is not `object` makes the manifest invalid.
