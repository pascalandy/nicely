# D004 Keep core minimal and make every domain an extension

Decided 2026-10-05, revised 2026-10-09. Core hosts extensions. Only what agents need to operate Nicely is bundled. [Architecture](../architecture.md#core-and-extensions) assigns each part.

**Why.** First-party domains prove the same extension contract that others use. Adding a domain never edits core. Nicely operates without an agent, so `agent` is external too. Pascal's personal tools stay in his private tap.

**Rejected.** Domains inside core create a privileged path and require a later rewrite. Bundled everyday domains break the minimal-archive rule. Separate first-party executables inside the archive would require cross-process services before a domain needed them.
