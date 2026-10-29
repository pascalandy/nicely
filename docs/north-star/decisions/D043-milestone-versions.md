# D043 Release each shipped backlog item

Decided 2026-10-09. Revised 2026-10-19 in [#272](https://github.com/pascalandy/nicely/issues/272), and in [#355](https://github.com/pascalandy/nicely/pull/355) on 2026-10-28.
Revised 2026-10-25 by [D065](D065-backlog-items.md), which owns version numbering.
Release when an item becomes Shipped on origin/main, without a request or an acceptance.
Release a patch when a merged fix changes a behavior of the latest release.
When one merge does both, release only the next minor version.
The agent that merges the last card of the item, or the fix, runs the release.
Follow [release-nicely](../../../.agents/skills/release-nicely/SKILL.md) for changelog, checks, tag, publication, and verification.
Releases remain outside the queue and never block the next item.

**Why.** Waiting for M00's release blocked the next milestone. Separate releases remove that dependency.
Releasing each shipped item keeps the published version current.
Independent verification and `just signoff` already prove each card before it merges.

**Rejected.** Release gates, a separate release request, and versions derived from plan IDs.
`0.1.<n>` numbering leaves no patch number between releases.
Pascal's explicit `accept <ID>` before each release, from 2026-10-19 to 2026-10-28, delayed releases and added no check that the cards lack.
