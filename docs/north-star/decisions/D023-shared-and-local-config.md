# D023 Configure with a shared TOML file and a local one

Decided 2026-10-05, revised 2026-10-13. Use shared TOML config with local overrides, as [Configuration](../contract/configuration.md) defines. Only setup-changing commands edit it, preserving comments and formatting. Nicely's own environment variables start with `NCLY_`.

**Why.** Pascal's tools use TOML. Shared setup travels between machines; local paths do not. Unknown keys warn so older versions remain usable. Manifests identify installed extensions' keys without execution. Tables for absent extensions stay unchecked, accepting that misspelled table names can go unreported.

**Rejected.** YAML, warnings for absent extensions' tables, and machine-only setup records that require every tap to be added again.
