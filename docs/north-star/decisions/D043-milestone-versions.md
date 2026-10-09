# D043 Number milestones by the release that ships them, and release separately

Decided 2026-10-09. Milestone `M<n>` ships in `v0.<n>.0`, and M00 in v0.0.1. A fix between two milestones takes the next patch number, such as v0.4.1. Releases happen when Pascal asks and are never a card, so a release never blocks a milestone.

**Why.** A milestone that waited for its release, as M00 did for v0.0.1, blocked the next one until Pascal acted, and agents worked around it with an exception. Minor numbers per milestone leave patch numbers free for fixes.

**Rejected.** One release per milestone as a gate. Numbering milestones `0.1.<n>`, because a fix between milestones would take the next milestone's number.
