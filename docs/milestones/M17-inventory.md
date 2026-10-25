# M17 First everyday task

Status: done
Version: v0.17.0

## Demo

Pascal records whether either candidate should become Nicely's first everyday command.

## Scope

The grooming of 2026-10-18 found two candidates in Pascal's own recipes.
On 2026-10-19, Pascal picked neither Markdown view nor screen-recording conversion.
No input or result was selected.
Both candidates, other everyday domains, and extra formats stay in [backlog](../backlog/README.md#everyday-domains).
The next milestone is [M18 Extension template](M18-extension-template.md).

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
| M17-T1 | Pick the first task | Pascal | — | done |
| M17-T2 | Plan one command | agent | M17-T1 | done |

### M17-T1 Pick the first task

Pascal picked neither candidate on 2026-10-19, so this card selects no everyday command.

### M17-T2 Plan one command

- **Read:** [Scope](#scope), [Make an item ready](../backlog/README.md#make-an-item-ready), [D061](../north-star/decisions/D061-smallest-useful-capability.md)
- **Proves:** this file's recorded pick, and the milestone that follows from it

The pick is recorded in [Scope](#scope).
Both candidates moved to [backlog](../backlog/README.md#everyday-domains), each with its revisit condition.
The Markdown milestone was deleted, and no other everyday task was planned.
