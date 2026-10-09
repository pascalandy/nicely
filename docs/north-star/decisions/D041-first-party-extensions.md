# D041 Keep first-party extensions in the Nicely repository

Decided 2026-10-09. First-party extensions live in `extensions/<name>/`, in the same Go module as core. Each folder holds what its extension owns: the manifest, `spec.md`, `SKILL.md`, catalogs, scenarios, and code. A lint rule keeps `sdk/` and `extensions/` away from `internal/`. The repository is also the official tap of these extensions.

**Why.** An agent changes the SDK and the extension that uses it in one pull request, with one checkout and one `just next`, while the contract is still young. The lint rule keeps the boundary that separate repositories would give.

**Rejected.** One repository per extension, as `gh` extensions do, because every SDK change would then take several pull requests.
