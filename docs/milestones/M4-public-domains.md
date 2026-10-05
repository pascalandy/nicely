# M4 Public domains

Status: planned
Release: v0.4.0

## Goal

A first-time visitor gets value from `ncly` with no configuration and no key. The domains come from an inventory of the tools Pascal uses.

## Depends on

M1. M2 and M3 are not required.

## Scope

Work in this order.

1. **Inventory.** Collect the tools on Pascal's machines:
   - macOS: `brew leaves`
   - Omarchy: `pacman -Qe`
   - Both, by frequency of use: `LC_ALL=C sed 's/^: [0-9]*:[0-9]*;//' ~/.zsh_history | awk '{print $1}' | sort | uniq -c | sort -rn | head -50`

   Add a table to this file with one row per tool: the tool, its domain, its layer, which is core, extension, or drop, and one line of reason. Done when Pascal approves the table.
2. **Pick the domains.** Choose the first domains from the approved table. The candidates so far are `markdown`, `video`, and `image`.
3. **Spec.** Add each command to [cli-spec.md](../north-star/cli-spec.md) before writing Go.
4. **Build.** One domain at a time, each with its scenarios.

**Conventions for these domains.**

- The main input is the one positional argument: a file, a folder, or a glob. A folder or a glob processes a batch with one overall progress bar.
- `--output` picks the destination. `--dry-run` lists every file the command would write. An existing file is replaced only with `--force`.
- Without input in interactive mode, a file picker opens in the folder set by `input` under `[paths]` in the config.
- The help of a command that wraps a tool names it, such as "Uses ffmpeg".

**Candidates, to confirm after the inventory.**

| Command | Built on |
|---|---|
| `ncly markdown view <file>` | Glamour, imported, so no external tool |
| `ncly video convert <input> --to mp4\|webm\|gif` | `ffmpeg`, wrapped |
| `ncly image convert <input> --to jpg\|png\|webp` | An image tool chosen during this milestone |

## Out of scope

- `--explain`, the `ncly tools` showcase, and the launcher: [M99](M99-parking-lot.md).
- Any command that needs a key.

## Acceptance

```
# A dry run on a folder lists every file it would write
exec ncly video convert clips/ --to mp4 --dry-run --json
stdout 'clips/a.mp4'
stdout 'clips/b.mp4'
! exists clips/a.mp4

# An existing file is not replaced without --force
exits 2 ncly video convert clips/a.mov --to mp4 --output existing.mp4 --json
stderr '"code":"CONFIRMATION_REQUIRED"'

# Markdown renders with no config file and no external tool
env NCLY_CONFIG=$WORK/missing.toml
exec ncly markdown view notes.md --no-color
stdout 'Title'
```

On a machine with no Nicely config, each new command works after the prerequisites that `ncly doctor` names.

## Done when

- [ ] The inventory table is in this file and approved by Pascal
- [ ] Every acceptance scenario passes in `just check`
- [ ] Each new command runs with no config file on macOS and on Omarchy
- [ ] v0.4.0 is tagged and released

## Open questions

1. Should the packages depend on `ffmpeg`, or should `ncly doctor` offer to install it? Recommendation: let doctor offer it, so the packages stay light for people who never convert a video.
