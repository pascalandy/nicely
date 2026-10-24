# D022 Keep docs private by default

Decided 2026-10-05. Any future publishing capability defaults projects to `private`, with explicit `public` or `exclude` choices.

**Scope update, 2026-10-17.** [M24](../../milestones/M24-docs-hub.md) uses an existing Markdown renderer for explicitly selected local docs. It keeps sources unchanged, generates disposable cache, and binds to loopback. [M99](../../milestones/M99-parking-lot/docs-hub.md) defers the registry, visibility settings, synchronization, and publication.

**Why.** Personal details can enter docs unnoticed. Publication must be a per-project choice; local viewing needs no publication machinery.
