# M6 Canadian French

Status: planned
Release: v0.6.0

## Goal

Every word `ncly` shows a human exists in Canadian French. The agent contract stays exactly the same.

## Depends on

M1. Translate the catalog as it stands when this milestone starts. Later milestones add their own `fr-CA` entries as they go.

## Scope

**Catalog.** A complete `fr-CA` catalog that passes the parity test against `en`.

**Workflow.** Repeat until the parity test passes:

1. An agent lists the keys missing from `fr-CA`.
2. It translates each one, using the description of the entry for context.
3. Pascal reviews the tone of the new entries.

**Formatting.** Plural and number rules follow the language. French uses the singular for zero, as in "0 fichier", and a decimal comma, as in "1,5 Mo".

**Detection.** `--lang fr-CA`, `NCLY_LANG=fr-CA`, or a system language such as `fr_CA.UTF-8` selects the catalog. A plain `fr` also selects `fr-CA` while it is the only French catalog.

**README.** One line says the interface speaks Canadian French and how to select it.

## Out of scope

- Other languages: [M99](M99-parking-lot.md).
- Translating the docs.

## Acceptance

```
# The system language selects French
env LANG=fr_CA.UTF-8
exec ncly --help
stdout 'Utilisation'

# The agent contract does not change
env LANG=fr_CA.UTF-8
env DEEPGRAM_API_KEY=
exits 78 ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json
stderr '"code":"AUTH_MISSING"'
stderr '"hint":"ncly auth login deepgram"'
```

A unit test of `internal/i18n` covers the plural and number rules, because no single command shows them all.

## Done when

- [ ] The parity test passes for `fr-CA`
- [ ] Pascal has reviewed every entry
- [ ] Every acceptance scenario passes in `just check`
- [ ] v0.6.0 is tagged and released
