# D051 Run a bundled extension as a child of the same ncly binary

Decided 2026-10-11, revised 2026-10-19. Run bundled extensions as child processes of `ncly`, selected by `argv[0]`. An explicit list in `cmd/ncly` supplies the same `extension.Extension` values that standalone executables use. [SDK](../contract/sdk.md) owns this interface.

**Why.** Both transports share arguments, environment, cancellation, and answer checks. A blocked child can be killed; Go context cancellation only requests cooperation. The bundled transport reuses the `ncly` binary and starts another process.

**Rejected.** In-process handlers restrict `os.Exit`, environment, and signal ownership. Registration through `init` hides what gets bundled.

[D004](D004-minimal-core.md) permits default extensions as separate executables.
