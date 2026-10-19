# D004 Keep core minimal and make every domain an extension

Decided 2026-10-05, revised 2026-10-19. Core hosts extensions. Only what agents need to operate Nicely is bundled. [Architecture](../architecture.md#core-and-extensions) assigns each part.

Nicely may ship everyday domains by default as external extensions, as settled in [#273](https://github.com/pascalandy/nicely/issues/273).
Each default extension is a separate executable, never compiled into core.
It uses the same extension path as any other external extension.

**Why.** First-party domains prove the same extension contract that others use. Default extensions exercise the same path as extensions installed separately. Adding a domain never edits core. Nicely operates without an agent, so `agent` is external too. Pascal's personal tools stay in his private tap.

**Rejected.** Domains inside core create a privileged path and require a later rewrite. Compiling default everyday domains into core would bypass the path for external extensions.
