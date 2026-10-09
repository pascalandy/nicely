# M07 Forms

Status: planned
Version: v0.7.0

## Demo

A human logs in through a masked field, confirms a replacement in a form, and reads a skill rendered as Markdown, in the looks that Pascal picked.

## Scope

Every look goes through a published mockup first, as [dev-preferences.md](../north-star/dev-preferences.md#decide-the-look-with-a-published-mockup) asks. One mockup page covers every look that the plan adds up to M16, so Pascal picks once: the `skill list` and `skill view` output, the masked key field, the confirmation, `auth status`, the doctor report, `agent profile list`, `run list` and `run view`, `tap list`, and the transcript spinner. M09, M10, M12, and M16 build their looks from the same pick. Until then, commands reuse the help and error styles of M00 and print the rest as plain text, as [AGENTS.md](../../AGENTS.md#code) says.

The first form needs Huh, so it brings the route around the `tmux info` call that Lip Gloss makes at load, as [D001](../north-star/decision-records.md#d001-build-from-scratch-in-go-with-cobra-and-charm) defines. `terminal.txtar` proves that no package asks the terminal anything at load.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M07-T1 | Pick the looks | Pascal | — | todo |
| M07-T2 | Lip Gloss route | agent | — | todo |
| M07-T3 | Login form and confirmation | agent | M07-T1, M07-T2 | todo |
| M07-T4 | Styled reports and skill view | agent | M07-T1, M07-T2 | todo |

### M07-T1 Pick the looks

An agent builds and publishes the mockup page that the scope lists, as dev-preferences.md asks, and gives Pascal its URL. Pascal picks one look for each item. The agent records the picks in the spec of each command, then sets this card to done.

### M07-T2 Lip Gloss route

- **Read:** [D001](../north-star/decision-records.md#d001-build-from-scratch-in-go-with-cobra-and-charm)
- **Proves:** `testdata/script/terminal.txtar`

Use a Lip Gloss version that no longer probes the terminal at load: an upstream release, or, after Pascal approves it, a fork wired through a `replace` line in `go.mod`. `terminal.txtar` proves that loading Huh, a Bubbles spinner, and Glamour asks the terminal nothing.

### M07-T3 Login form and confirmation

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [Modes](../north-star/contract.md#modes)
- **Proves:** `testdata/script/auth_form.txtar`

In interactive mode, `ncly auth login` asks for the key in a masked field, and a replacement or a logout asks for confirmation, in the picked looks. After the form, `ncly` prints the equivalent command after `Next time:`.

### M07-T4 Styled reports and skill view

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Output](../north-star/contract.md#output)
- **Proves:** `testdata/script/styles.txtar`

`skill list`, `auth status`, and the doctor report use the picked styles on a terminal, and plain text elsewhere. `skill view` renders its Markdown through Glamour in interactive mode.
