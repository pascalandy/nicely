# D054 Find the skills of installed extensions through ncly describe

Decided 2026-10-12. The `skill` extension gets installed skill folders through `ncly describe`, under [Extensions](../contract/extensions.md) and [ncly skill](../../../extensions/skill/spec.md#ncly-skill). M03 therefore precedes M04.

**Why.** Host discovery stays private while `skill` sees exactly the selected extensions. Slug-named skills need no separate lookup, under [Extension identity](../contract/extension-identity.md). Folders permit resources beside `SKILL.md`; the bundled `nicely` skill also teaches the skill domain.

**Rejected.** Public SDK discovery, environment-injected inventories, duplicate `PATH` scans, and a lone skill file without its resources.
