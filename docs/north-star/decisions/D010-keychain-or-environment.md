# D010 Keep keys in the keychain or the environment

Decided 2026-10-05, revised 2026-10-17. Read keys from the environment first, then Nicely's own OS keychain entries through go-keyring. [Keys](../contract/keys.md) owns service names, lookup, time limits, and unknown writes. Agents relay hints and never handle keys.

**Why.** Environment variables support other secret stores and SSH sessions without an unlocked keychain. Separate entries keep logout from removing another tool's keys. Open service names let manifests declare keys without changes to core. The reserved names `keychain` and `grants` belong to doctor's checks.

The limits are 60 seconds interactively and 10 otherwise, with one unavailable-keychain wait per command. A human can answer an unlock prompt; an unattended agent cannot. A timed-out write can still complete, so its effect stays unknown. [Programs that ncly runs](../contract/programs-that-ncly-runs.md) explains the macOS `security` exception.

**Rejected.** Fallback files and shared ownership of another tool's entries. Remote commands that need keys remain in [M99](../../milestones/M99-parking-lot.md). Moving `security` through the higher-layer runner would not make an unknown write safe.
