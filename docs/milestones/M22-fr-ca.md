# M22 Canadian French

Status: planned
Version: v0.22.0

## Demo

`ncly` and its first-party extensions speak Canadian French, while the agent contract stays exactly the same.

## Scope

Translate core's catalog and the catalog of each first-party extension as they stand when this milestone starts. Later milestones add their own `fr-CA` entries as they go. Each command's scenarios have run once under the pseudo-locale since M00, so most missing strings surfaced in the milestone that added them. A string that no scenario reaches shows up as English during Pascal's review.

Repeat until the parity test passes: an agent lists the keys missing from `fr-CA`, then translates each one, using the description of the entry for context.

The plural rules and formats of the i18n package apply. French uses the singular for zero, as in "0 fichier", and a decimal comma, as in "1,5 Mo". Machine fields keep their types, units, enum values, and defined ordering in every locale. A duration such as `duration_ms` stays a number of milliseconds. Run IDs, step statuses, and retry decisions never depend on a translated string.

The language matching of M00 selects the new catalog for `fr_CA.UTF-8` and for a plain `fr`, as [Configuration](../north-star/cli-spec.md#configuration) says, so detection needs no new code.

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

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M22-T1 | Translate core | agent | — | todo |
| M22-T2 | Translate the first-party extensions | agent | M22-T1 | todo |
| M22-T3 | Formats and README | agent | M22-T1 | todo |
| M22-T4 | Review the tone | Pascal | M22-T1, M22-T2 | todo |

### M22-T1 Translate core

- **Read:** [Configuration](../north-star/cli-spec.md#configuration)
- **Proves:** `testdata/script/language_fr.txtar`

A complete `fr-CA` catalog for core that passes the parity test against `en`. The system language selects it.

### M22-T2 Translate the first-party extensions

- **Read:** [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/language_fr_extensions.txtar`

A complete `fr-CA` catalog for each first-party extension. The agent contract stays the same under English, the pseudo-locale, and French.

### M22-T3 Formats and README

- **Read:** [Configuration](../north-star/cli-spec.md#configuration)
- **Proves:** the plural and number cases in the i18n package's unit test

A unit test covers the French plural and number rules, because no single command shows them all. README.md says in one line that the interface speaks Canadian French and how to select it.

### M22-T4 Review the tone

Pascal reviews the tone of every new entry, and an agent applies his changes.
