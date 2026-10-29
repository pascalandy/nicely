# QA north star

This file answers one question: what does every UAT of Nicely and its extensions aim for? Only Pascal changes it.

Each UAT judges two axes, in addition to the other [criteria families](protocol.md#criteria-families). Each axis is a family of that table and a section of the [report](protocol.md#report).

## Automation

An agent proves each acceptance criterion alone, with a test as code, such as `just check`, a scenario in `testdata/script/`, or [verify-nicely](../../.agents/skills/verify-nicely/SKILL.md).

The report names:

- Each manual step that a test as code could replace
- Each acceptance criterion that no test as code covers
- Each step that needs a human outside a case of [Human escalation](protocol.md#human-escalation)
- The checks of the latest commit on `main`: its statuses from `gh api repos/<owner>/<repo>/commits/main/status`, and its check runs from `gh api repos/<owner>/<repo>/commits/main/check-runs`

In Nicely, the status is `signoff`, which `just signoff` posts, and GitHub Actions gives the check runs. A `total_count` of 0 means that the commit has none of that kind. The status endpoint then still answers `"state": "pending"`, which is no work in progress. A repository with neither kind is a finding.

## Economy

The tester uses the CLI as a human and as an agent, and looks for edge cases. The CLI stays light and fast, and costs few tokens to read.

- The default output is short and gives what the common case needs
- The `--json` answer is short, and gives an agent what it needs without a second call
- An error is short, names its fix, and has the correct exit code
- The default value of each flag suits the common case
- The [nicely skill](../../extensions/skill/nicely/SKILL.md) and the skill of the extension are short, agree with the current CLI, and have current examples

When possible, the report gives the time and the bytes of output of each common command. Measure stdout and stderr apart, with `--json` and without it. The unit is the byte. A report can add tokens, counted with tiktoken and its `o200k_base` encoding.

### Budgets

Each common command has a budget: a size that its output must not exceed, with `--json` and without it. The budgets hold for core and the extension under test. The [ledger](ledger.md#economy-measurements) keeps the measurements, with the reference row that set these budgets.

| Common command | Stream | Budget in bytes |
|---|---|---|
| `ncly skill view nicely` | stdout | 2,500 |
| `ncly skill view <slug>` | stdout | 1,600 |
| `ncly describe` | stdout | 2,500 |
| `ncly describe <cmd>` | stdout | 4,500 |
| `ncly doctor <slug>` | stdout | 500 |
| The dry run of the extension | stdout | 1,000 |
| The run of the extension | stdout | 1,000 |
| An error of one of these commands | stderr | 500 |
