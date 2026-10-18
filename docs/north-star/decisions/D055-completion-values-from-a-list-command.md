# D055 Complete an extension's values from a list command

Decided 2026-10-13. Reuse a qualifying list command through `values_from`, under [Completion values](../contract.md#completion-values). Free network reads qualify; writes and billing do not. Core dispatches the exact extension it checked.

**Why.** Agents and completion need the same data. Reuse works across domains and languages without a second protocol. Suggestions do not constrain valid input.

**Rejected.** Cobra's private completion protocol, stale static values or caches, and proxy commands. A second `ncly` discovery could run a different installation than the one checked.
