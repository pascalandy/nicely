# D022 Keep docs private by default

Decided 2026-10-05. The original docs hub plan classified projects as `public`, `private`, or `exclude`, with `private` as the default. Markdown sources let an existing engine render them.

**Scope update, 2026-10-17.** [M21](../../milestones/M21-docs-hub.md) serves only an explicit set of project docs on a loopback address. Source folders remain unchanged, and generated files stay in disposable cache. A project registry, visibility settings, synchronization, and publication remain in [M99](../../milestones/M99-parking-lot.md#docs-hub). Local viewing requires none of them.

**Why.** Personal details slip into docs without anyone noticing. A later publishing capability must make publication a choice for each project. The first local viewer publishes nothing.
