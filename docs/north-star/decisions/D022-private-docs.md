# D022 Keep docs private by default

Decided 2026-10-05. Any future publishing capability defaults projects to `private`, with explicit `public` or `exclude` choices.

**Scope update, 2026-10-17.** [XT-104](../../backlog/XT-104-docs-hub.md) uses an existing Markdown renderer for explicitly selected local docs. It keeps sources unchanged, generates disposable cache, and binds to loopback. [XT-104](../../backlog/XT-104-docs-hub.md#later) defers the registry, visibility settings, synchronization, and publication.

**Why.** Personal details can enter docs unnoticed. Publication must be a per-project choice; local viewing needs no publication machinery.
