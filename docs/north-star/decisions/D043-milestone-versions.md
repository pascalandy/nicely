# D043 Number milestones by the release that ships them, and release separately

Decided 2026-10-09. Use `v0.<n>.0` for `M<n>`, v0.0.1 for M00, and patch versions for intervening fixes. Release only on Pascal's request, outside the card queue, as the [release-nicely](../../../.agents/skills/release-nicely/SKILL.md) skill says.

**Why.** Waiting for M00's release blocked the next milestone. Separate releases remove that dependency.

**Rejected.** Release gates and `0.1.<n>` numbering, which leaves no patch number between milestones.
