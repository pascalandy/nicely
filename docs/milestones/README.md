# Milestones

This folder holds the plan: one file per milestone, named like `M01-short-title.md`, and [M99](M99-parking-lot.md), the parking lot. M99 keeps each parked capability in its own file in `M99-parking-lot/`. [D042](../north-star/decisions/D042-milestones-and-cards.md) sets the rules of the plan.

## Make a milestone ready

When `just next` names a planned milestone with open questions, make it ready. [dev-preferences.md](../north-star/dev-preferences.md#merge-after-an-independent-review) says how many pull requests this takes.

1. Settle each open question in the spec that it names, then delete the **Open questions** section. Ask Pascal only what is his to decide, such as a look, a cost, or a principle.
2. Finish the cards. Trace each card through the [code map](../north-star/architecture.md#code-map): name the package that it changes, and check that this package may reach what the card needs. A part that the package may not import, such as the host's discovery for an extension, is reached through the CLI, after the milestone that adds its command. A card changes one capability, in one pull request that a reviewer reads in one sitting. It links what to read and names the scenarios that prove it. Split a milestone of more than five agent cards into two milestones. A decision that only Pascal can make, such as a pick, is a card whose owner is `Pascal`, and the cards that wait for it depend on it. A step that waits on Pascal while no card waits on it, such as a check on his Mac, goes under **After this milestone** instead, so it never blocks the next milestone.
3. Keep the demo to one sentence that the cards prove together.
4. Set the status to `ready` once the sections that its cards read are written: in contract.md, in core-spec.md, or in the extension's spec. `just next` then prints the first card.

A planned milestone without open questions needs no separate pull request: `just next` prints its first card, and that card's pull request also sets the milestone to `active`.

After a plan pull request merges, fix the stale card and milestone references of each open issue labeled `4-epic`. List the updated issues in that pull request.

## Format

A milestone file holds `Status`, `Version`, and the sections **Demo**, an optional **Scope**, **Open questions** while it is planned, **Cards**, and an optional **After this milestone**. **Cards** holds a table with the columns Card, Title, Owner, Depends on, and Status, then one `### <card> <title>` section per card. An agent card's section starts with a **Read** line that links what to read and a **Proves** line that names its proof. No milestone file holds a checkbox: work that matters is a card. A new milestone takes the number after the last one. Pascal may give it a free number before a planned milestone, so that it runs first. `just test` checks this format, including the links.

Each code span of a **Proves** line is a path from the repository root, such as `extensions/transcript/python/` or `justfile`. Name a command, such as just check, without a code span. Once the card is `done`, `just test` checks that each path exists.

## Close a card

The pull request of a card sets the card's status to `done` in the cards table. The milestone becomes `active` with its first done card and `done` with its last, and `just test` rejects any other combination.
