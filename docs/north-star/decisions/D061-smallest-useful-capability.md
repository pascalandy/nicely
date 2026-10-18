# D061 Ship the smallest useful capability before its variants

Decided 2026-10-17. Start with one concrete job and its necessary inputs, outputs, and steps. Each variant needs its own observed use. Each milestone delivers a useful increment without later milestones as acceptance criteria.

Keep deferred capabilities and revisit conditions in [M99](../../milestones/M99-parking-lot.md), as options rather than obligations. Remove deferred milestones from the queue. The repository plan remains authoritative over old issues.

**Why.** [PR #156](https://github.com/pascalandy/nicely/pull/156) found brainstormed variants had become first-release requirements. Scope cuts retain applicable safety guarantees:

- Key grants and source identity
- Durable intent before paid effects, verified results, inspection, and refusal of unknown repeats
- Dry run, truthful exits, cancellation, output validation, and overwrite consent

**Rejected.** A Codex-only product, all variants hidden in one card, safety or proven workflows removed for line counts, and a new deferral framework. Nicely remains a CLI toolbox.
