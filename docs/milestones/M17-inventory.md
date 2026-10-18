# M17 First everyday task

Status: planned
Version: v0.17.0

## Demo

Pascal chooses one recurring task, from a concrete before and after, as Nicely's first useful command requiring no key or configuration.

## Scope

The grooming of 2026-10-18 found two candidates in Pascal's own recipes.
Pascal picks one of them, or neither.
Other everyday domains and extra formats stay in [M99](M99-parking-lot/everyday-domains.md).
This milestone confirms or replaces one candidate; it does not approve a domain family.

| | View one Markdown file | Convert one screen recording |
|---|---|---|
| Today | `just sk-read <name>` in Pascal's SKILLS_MONO repository prints a skill's raw Markdown | `video-archive convert --input ./clip.mov -o ./clip.mp4` in Pascal's dotfiles |
| Result | [M18](M18-markdown.md) renders one file in the terminal | One converted file |
| Cost | One card, since Nicely already ships the renderer | Several cards: tool checks, overwrite consent, retry safety |
| Unknown | How often Pascal opens a Markdown file in a terminal | How often Pascal converts one file by hand |

### Conventions for the everyday domains

Move only the conventions needed by the approved task into its spec.

- Start with one explicit input and one useful result.
- A writing command reports its destination through `--dry-run` and accepts `--output`.
  Validate the input and destination before effects.
  Require consent before overwrite, recheck under the destination lock, and publish through a temporary file.
- Declare the required tools, and let `ncly doctor` show their install commands.
  Wrapping a tool does not make it a mandatory package dependency.
- Run with no config file on macOS and Omarchy.
- Prove the chosen command's failure and retry behavior against actual files and counted tool calls.
  A free write can still make a whole invocation unsafe to repeat.
- A read-only command needs no record.
  A single-output task does not acquire batch, glob, progress, file-picker, or resume requirements merely by joining Nicely.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M17-T1 | Pick the first task | Pascal | — | todo |
| M17-T2 | Plan one command | agent | M17-T1 | todo |

### M17-T1 Pick the first task

Pascal picks Markdown, the recording, or neither.
With a pick, he names one real input he uses and the result it must produce.

### M17-T2 Plan one command

- **Read:** [Scope](#scope), [Make a milestone ready](README.md#make-a-milestone-ready), [D061](../north-star/decisions/D061-smallest-useful-capability.md)
- **Proves:** this file's recorded pick, and the milestone that follows from it

Record Pascal's pick and his real input here, and set M17-T1 to done.
For Markdown, write M18's spec and its one card.
For the recording, replace M18 with one command's spec and cards.
Move each rejected candidate to M99 with its revisit condition.
For neither, move both to M99, delete M18, and plan no other task.
