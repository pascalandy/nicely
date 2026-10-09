# M17 Inventory

Status: planned
Version: v0.17.0

## Demo

Pascal approves a table of the tools he uses, each with the domain that would carry it, and the milestones of the first everyday domains follow from it.

## Scope

A first-time visitor should get value from Nicely with no configuration and no key, so a command that needs a key belongs to no everyday domain. The everyday domains come from an inventory of the tools Pascal uses. [M18](M18-markdown.md), [M19](M19-video.md), and [M20](M20-image.md) hold the current candidates, and this milestone confirms, replaces, or adds them.

Collect the tools on Pascal's machines:

- macOS: `brew leaves`, and by frequency of use in zsh: `LC_ALL=C sed 's/^: [0-9]*:[0-9]*;//' ~/.zsh_history | awk '{print $1}' | sort | uniq -c | sort -rn | head -50`
- Omarchy: `pacman -Qe`, and by frequency of use in bash: `awk '{print $1}' ~/.bash_history | sort | uniq -c | sort -rn | head -50`

The table has one row per tool: the tool, its domain, where it goes, which is a first-party extension, Pascal's private tap, or drop, and one line of reason.

### Conventions for the everyday domains

Move these into each domain's spec when its milestone becomes ready.

- The main input is the one positional argument: a file, a folder, or a glob. A folder or a glob processes a batch with one overall progress bar, and `--json` answers with `results`, as [Output](../north-star/cli-spec.md#output) defines
- `--output` picks the destination. `--dry-run` lists every file the command would write
- Preparation expands and orders inputs once, detects duplicate destinations and input/output collisions, and validates the whole batch before effects. Execution revalidates destinations under their resource locks
- Publish a completed artifact from a temporary file, preserve successful items after another item fails, and record reusable partial work through the SDK's records. Resume verifies artifacts before skipping them. The overall exit follows Retry safety across all items, even though these commands make no paid request
- Without input in interactive mode, a file picker opens in the folder set by `input` under `[paths]` in the config
- A wrapped tool, such as `ffmpeg`, is no package dependency. `ncly doctor` offers to install it, and the AUR package lists it in `optdepends`. Homebrew discourages optional dependencies, so the formula leaves it out
- Each command runs with no config file on macOS and on Omarchy
- Each domain also verifies an interrupted batch, changed destinations after dry run, and a repeat invocation, comparing actual files and tool calls. A free conversion that already overwrote a non-repeatable destination must not report 75 merely because its later error was temporary

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M17-T1 | Collect the inventory | agent | — | todo |
| M17-T2 | Approve the table | Pascal | M17-T1 | todo |
| M17-T3 | Plan the domains | agent | M17-T2 | todo |

### M17-T1 Collect the inventory

- **Read:** [Scope](#scope)
- **Proves:** the table in this file

Run the commands above on both machines and add the table to this file.

### M17-T2 Approve the table

Pascal reviews the table and approves it, with any changes he asks for.

### M17-T3 Plan the domains

- **Read:** [Scope](#scope), [AGENTS.md](../../AGENTS.md#make-a-milestone-ready)
- **Proves:** `docs/milestones/`, with the milestones of the approved domains

Confirm or replace M18 to M20 from the approved table. Give a further domain the number after M20 and renumber the milestones that follow, as [AGENTS.md](../../AGENTS.md#make-a-milestone-ready) allows.
