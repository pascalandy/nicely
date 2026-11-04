---
id: COR-630
core: "core diagnostics"
state: "Idea from #447. Revisit with the first core command that needs debug or verbose details"
---

# COR-630 core diagnostics

The SDK would let core commands add verbose step details and debug diagnostics on stderr.
Core adds neither today.
The current `--verbose` help entry says `Add step details on stderr`, although core adds none.
Revisit that catalog entry with the shared support.
An extension can already add diagnostics when it needs them, as [Global flags](../north-star/contract/global-flags.md) describes for transcript.

The first core command that needs these details must add shared SDK support and prove its output.
Keys must stay out of diagnostics.
