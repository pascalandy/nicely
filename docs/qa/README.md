# QA and UAT

This folder answers one question: how does a human accept a delivery of Nicely on a real host, with real data?

## What a UAT tests

- **An extension**: does it do its job on real input?
- **The extension template and its workflow**: each tested extension is evidence about the template. Each UAT records what the template should have prevented.
- **Discovery by a fresh agent**: does an agent without hints find the command?

A UAT is acceptance by the user. [verify-nicely](../../.agents/skills/verify-nicely/SKILL.md) proves a card, and `testdata/script/` holds the scenarios.

## Start a UAT

1. Read the assigned issue with the label `2-type:uat`.
2. Read [protocol.md](protocol.md), then the `qa.md` file of the extension repository.
3. Check the fixtures in [fixtures.md](fixtures.md) before any run.
4. Post the report in the issue, then update this folder.

## Write a UAT issue

Use the title `UAT: <what the user does>`. Never name a host in a title.
Add the label `2-type:uat` and a priority, and link the Epic without a closing keyword.
The body holds these sections: Goal, Tested release, Prerequisites, Human escalation, Protocol deltas, Prompt for the fresh agent, Report, and Done when.
Human escalation names each case where the agent calls Pascal, as [protocol.md](protocol.md#human-escalation) defines.

## Update after each UAT

A finished UAT changes at most three files:

- [ledger.md](ledger.md): one row, always
- The `qa.md` of the extension: only a new criterion, limit, or gotcha
- [protocol.md](protocol.md): only a lesson for every extension

A lesson about the template goes in the ledger and in an issue about the template.
