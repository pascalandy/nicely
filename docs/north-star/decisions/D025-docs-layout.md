# D025 Split the docs by question, and work in milestones

Decided 2026-10-05, revised 2026-10-09. `docs/north-star/` holds one file per question. Its `decisions/` folder holds one file per decision. [Find what you need](../../../AGENTS.md#find-what-you-need) maps each question to its file. Each extension keeps its spec in `extensions/<name>/spec.md`. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. [Keep the docs a tower](../dev-preferences.md#keep-the-docs-a-tower) defines this layout. Before a milestone becomes ready, write the spec sections that its agent cards read and link each agent card to them.

**Why.** Each rule has one source, so milestones link to specs instead of repeating them. An extension's spec stays with the extension. The files read each session stay short. Agents load other files only when a card or question needs them.

**Rejected.** A spec inside each milestone file would repeat rules. One guide would make every session load the vision, principles, architecture, terms, and doc map. One file holding every decision would make a reader load unrelated decisions.
