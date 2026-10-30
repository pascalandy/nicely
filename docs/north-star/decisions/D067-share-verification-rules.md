# D067 Share verification rules and keep extension facts

Decided for [COR-103-T13](../../backlog/COR-103-verification-loop.md#cor-103-t13-rules-and-prose).
Core's imported SDK and test kit own shared checks.
Each extension supplies its inputs, fixtures, expected values, and independent checks.
The template copies only tool calls and forms for the author.

## Choose the owner

1. Describe the failure that the user or agent sees, with its input and host.
2. If one owner can prevent the failure, fix that owner's code.
   The SDK owns shared behavior, core owns host behavior, and the extension owns its domain.
   This branch ends the search.
3. If contract or manifest facts determine the expected value, put the evaluator in the kit.
   The extension supplies any missing declaration.
   This branch ends the search.
4. Otherwise, put the check in the extension's scenario, Go test, or `acceptance.sh`.
   Acceptance runs through kit commands and judges results with a tool other than the extension.
5. Copy a template file only for tool calls or an author form.
   Copied recipes hold no check logic.

When a second extension needs the same evaluator, move the evaluator to the kit.
Prose holds judgment that a machine cannot make.
Human UAT holds authority, perception, and consent.

## Prove the checks

- Give each kit check a fixture that fails without that check
- Run first-party extensions through the kit's checks with [core's runner adapter](../../../cmd/ncly/script_test.go)
- For each release, pass the [published-tag author path](../../../.agents/skills/release-nicely/scripts/author-path.sh)
- Before a UAT closes, turn each finding into a scenario, an acceptance check, or a known limit
- Fix each failure at the highest shared owner that can prevent it
- Add the correction's rule and enforcer to the [template rule index](../../../template/AGENTS.md.jinja#rule-index)

The template index owns the extension rules.
Reviewers approve each `untested` reason and its linked issue before merge.
The UAT report lists these exemptions.
A tag without a release receipt is a UAT finding.

**Why.** Imported checks reach extensions with their pinned SDK.
Copied check logic needs merges with author edits and can drift.
A rule index sends agents to the check that decides the result.

**Rejected.** More check logic in copied recipes or repeated prose.
Core checks that know each extension's domain inputs.
A separate verification skill for each extension when `acceptance.sh` can drive its commands.
