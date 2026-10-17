# D021 Move skill requirements to a file

Decided 2026-10-05. The original plan moved skill requirements from frontmatter to a
`nicely.toml` beside `SKILL.md` when taps began distributing standalone skills.
Extensions already keep their versioned `<executable>.toml` beside the executable.

**Scope update, 2026-10-17.** Taps wait in [M99](../../milestones/M99-parking-lot.md#taps),
with standalone tap skills, their prerequisite manifests, and related doctor checks.
`ncly doctor skill` keeps its existing structural checks.

**Why.** If a later skill-distribution slice needs prerequisites, it should reuse the extension
manifest concepts instead of creating another frontmatter reader. That possible reuse does
not require a second manifest consumer in the first installation milestone.
