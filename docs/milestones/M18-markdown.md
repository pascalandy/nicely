# M18 Markdown

Status: planned
Version: v0.18.0

## Demo

`ncly markdown view notes.md` renders a Markdown file with no config file and no external tool.

## Scope

A candidate everyday command, awaiting Pascal's choice in [M17](M17-inventory.md).
`ncly markdown view <file>` imports Glamour, so it needs no external tool.
It follows the relevant [everyday conventions](M17-inventory.md#conventions-for-the-everyday-domains).
It writes nothing, keeps no record, reads stdin only for `-`, and answers text even under `--json`.
It runs no program and fetches nothing that a document names.
It renders through `sdk/tui.Markdown` directly, since the handler of `skill view` returns raw Markdown outside a terminal.
It wraps at the terminal width only in a terminal, which extends the one wrapping exception of the contract's Output part.
A missing input file is `NOT_FOUND` with exit 2, and an unreadable one is `RUNTIME` with exit 1.

```
# Markdown renders with no config file and no external tool
env NCLY_CONFIG=$WORK/missing.toml
exec ncly markdown view notes.md --no-color
stdout 'TITLE'
```

## Open questions

- Pascal's choice in M17, then the spec of this one command in `extensions/markdown/spec.md` and its cards

## Cards
