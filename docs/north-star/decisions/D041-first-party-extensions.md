# D041 Keep first-party extensions in the Nicely repository

Decided 2026-10-09. Keep first-party extensions in `extensions/<name>/`, in core's Go module and official tap. [Architecture](../architecture.md#code-map) assigns their files; lint enforces the `internal/` boundary.

**Why.** One checkout and pull request can change the young SDK and its consumer together.

**Rejected.** Separate repositories would require several pull requests per SDK change.
