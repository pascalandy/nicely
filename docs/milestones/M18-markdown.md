# M18 Markdown

Status: planned
Version: v0.18.0

## Demo

`ncly markdown view notes.md` renders a Markdown file with no config file and no external tool.

## Scope

A candidate everyday domain, to confirm in [M17](M17-inventory.md). `ncly markdown view <file>` imports Glamour, so it needs no external tool. It follows the [conventions for the everyday domains](M17-inventory.md#conventions-for-the-everyday-domains), except those for batches and records: as a viewer, it writes nothing, keeps no record, reads stdin only for `-`, and answers text even under `--json`. It runs no program and fetches nothing that a document names.

```
# Markdown renders with no config file and no external tool
env NCLY_CONFIG=$WORK/missing.toml
exec ncly markdown view notes.md --no-color
stdout 'Title'
```

## Open questions

- The spec of each command, in `extensions/markdown/spec.md`, and the cards

## Cards
