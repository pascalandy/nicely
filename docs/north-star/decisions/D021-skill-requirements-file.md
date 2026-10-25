# D021 Move skill requirements to a file

Decided 2026-10-05. Standalone skill prerequisites would move from frontmatter to `nicely.toml` beside `SKILL.md`, reusing extension manifest concepts.

**Scope update, 2026-10-17.** [COR-622](../../backlog/COR-622-taps-offer.md#later) defers standalone skill distribution, prerequisite manifests, and related doctor checks. `ncly doctor skill` keeps its structural checks.

**Why.** Reuse avoids another frontmatter reader. No second manifest consumer is needed before that distribution capability.
