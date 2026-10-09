# D021 Move skill requirements to a file

Decided 2026-10-05. Skills declare their requirements in frontmatter today. Extensions use one versioned manifest format from M01, and skills join it when taps carry them in M12. A skill keeps `nicely.toml` beside `SKILL.md`; an extension keeps `<executable>.toml` beside its executable so several extensions can share a directory. Until then, `ncly doctor skill` checks only the structure of a skill.

**Why.** One file format serves skills and extensions alike. A frontmatter reader for requirements would be thrown away once manifests exist.
