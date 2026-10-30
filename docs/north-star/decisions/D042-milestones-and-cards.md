# D042 Plan work as backlog items of small cards

Decided 2026-10-09. Revised 2026-10-25 by [D065](D065-backlog-items.md), and on 2026-10-30 by Pascal for COR-103.
A backlog item demonstrates one capability, with the cards that its feature needs.
A card is one pull request.
[The backlog README](../../backlog/README.md) owns readiness, ownership, dependencies, and format.
D065 owns IDs.
`just next` selects the work; `just status` reports progress.

Keep one extension's work together, and use the backlog queue for delivery order.
[dev-preferences.md](../dev-preferences.md#merge-after-an-independent-review) owns pull-request sizing.

**Why.** The former mix of tasks and acceptance lists let agents finish isolated tasks while leaving progress unclear. Each card now has one owner and proof.

**Rejected.** Unchecked conventions and GitHub as the queue. The plan travels with the code.
