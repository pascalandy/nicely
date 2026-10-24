# D056 Let an extension add doctor checks through its status command

Decided 2026-10-13. Doctor calls an extension's report command only when its manifest marks `status` with `doctor = true`. [Status command](../contract/status-command.md) owns effects, timeouts, and report validation.

**Revision, 2026-10-24.** [D064](D064-shared-domains.md) applies this rule to named extensions within a shared domain.

**Why.** Domain checks need domain code. Agents can also call the same report directly. Static program prerequisites remain in manifests so doctor can check an extension that cannot start.

**Rejected.** A declarative language for arbitrary checks, hooks inside core, and calling unmarked status commands that may serve another purpose.
