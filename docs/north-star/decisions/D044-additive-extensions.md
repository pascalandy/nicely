# D044 Let extensions add commands, never change core

Decided 2026-10-09. Extensions add commands without hooks that change core or other extensions. Shared behavior keeps its [code-map owner](../architecture.md#code-map). Core calling a declared completion or status command is contract use, not a hook.

**Why.** Each command stays understandable from its spec. A cross-process toolbox needs no lifecycle hooks like Pi's in-process agent loop.

**Rejected.** Hooks create permanent protocol obligations before a domain needs them. [COR-615](../../backlog/COR-615-hooks.md) defers them. Use by several commands alone does not transfer behavior to core.
