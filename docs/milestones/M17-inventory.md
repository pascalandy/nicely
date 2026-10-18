# M17 First everyday task

Status: planned
Version: v0.17.0

## Demo

Pascal chooses one recurring task, supported by concrete recipes, as Nicely's first useful command requiring no key or configuration.

## Scope

Review a small sample of Pascal's actual recurring recipes.
Show enough context to choose one useful task: today's command, usual input, expected result, and current friction.
Use relevant `justfile` recipes or skills when a shell command hides the actual work.
A full package inventory, both machines' shell histories, and a catalogue of future domains are unnecessary.

[Markdown viewing](M18-markdown.md) is the current candidate.
Other everyday domains and extra formats stay in [M99](M99-parking-lot/everyday-domains.md).
This milestone confirms or replaces that one candidate; it does not approve a domain family.

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
| M17-T1 | Inspect concrete recipes | agent | — | todo |
| M17-T2 | Choose the first task | Pascal | M17-T1 | todo |
| M17-T3 | Plan one command | agent | M17-T2 | todo |

### M17-T1 Inspect concrete recipes

- **Read:** [Scope](#scope)
- **Proves:** the short task comparison in this file

Read enough actual recipes to support a choice.
Describe their inputs, results, and friction, then recommend one task.
Stop when the evidence supports that decision.

### M17-T2 Choose the first task

Pascal chooses the first task from the evidence and confirms the result it must produce.

### M17-T3 Plan one command

- **Read:** [Scope](#scope), [Make a milestone ready](README.md#make-a-milestone-ready)
- **Proves:** the one candidate milestone and its command spec

Confirm or replace M18 with the approved task.
Write one command's scope and cards, without scheduling the rejected alternatives.
