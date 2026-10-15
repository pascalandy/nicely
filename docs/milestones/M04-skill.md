# M04 Skill

Status: active
Version: v0.4.0

## Demo

`ncly skill list` and `ncly skill view` run as the first bundled extension, over the configured folders, the installed extensions, and the bundled `nicely` skill, which teaches an agent to use Nicely.

## Scope

`skill` is bundled because an agent needs it to operate Nicely: one line in an agent's instructions, `ncly skill view nicely`, replaces the long text that README.md used to plan. Its spec is [extensions/skill/spec.md](../../extensions/skill/spec.md). Output stays plain in this milestone. The styled list and the Markdown rendering of `view` arrive in [M07](M07-forms.md).

[M03](M03-describe.md) comes first, so `skill` reads the skills of installed extensions through `ncly describe --json`, as an agent would, and needs no discovery code of its own.

The `nicely` skill teaches only what this release can do, as [The nicely skill](../../extensions/skill/spec.md#the-nicely-skill) says, and [M05](M05-doctor.md), [M06](M06-auth.md), [M10](M10-agent-run.md), and [M15](M15-resume.md) add their lines. Its draft for this milestone:

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
| M04-T1 | Skill list | agent | — | done |
| M04-T2 | Skill view | agent | M04-T1 | done |
| M04-T3 | Skills from extensions | agent | M04-T2 | todo |
| M04-T4 | The nicely skill | agent | M04-T2 | todo |
| M04-T5 | Skill name completion | agent | M04-T2 | todo |

### M04-T1 Skill list

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Configuration](../north-star/contract.md#configuration), [SDK](../north-star/contract.md#sdk)
- **Proves:** `extensions/skill/testdata/script/skill_list.txtar`, `extensions/skill/testdata/script/skill_shadowed.txtar`, `extensions/skill/testdata/script/skill_invalid.txtar`, `extensions/skill/testdata/script/skill_links.txtar`

`ncly skill list` over `[skill] paths`, as the first bundled extension, with `source` and `SKILL_SHADOWED`, `SKILL_NO_SOURCE`, `SKILL_INVALID`, and `SKILL_SOURCE_MISSING`. As the first command that fails on a broken config, where discovery falls back to defaults, it maps an unreadable, malformed, or wrongly typed config to `CONFIG_INVALID`, including a relative folder in `paths`. Its `[skill]` table decodes through `sdk/config`, which keeps each value's source from [M03-T4](M03-describe.md#m03-t4-config-sources). Add `skill` to the bundled list of `cmd/ncly`, and update that list's comment. The extension lives in `extensions/skill/`. It reads the frontmatter of a `SKILL.md` with `go.yaml.in/yaml/v3`, through a reader in `sdk/extension`, because an extension imports only `sdk/` and `Check` reuses that reader in [M04-T3](#m04-t3-skills-from-extensions). The scenarios cover these edge cases:

- Two subfolders of one folder declare the same `name`, and the first byte by byte wins with `SKILL_SHADOWED`
- A `SKILL.md` lacks `description`, so `list` omits it, and the `SKILL_INVALID` hint names its path
- A valid folder comes before a missing one in `paths`, so `list` returns the valid skills and adds `SKILL_SOURCE_MISSING`, whose hint names the config file that lists the missing folder
- A linked skill folder is listed once with the path of the link, and a broken link is skipped

### M04-T2 Skill view

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Answer check](../north-star/contract.md#answer-check)
- **Proves:** `extensions/skill/testdata/script/skill_view.txtar`, `testdata/script/extension_protocol.txtar`

`ncly skill view <name>`, with or without `--json`. The text view adds the folder line only when the skill has a folder, and `--json` returns the fields that [ncly skill](../../extensions/skill/spec.md#ncly-skill) specifies. An unknown name, or a skill that `list` omits as invalid, answers `NOT_FOUND` with exit 2 and the `next` `ncly skill list`. As the first answer with `next` that crosses the answer check, it proves that core passes `next` on from an extension. The answer check in `sdk/contract` gains the `next` part of its rule 4, and `extension_protocol.txtar` proves with the shell fixture that core rejects a `next` that is not a non-empty array of strings, an answer that an extension built on the SDK cannot give.

### M04-T3 Skills from extensions

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Extensions](../north-star/contract.md#extensions), [SDK](../north-star/contract.md#sdk), [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `extensions/skill/testdata/script/skill_extension.txtar`

The skill of an installed extension appears in `list` with the source `extension`, and `view` reads it, so every extension can teach agents how to use it. `skill` finds it through the `skill` key of `"$NCLY_BIN" describe --json`, which it runs through the program runner of `sdk/runner`, never by searching for extensions itself, so the skill of an extension that core hides never appears. The shell fixture of M01 gets a `ncly-hello.skill/` folder. `extension.Extension` in `sdk/extension` gains `Skill`, and the Go fixture embeds its skill there. The `install` command of the test binary in `cmd/ncly` writes that skill as `ncly-hello.skill/`, as installing an external extension does. `Check` refuses a skill whose `name` is not the domain, read with the reader of M04-T1.

### M04-T4 The nicely skill

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Agents in both directions](../north-star/vision.md#agents-in-both-directions)
- **Proves:** `extensions/skill/testdata/script/skill_nicely.txtar`, and the session's transcript, attached to the pull request

The skill extension embeds the `nicely` skill as the data of its bundled source, not as a skill of its own domain, and `list` shows it with the source `bundled` and a `null` path. A skill named `nicely` in `[skill] paths` wins with `SKILL_SHADOWED`. The first example of `ncly --help`, in `internal/cli`, becomes `ncly skill view nicely`. README.md tells users to add one line to their agent instructions: run `ncly skill view nicely` before using `ncly`. The scenario fails when the skill names an `ncly` command or an error code that this `ncly` lacks, or holds more than 60 lines. It reads the commands from `ncly describe --json`, and the codes from a command of the test binary in `cmd/ncly` that lists the registry of `sdk/contract`, because no command of `ncly` lists its codes.

Then an agent in a fresh harness session, given only that line, finds and reads a skill using only `ncly`. Fix the `nicely` skill until it does.

### M04-T5 Skill name completion

- **Read:** [Completion values](../north-star/contract.md#completion-values), [Manifest](../north-star/contract.md#manifest), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs), [ncly completion](../north-star/core-spec.md#ncly-completion), [ncly describe](../north-star/core-spec.md#ncly-describe)
- **Proves:** `extensions/skill/testdata/script/skill_completion.txtar`, `testdata/script/completion.txtar`, `testdata/script/extension_manifest.txtar`, `sdk/contract/compat_test.go`

Add `values_from` to the manifest in `sdk/contract`, with the one check that a list command qualifies, which the manifest and completion share. Core's completion in `internal/cli` dispatches the list command through the extension that discovery found, and the program runner in `sdk/runner` gains a stop at the limit with neither cleanup nor drain limit. The `name` argument of `skill view` declares `values_from` on `skill list`, so `ncly skill view <Tab>` suggests the names that `ncly skill list` returns, and `ncly describe 'skill view' --json` shows that `values_from` on `name`. A consumer case in `compat_test.go` reads an argument with and without `values_from`. With the shell fixture, `completion.txtar` proves that:

- A flag that takes a value completes from its `values_from` with the values that match the typed prefix
- A `values_from` that names the list command of a second fixture completes from it, and one whose foreign list command is not installed or does not qualify suggests nothing
- An argument value and a flag value that the list does not hold both reach the fixture's handler, because `values_from` never restricts what a command accepts
- A list command that fails, that answers without the key, or that outlasts the 2 seconds suggests nothing and prints nothing, and the one that outlasts them gets SIGKILL and never SIGTERM. With `BASH_COMP_DEBUG_FILE` set, that file names the timeout as the reason
- A list command whose descendant keeps stdout open still ends completion within the 2 seconds

`extension_manifest.txtar` proves that a `values_from` on a switch makes the manifest invalid, and so does one that names a command of its own domain that is undeclared or does not qualify: a group, a command whose verb is not `list`, that takes an argument, whose kind is not `object`, that omits `effects`, or whose effects declare `user_writes` or `paid`.
