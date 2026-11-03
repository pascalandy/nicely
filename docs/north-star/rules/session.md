# Session and card rules

This part of the [core rule index](README.md) holds the rules of `AGENTS.md` and `RULES.md`.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| G01 | Read each file of Every session, and `RULES.md`, in each session | [AGENTS.md][agents] | Prose only |
| G02 | Start with `just next` on the latest `origin/main`, or on the top of the stack | [AGENTS.md][agents] | Prose only |
| G03 | Read the links of the card, and take rules only from routed files, never from `docs/archived/` | [AGENTS.md][agents] | Prose only |
| G04 | Run `lefthook install` once per clone | [AGENTS.md][agents] | Prose only |
| G05 | Update the file that owns a question when its subject changes | [AGENTS.md][agents] | Prose only |
| G06 | Put a new rule in the file whose question it answers, never in `AGENTS.md` | [AGENTS.md][agents], [dev-preferences.md][dev] | Prose only; [limits][limits] bound only the size of `AGENTS.md` |
| G07 | Read D067 before you add a check, and place the check as it says | [AGENTS.md][agents] | Prose only |
| G08 | One card is one pull request: work only that card, and say when an issue is outside the plan | [AGENTS.md][agents] | Prose only |
| G09 | Prepare a card that waits on Pascal, tell him, and stop when no agent card can start | [AGENTS.md][agents] | Prose only |
| G10 | Copy the checklist into the pull request, and tick each step or give its `n/a` reason | [AGENTS.md][agents] | Prose only |
| G11 | The spec covers each command, flag, JSON key, and code that a card adds or changes | [AGENTS.md][agents] | [Spec test][spec-test], partial: declared commands and flags |
| G12 | Each scenario of a card fails for the missing behavior before the code | [AGENTS.md][agents], [dev-preferences.md][dev] | Prose only |
| G13 | The code passes the scenarios of the card | [AGENTS.md][agents] | [Scenarios][script-test] of `just test` |
| G14 | `just check` passes before a card is done | [AGENTS.md][agents] | [check.sh][check], [CI][ci] |
| G15 | An agent that did not write the change verifies each feature, and the pull request quotes its evidence | [AGENTS.md][agents] | Prose only |
| G16 | The card is done in its backlog item | [AGENTS.md][agents] | Prose only; the [plan test][plan-test] checks only the format |
| G17 | `README.md` shows each change that a user sees | [AGENTS.md][agents] | Prose only |
| G18 | A skill teaches each new command | [AGENTS.md][agents] | Prose only |
| G19 | Sign off only through `just signoff` or `just signoff-layers`, never through a bare `gh signoff` | [AGENTS.md][agents] | Prose only |
| G20 | The signoff recipe checks and signs the same committed SHA | [AGENTS.md][agents], [dev-preferences.md][dev] | [Signoff test][signoff-test], [signoff][signoff] |
| G21 | Write the spec first, and fix it first when a source that it cites changes | [AGENTS.md][agents] | Prose only |
| G22 | Stop and report when one check fails three times | [AGENTS.md][agents] | Prose only |
| G23 | Follow the code map and the rules of each SDK package comment | [AGENTS.md][agents] | Prose only |
| W01 | Write code, comments, docs, and commit messages in English | [RULES.md][rules] | Prose only |
| W02 | Use placeholders for personal paths and private addresses | [RULES.md][rules] | [Gitleaks rules][gitleaks], partial: known patterns |
| W03 | Keep keys out of every log | [RULES.md][rules] | [Auth login scenario][auth-login], partial: tested outputs |
| W04 | Write each commit message as `<emoji> <type>: <scope>: <summary>` | [RULES.md][rules] | Prose only |
| W05 | Review an external pull request: check it out, compare it with its card and specs, then sign off | [RULES.md][rules] | Prose only |

[agents]: ../../../AGENTS.md
[rules]: ../../../RULES.md
[dev]: ../dev-preferences.md
[limits]: ../../../tools/cmd/plan/limits.go
[spec-test]: ../../../internal/cli/spec_test.go
[script-test]: ../../../cmd/ncly/script_test.go
[check]: ../../../scripts/check.sh
[ci]: ../../../.github/workflows/check.yml
[plan-test]: ../../../tools/cmd/plan/main_test.go
[signoff-test]: ../../../scripts/signoff-test.sh
[signoff]: ../../../sdk/extensionrepo/signoff.go
[gitleaks]: ../../../.gitleaks.toml
[auth-login]: ../../../testdata/script/auth_login.txtar
