# D044 Let extensions add commands, never change core

Decided 2026-10-09. Extensions add commands without hooks that change core or other domains. Shared behavior keeps its [code-map owner](../architecture.md#code-map). Core calling a declared completion or status command is contract use, not a hook.

**Revision, 2026-10-24.** [D064](D064-shared-domains.md) applies this rule to named extensions within a shared domain.

**Why.** Each command stays understandable from its spec. A cross-process toolbox needs no lifecycle hooks like Pi's in-process agent loop.

**Rejected.** Hooks create permanent protocol obligations before a domain needs them. [M99](../../milestones/M99-parking-lot.md) defers them. Use by several commands alone does not transfer behavior to core.
