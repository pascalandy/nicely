---
id: COR-102
core: "Canadian French"
state: "Ready. Three agent cards from M25. Later: more languages, catalog folders"
---

# COR-102 Canadian French

## Demo

`ncly` and its shipped first-party extensions speak Canadian French, while the agent contract stays exactly the same.

## Scope

Translate core's catalog, the SDK's catalog, and the catalogs of first-party extensions that have actually shipped when this item starts.
Parked commands and specifications require no translation.
Keep the existing catalog parity requirement unchanged.
Later items add their own `fr-CA` entries as they go.
Each command's scenarios have run once under the pseudo-locale since M00, so most missing strings surfaced in the milestone that added them.

Repeat until the parity test passes: an agent lists the keys missing from `fr-CA`, then translates each one, using the description of the entry for context.

The plural rules and formats of the i18n package apply. French uses the singular for zero, as in "0 fichier", and a decimal comma, as in "1,5 Mo". Machine fields keep their types, units, enum values, and defined ordering in every locale. A duration such as `duration_ms` stays a number of milliseconds. Run IDs, step statuses, and retry decisions never depend on a translated string.

The language matching of M00 selects the new catalog for `fr_CA.UTF-8` and for a plain `fr`, as [Configuration](../north-star/contract/configuration.md) says, so detection needs no new code. The docs stay in English, and other languages wait in [Later](#later).

```
# The system language selects French
env LC_ALL=
env LC_MESSAGES=
env NCLY_LANG=
env NCLY_CONFIG=$WORK/missing.toml
env LANG=fr_CA.UTF-8
exec ncly --help
stdout 'UTILISATION'

# The agent contract does not change
exits 2 ncly auth logout deepgram --json
stderr '"code":"NOT_FOUND"'
```

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-102-T1 | Translate core | agent | — | todo |
| COR-102-T2 | Translate shipped first-party extensions | agent | COR-102-T1 | todo |
| COR-102-T3 | Formats and README | agent | COR-102-T1 | todo |

### COR-102-T1 Translate core

- **Read:** [Configuration](../north-star/contract/configuration.md)
- **Proves:** `testdata/script/language_fr.txtar`

A complete `fr-CA` catalog for core and for the SDK, which renders the help pages, the global flags, and the error labels of every command, each passing the parity test against `en`. The system language selects them.
A confirmation shows `[o/N]` and accepts `o` and `oui`.
Each templated message keeps its fields, flags, paths, and placeholders.
Assert a `fr-CA.toml` for each owner, since parity passes without one.
`TestMatch` expects `fr-CA`, and `extension_help.txtar` changes where French help expects English completion text.

### COR-102-T2 Translate shipped first-party extensions

- **Read:** [Extensions](../north-star/contract/extensions.md)
- **Proves:** `testdata/script/language_fr_extensions.txtar`

A complete `fr-CA` catalog for each shipped first-party extension.
Under `en`, `en-XA`, and `fr-CA`, one call gives the same exit, `ok`, codes, `next`, numbers, enums, and run IDs.

### COR-102-T3 Formats and README

- **Read:** [Configuration](../north-star/contract/configuration.md)
- **Proves:** the plural and number cases in the i18n package's unit test

A unit test covers the French plural and number rules, because no single command shows them all. README.md says in one line that the interface speaks Canadian French and how to select it.

## After

Pascal reviews the tone of every new `fr-CA` entry. An agent applies any changes he requests.

## Later

### More languages

Spanish, Japanese, and others, through the COR-102 workflow.

Revisit condition: A user asks, or `fr-CA` proves the workflow.

### Catalog folders

One folder per language, such as `locales/en/`, lets a catalog split under the size limit. [Text of an extension](../north-star/contract/text-of-an-extension.md) asks for one file per language, so `just limits` skips catalogs.

Revisit condition: A catalog grows too large to read whole. The change needs a contract decision.
