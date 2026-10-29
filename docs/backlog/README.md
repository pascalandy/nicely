# Backlog

This folder holds one file per extension or core item and the ordered delivery queue.
[D065](../north-star/decisions/D065-backlog-items.md) owns IDs, number bands, and release versions.
[D042](../north-star/decisions/D042-milestones-and-cards.md) keeps each item small and each card one pull request.

## Queue

1. COR-101
2. XT-102
3. COR-102

## Plan rules

The queue lists exactly the Planned, Ready, and Active items.
`just next` serves its first item.
That item has a GitHub milestone named after its ID and title, such as `COR-101 release checks`, which holds its open issues.
The pull request that changes the first item moves that milestone.
`just status` reads the frontmatter and prints the extension and core tables.
Parked items and ideas wait outside this queue.
Process work stays in issues and gets no item ID.

## Format

Each file is named `<ID>-<1 to 3 lowercase keywords>.md`.
Its frontmatter holds exactly three keys, in this order: `id`, `extension` or `core`, then `state`.
The first word of `state` is its status.
The rest carries versions, repositories, Epics, frozen history, Next, Later, or a revisit condition.
Quote values that contain `#`, and keep frontmatter free of comments and backslashes.

```yaml
---
id: XT-102
extension: transcript
state: "Ready. Next card is the dry-run lock fix. Shipped by M13 and M16"
---
```

Planning an idea gives it the next free 1xx ID in that plan-only pull request.
Only an item without done cards can be renumbered.

A 1xx item's body uses Demo, optional Scope, Open questions, Cards, optional After, and Later.
An idea says what it would do, and its state gives its revisit condition.
Later keeps deferred source text and its conditions.
Deferred options are not MVP acceptance criteria.
Select one small useful slice before planning an idea.
[D061](../north-star/decisions/D061-smallest-useful-capability.md) records the scope audit of PR #156.
The 16 files in [milestones](../milestones/README.md) are frozen history and are only link-checked.
Create no new M file.

## Statuses

| Status | Meaning |
|---|---|
| Idea | A 6xx item, with no cards, outside the queue |
| Parked | Chosen 1xx work outside the queue, with no card started; questions and todo cards may remain |
| Planned | In the queue, with no done card; holds open questions or todo cards |
| Ready | In the queue, without open questions; has at least one todo card and none done |
| Active | In the queue, with done and todo cards and no open questions |
| Shipped | Has no todo card or open questions |

An extension maintained in its own repository, such as XT-101, keeps its cards in that repository's Epic.

## Make an item ready

When `just next` names a Planned item with open questions, make it ready.
[dev-preferences.md](../north-star/dev-preferences.md#merge-after-an-independent-review) says how many pull requests this takes.

1. Settle each open question in the named spec, then delete Open questions.
   Ask Pascal only what is his to decide, such as a look, a cost, or a principle.
2. Finish the cards and trace each through the [code map](../north-star/architecture.md#code-map).
   Name the package and check that it may reach what the card needs.
   Reach a forbidden import through the CLI after the item that adds its command.
   Keep at most five agent cards per item, each proving one capability in one pull request.
   Each card links what to read and names its scenarios.
   A decision that only Pascal can make is a card owned by Pascal.
   Cards waiting for that decision depend on it.
   A Pascal check that blocks no card belongs under After.
3. Keep Demo to one sentence that the cards prove together.
4. Set the first word of state to Ready when the sections that the cards read are written.
   `just next` then prints the first card when this item reaches the front of the queue.

After a plan pull request merges, fix stale item and card references in each open issue labeled `4-epic`.
List the updated issues in that pull request.

## Cards

Cards have IDs such as `XT-102-T1`.
The table columns are Card, Title, Owner, Depends on, and Status.
Each card has one `### <card> <title>` section.
An agent card starts with Read links and a Proves line.
Work that matters is a card, so item files hold no checkboxes.

Each code span in Proves is a path from the repository root.
Name commands, such as just check, without code spans.
For a done card, `just test` checks that each path exists.

The card's pull request sets its table status to done.
With done and todo cards, the item becomes Active.
With no todo card, it becomes Shipped and leaves the queue.
`just test` rejects any other combination.

## Release a shipped item

[D065](../north-star/decisions/D065-backlog-items.md) owns release versions.
[D043](../north-star/decisions/D043-milestone-versions.md) owns the release triggers, and
[release-nicely](../../.agents/skills/release-nicely/SKILL.md) owns the release procedure.

## Everyday domains

- [Markdown view](XT-103-markdown-view.md)
- [Video convert](XT-101-video-convert.md)
- [Video and image jobs](XT-602-video-image-jobs.md)

## Other deferred scope

Parked from M17: exhaustive inventories of two machines, histories and packages, and the rule
that every writing command supports files, folders, globs, batches, progress, records, and
resume. A safe one-file transformation does not inherit all of those capabilities.

Writing commands still need dry run, destination validation, overwrite consent, and truthful retry behavior for their actual effects.
