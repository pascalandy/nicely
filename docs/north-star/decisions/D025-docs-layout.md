# D025 Split the docs by question, and work in milestones

Decided 2026-10-05, revised 2026-10-09. `docs/north-star/` holds one file per question, as [Find what you need](../../../AGENTS.md#find-what-you-need) maps them, and one file per decision. Each extension keeps its spec in `extensions/<name>/spec.md`. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. [Keep the docs a tower](../dev-preferences.md#keep-the-docs-a-tower) holds the rules that keep this shape. A milestone becomes ready only once the spec sections that its cards read are written, and its cards link to them.

**Why.** Each rule has one source, so milestones cannot drift from the spec. An extension's spec lives with the rest of what the extension owns. A file costs an agent its size each time it loads, so the files read every session stay short, and the rest load only when a card or a question needs them. When each file answers one named question, agents and Pascal know where to look and where to add.

**Rejected.** A spec inside each milestone file. One guide holding the vision, the principles, the architecture, the terms, and the doc map, because every session loaded all of it and its name said none of it. One file for every decision, because reading one rule loaded all 43.
