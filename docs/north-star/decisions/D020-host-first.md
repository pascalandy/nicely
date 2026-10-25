# D020 Build the extension host before any domain

Decided 2026-10-05, revised 2026-10-09 and 2026-10-12. Build the extension host before domains, then add capabilities on proven layers. [The backlog queue](../../backlog/README.md#queue) owns the current order and scope. `skill` follows `describe` for the reason in [D054](D054-extension-skills-through-describe.md).

**Why.** Domains built before the host need a rewrite. A one-step agent task proves records before more complex workflows. Early integration of the real Python program exposes protocol mismatches.

**Rejected.** The original eleven-task day-one milestone mixed many capabilities and separate acceptance lists. Agents finished isolated tasks, leaving progress unclear.
