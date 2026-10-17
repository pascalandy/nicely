# D021 Move skill requirements to a file

Decided 2026-10-05. The original plan moved skill requirements from frontmatter to a
`nicely.toml` beside `SKILL.md` when taps began distributing standalone skills.
Extensions already keep their versioned `<executable>.toml` beside the executable.

**Scope update, 2026-10-17.** M12 installs one extension with its existing manifest, catalogs,
and bundled skill folder. Standalone tap skills, their prerequisite manifests, and related
doctor checks remain in [M99](../../milestones/M99-parking-lot.md#taps).
`ncly doctor skill` keeps its existing structural checks. M12 adds no new skill schema.

**Why.** If a later skill-distribution slice needs prerequisites, it should reuse the extension
manifest concepts instead of creating another frontmatter reader. That possible reuse does
not require a second manifest consumer in the first installation milestone.
