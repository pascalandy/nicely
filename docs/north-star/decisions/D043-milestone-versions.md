# D043 Release each accepted backlog item

Decided 2026-10-09. Revised 2026-10-19 in [#272](https://github.com/pascalandy/nicely/issues/272).
Revised 2026-10-25 by [D065](D065-backlog-items.md), which owns version numbering.
Release after each completed item that Pascal accepts, without a separate release request.
Acceptance requires Pascal's explicit `accept <ID>`, such as `accept COR-102`, or equivalent wording for that item.
He accepts it in conversation or on its Epic; a Shipped item alone is not accepted.
Follow [release-nicely](../../../.agents/skills/release-nicely/SKILL.md) for changelog, checks, tag, publication, and verification.
Releases remain outside the queue and never block the next item.

**Why.** Waiting for M00's release blocked the next milestone. Separate releases remove that dependency.
Releasing each accepted item keeps the published version current.

**Rejected.** Release gates, a separate release request, and versions derived from plan IDs.
`0.1.<n>` numbering leaves no patch number between releases.
