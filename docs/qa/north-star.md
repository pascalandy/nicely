# QA north star

This file answers one question: what does every UAT of Nicely and its extensions aim for? Only Pascal changes it.

Each UAT judges two axes, in addition to the other [criteria families](protocol.md#criteria-families). Each axis is a family of that table and a section of the [report](protocol.md#report).

## Automation

An agent proves each acceptance criterion alone, with a test as code, such as `just check`, a scenario in `testdata/script/`, or [verify-nicely](../../.agents/skills/verify-nicely/SKILL.md).

The report names:

- Each manual step that a test as code could replace
- Each acceptance criterion that no test as code covers
- Each step that needs a human outside a case of [Human escalation](protocol.md#human-escalation)
- The status of the latest commit on `main`, from `gh api repos/<owner>/<repo>/commits/main/status`

In Nicely, that status is `signoff`, which `just signoff` posts. A repository without a status is a finding.

## Economy

The tester uses the CLI as a human and as an agent, and looks for edge cases. The CLI stays light and fast, and costs few tokens to read.

- The default output is short and gives what the common case needs
- The `--json` answer is short, and gives an agent what it needs without a second call
- An error is short, names its fix, and has the correct exit code
- The default value of each flag suits the common case
- The [nicely skill](../../extensions/skill/nicely/SKILL.md) and the skill of the extension are short, agree with the current CLI, and have current examples

When possible, the report gives the time and the bytes of output of each common command. Measure stdout and stderr apart, with `--json` and without it. The common commands are `ncly skill view nicely`, `ncly describe`, and the dry run and the run of the extension.
