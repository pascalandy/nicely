# M04 Skill

Status: planned
Version: v0.4.0

## Demo

`ncly skill list` and `ncly skill view` run as the first bundled extension, over the configured folders, the installed extensions, and the bundled `nicely` skill, which teaches an agent to use Nicely.

## Scope

`skill` is bundled because an agent needs it to operate Nicely: one line in an agent's instructions, `ncly skill view nicely`, replaces the long text that README.md used to plan. Its spec is [extensions/skill/spec.md](../../extensions/skill/spec.md). Output stays plain in this milestone. The styled list and the Markdown rendering of `view` arrive in [M07](M07-forms.md).

[M03](M03-describe.md) comes first, so `skill` reads the skills of installed extensions through `ncly describe --json`, as an agent would, and needs no discovery code of its own.

The draft of the `nicely` skill, from the agent instructions that README.md planned. Lines about transcript move to the transcript skill in [M16](M16-transcript-finish.md):

```text
Set NCLY_NO_INPUT=1 and NCLY_JSON=1 for ncly calls
Use a targeted ncly describe, the appropriate ncly doctor component, then the command's --dry-run before execution
Find skills with ncly skill list and read one with ncly skill view NAME
Parse codes and fields, never translated messages
Retain the invocation's RFC3339 start time, exit code, full stdout and stderr, and run_id
Give the shell a timeout longer than ncly's timeout plus 15 seconds of cleanup and a margin
Only overall exit 75 permits an automatic repeat of the same invocation. Item errors do not permit a retry. Relay human-action hints
After lost output, use ncly run list --path PATH --key KEY --since RFC3339 --json with the original command path, item key, and start time. If the time is lost, omit --since. Preserve ambiguous matches
Inspect candidates with ncly run view RUN_ID. active reports the execution lock. Never infer that a run is idle from its saved status
For an inactive run, inspect ncly run resume RUN_ID --dry-run before explicitly resuming supported work
```

## Open questions

Settle each one in [extensions/skill/spec.md](../../extensions/skill/spec.md), then delete this section.

- Where an installed extension keeps its `SKILL.md`, such as beside its manifest, and how `list` names that source
- The final text of the `nicely` skill, and how it names commands that later milestones add
- How completion reaches values that only an extension knows, such as skill names, and profile names from [M09](M09-agent-profiles.md). Settle it in [Extensions](../north-star/contract.md#extensions)

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

`ncly skill list` over `[skill] paths`, as the first bundled extension, with `SKILL_SHADOWED`, `SKILL_NO_SOURCE`, `SKILL_INVALID`, and `SKILL_SOURCE_MISSING`. As the first command that reads the config, it maps an unreadable, malformed, or wrongly typed config to `CONFIG_INVALID`. The scenarios cover these edge cases:

- Two subfolders of one folder declare the same `name`, and the first byte by byte wins with `SKILL_SHADOWED`
- A `SKILL.md` lacks `description`, so `list` omits it, and the `SKILL_INVALID` hint names its path
- A valid folder comes before a missing one in `paths`, so `list` returns the valid skills and adds `SKILL_SOURCE_MISSING`
- A linked skill folder is listed once with the path of the link, and a broken link is skipped

### M04-T2 Skill view

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `extensions/skill/testdata/script/skill_view.txtar`

`ncly skill view <name>`, with or without `--json`, and the `ncly skill <name>` shortcut. The first line names the skill's folder. An unknown name, or a skill that `list` omits as invalid, answers `NOT_FOUND` with exit 2.

### M04-T3 Skills from extensions

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Extensions](../north-star/contract.md#extensions), [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `extensions/skill/testdata/script/skill_extension.txtar`

The `SKILL.md` of an installed extension appears in `list` with its source, and `view` reads it, so every extension can teach agents how to use it. `skill` finds it through `ncly describe --json`, never by searching for extensions itself.

### M04-T4 The nicely skill

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Agents in both directions](../north-star/vision.md#agents-in-both-directions)
- **Proves:** `extensions/skill/testdata/script/skill_nicely.txtar`, and the session's transcript, attached to the pull request

The skill extension embeds the `nicely` skill, which `list` shows with its bundled source. A skill named `nicely` in `[skill] paths` wins with `SKILL_SHADOWED`. README.md tells users to add one line to their agent instructions: run `ncly skill view nicely` before using `ncly`.

Then an agent in a fresh harness session, given only that line, finds and reads a skill using only `ncly`. Fix the `nicely` skill until it does.

### M04-T5 Skill name completion

- **Read:** [ncly completion](../north-star/core-spec.md#ncly-completion), [Extensions](../north-star/contract.md#extensions)
- **Proves:** `extensions/skill/testdata/script/skill_completion.txtar`

Completion suggests skill names after `ncly skill view`, through the mechanism that this milestone settles for the values that only an extension knows.
