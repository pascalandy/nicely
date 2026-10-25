# D065 Plan work by extension and core backlog items

Decided 2026-10-25 by Pascal.
The plan moves to `docs/backlog/`, with one file per item and one queue in its README.
`XT-nnn` identifies extension work, and `COR-nnn` identifies core work.
Process work, such as limits, signoff, the plan tool, and Epics 12 and 13, stays in issues without an item ID.

## Identity and state

Only two number bands exist.
The 1xx band holds work chosen to build, whatever its status.
The 6xx band holds ideas not chosen.
Planning an idea assigns the next free 1xx ID in that pull request.
Only an item without done cards may be renumbered.

A file is named `<ID>-<1 to 3 lowercase keywords>.md`, such as `XT-102-transcript.md`.
It opens with exactly three frontmatter keys, each on one line, in this order.

```yaml
id: XT-102
extension: transcript
state: "Ready. Next card is the dry-run lock fix. Shipped by M13 and M16"
```

A core item uses `core` instead of `extension`.
Quote values that contain `#`.
Frontmatter carries no comments.
The first word of `state` is its status.
The remaining text carries versions, repository, Epic, history, next work, and later work.

| Status | Meaning |
|---|---|
| Idea | A 6xx item with no cards. Its state gives the revisit condition |
| Parked | A 1xx item outside the queue. It may keep open questions and todo cards, with none started |
| Planned | In the queue, with open questions or todo cards, and no done card |
| Ready | In the queue, with at least one todo card, no done card, and no open questions |
| Active | In the queue, with at least one done card and one todo card, and no open questions |
| Shipped | Outside the queue, with no todo card or open questions |

The queue lists exactly the Planned, Ready, and Active items, in delivery order.
`just next` serves its first item.
`just status` prints separate XT and COR tables from frontmatter.

## Body and cards

The sections are Demo, optional Scope, Open questions, Cards, optional After, and optional Later.
Later preserves parked source text, while state keeps its short list.
Status and version have no separate body lines.
A 1xx item with cards has a one-sentence Demo and at most five agent cards.
A card is one pull request, named `<item ID>-T<n>`, such as `XT-102-T1`.
The card table and AGENTS.md checklist keep their shape.
An external extension runs its work through its own Epic and repository, as Epic 14 did.
Its Nicely item holds no cards.

## History and releases

M00 to M10, M13, M16, M17, M18, and M23 stay as frozen history.
Their text changes only to repoint links.
The plan tool checks their links and never reads their status, version, order, or cards as the plan.
No new M file is allowed.
The milestones README becomes a note pointing to the backlog.

Each accepted item released from Nicely takes the next minor version, starting with v0.24.0.
Patch versions cover intervening fixes.
Pascal explicitly accepts an item with `accept <ID>`, such as `accept COR-102`, or equivalent wording for that item.
A Shipped item alone is not accepted.
An external extension, such as XT-101, keeps its own versions.
Releases remain outside the queue and do not block the next item.

**Revises.** [D042](D042-milestones-and-cards.md) keeps small cards and moves numbering and readiness to backlog items.
[D043](D043-milestone-versions.md) keeps explicit acceptance and replaces milestone versions with the next minor per accepted item.

**Why.** Ownership stays visible as an extension evolves across releases.
The queue expresses delivery order without renumbering unrelated capabilities.
Ideas retain their conditions without becoming promised work.

**Rejected.** New milestones, a separate parking lot, and versions derived from item numbers.
