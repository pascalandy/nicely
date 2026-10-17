# M07 Forms

Status: active
Version: v0.7.0

## Demo

A human fills a missing value in a form, logs in through a masked field, confirms a replacement, and reads a skill rendered as Markdown, in the looks that Pascal picked.

## Scope

Every look goes through a published mockup first, as [dev-preferences.md](../north-star/dev-preferences.md#decide-the-look-with-a-published-mockup) asks. One mockup page covers every look that the plan adds up to M16, so Pascal picks once: the form for missing values, the masked key field, the confirmation, including the grant of a key in [M08](M08-grants.md), the `skill list` and `skill view` output, `auth status` with the grants of M08, the doctor report, `agent profile list`, `run list` and `run view`, `tap list`, and the transcript spinner. M08, M09, M10, M12, and M16 build their looks from the same pick. Until then, commands reuse the help and error styles of M00 and print the rest as plain text, as [AGENTS.md](../../AGENTS.md#code) says.

The first form needs Huh, so M07-T2 brings, ahead of it, the route around the `tmux info` call that Lip Gloss makes at load, through the fork that [D001](../north-star/decisions/D001-go-cobra-charm.md) approves until upstream ships the fix. `terminal.txtar` proves that no package asks the terminal anything at load. Lip Gloss and Bubble Tea still detect the color profile at run time, which runs `tmux info`: Lip Gloss on its first print, and a Bubble Tea program as it starts, unless its options give it a profile. So each form and screen takes its profile from sdk/tui, as [Output](../north-star/contract.md#output) requires. Go applies the `replace` line of go.mod only when nicely is the main module, so a Go extension built in another module against sdk/ takes Lip Gloss from its own go.mod, and needs the same line until upstream ships the fix.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M07-T1 | Pick the looks | Pascal | — | done |
| M07-T2 | Lip Gloss route | agent | — | done |
| M07-T3 | Login form and confirmation | agent | M07-T1, M07-T2 | done |
| M07-T4 | Styled reports and skill view | agent | M07-T1, M07-T2 | todo |
| M07-T5 | Form for missing values | agent | M07-T3 | todo |

### M07-T1 Pick the looks

An agent builds and publishes the mockup page that the scope lists, as dev-preferences.md asks, and gives Pascal its URL. Pascal picks one look for each item. The agent records the picks in the spec of each command, then sets this card to done.

### M07-T2 Lip Gloss route

- **Read:** [D001](../north-star/decisions/D001-go-cobra-charm.md), [Output](../north-star/contract.md#output)
- **Proves:** `testdata/script/terminal.txtar`

Use a Lip Gloss version that no longer probes the terminal at load. When an upstream release fixes [lipgloss#749](https://github.com/charmbracelet/lipgloss/issues/749), use it. Otherwise fork `charmbracelet/lipgloss` to `pascalandy/lipgloss`, apply the fix of an upstream pull request, such as [#759](https://github.com/charmbracelet/lipgloss/pull/759) or [#760](https://github.com/charmbracelet/lipgloss/pull/760), and wire the fork through a `replace` line in `go.mod`, with a comment that links #749, so the line goes once upstream ships the fix. `terminal.txtar` proves, with its stub `tmux`, that loading Huh, a Bubbles spinner, and Glamour asks the terminal nothing.

### M07-T3 Login form and confirmation

- **Read:** [ncly auth](../north-star/core-spec.md#ncly-auth), [Modes](../north-star/contract.md#modes)
- **Proves:** `testdata/script/auth_form.txtar`

In interactive mode, `ncly auth login` asks for the key in a masked field, and a replacement or a logout asks for confirmation, in the picked looks, in place of the plain prompts of [M06](M06-auth.md). Neither prints a `Next time:` line, as [Modes](../north-star/contract.md#modes) defines. Under `pty`, `auth_form.txtar` proves that the typed key never shows, that a declined confirmation changes nothing and exits 2 with `CONFIRMATION_REQUIRED`, and that Ctrl-C at either form exits 130 and leaves the terminal as it found it.

### M07-T4 Styled reports and skill view

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [ncly auth](../north-star/core-spec.md#ncly-auth), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [Output](../north-star/contract.md#output)
- **Proves:** `testdata/script/styles.txtar`

`skill list`, `auth status`, and the doctor report use the picked styles on a terminal, and plain text elsewhere. `skill view` renders its Markdown through Glamour in interactive mode.

### M07-T5 Form for missing values

- **Read:** [Modes](../north-star/contract.md#modes), [Usage](../north-star/contract.md#usage), [Dispatch](../north-star/contract.md#dispatch), [Completion values](../north-star/contract.md#completion-values), [ncly describe](../north-star/core-spec.md#ncly-describe), [D009](../north-star/decisions/D009-ask-in-a-terminal.md)
- **Proves:** `testdata/script/form.txtar`

Add the form for missing values of [Modes](../north-star/contract.md#modes) to `sdk/command`, built from a form of `sdk/tui` in the picked look, so core's commands, the bundled extension, and a Go extension started directly share one form. Core shows it for an external extension's command before it starts the extension. Under `pty`, `form.txtar` covers:

- A fixture command without its required argument, whose form asks for it, then runs the command with the typed value and prints `Next time:` with the equivalent command on stderr
- A required flag, a flag with `values`, which the form offers as a choice, and a flag group, which the form asks to pick from first
- `ncly auth login` without `<service>`, whose field suggests the declared services, as completion does, then asks for the key in the masked field of M07-T3, and whose `Next time:` line holds no key
- The same calls in non-interactive mode, which exit 2 with `USAGE_INVALID` and a hint that shows the full command, and show no form
- Ctrl-C at the form, which exits 130 before an execution marker of the fixture appears

The scenario runs once under the pseudo-locale.
