# D055 Complete an extension's values from a list command

Decided 2026-10-13. Reuse a qualifying list command through `values_from`, under [Completion values](../contract/completion-values.md). Free network reads qualify; writes and billing do not. Core dispatches the exact extension it checked.

**Revision, 2026-10-24.** [D064](D064-shared-domains.md) applies this rule to named extensions within a shared domain.

**Why.** Agents and completion need the same data. Reuse works across domains and languages without a second protocol. Suggestions do not constrain valid input.

**Rejected.** Cobra's private completion protocol, stale static values or caches, and proxy commands. A second `ncly` discovery could run a different installation than the one checked.
