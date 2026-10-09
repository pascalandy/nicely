# D025 Keep rules in north-star and the extension specs, and work in milestones

Decided 2026-10-05, revised 2026-10-09. `docs/north-star/` holds the guide, the CLI spec of core, Pascal's developer preferences, and this log. Each extension keeps its spec in `extensions/<name>/spec.md`. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. A milestone becomes ready only once the spec sections that its cards read are written, and its cards link to them.

**Why.** Each rule has one source, so milestones cannot drift from the spec. An extension's spec lives with the rest of what the extension owns. Agents read the guide every session and the rest only when a card links to it.

**Rejected.** A spec inside each milestone file.
