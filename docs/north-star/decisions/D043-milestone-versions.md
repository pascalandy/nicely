# D043 Release each accepted milestone with its milestone version

Decided 2026-10-09. Revised 2026-10-19 in [#272](https://github.com/pascalandy/nicely/issues/272).
Use `v0.<n>.0` for `M<n>`, v0.0.1 for M00, and patch versions for intervening fixes.
Release after each completed milestone that Pascal accepts, without a separate release request.
Follow [release-nicely](../../../.agents/skills/release-nicely/SKILL.md) for signoff, authorized merge, changelog, checks, tag, publication, and verification.
Releases remain outside the card queue and never block the next milestone.

**Why.** Waiting for M00's release blocked the next milestone. Separate releases remove that dependency.
Releasing each accepted milestone keeps the published version current.

**Rejected.** Release gates, a separate release request, and `0.1.<n>` numbering, which leaves no patch number between milestones.
