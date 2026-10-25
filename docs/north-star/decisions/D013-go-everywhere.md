# D013 Write core and every extension in Go

Decided 2026-10-05, revised 2026-10-09 and 2026-10-20. Core and every extension target Go with the SDK, the one supported way for now. Other languages for an extension wait in [COR-617](../../backlog/COR-617-extension-languages.md). The wire contract stays open to any language, as [D040](D040-wire-contract.md) decides. During migration, Nicely's own code may embed Python that returns JSON only. A worker, a user's script that an extension embeds, keeps its language, as [Programs that ncly runs](../contract/programs-that-ncly-runs.md) defines. Go owns display, translation, and exit codes.

**Why.** One display layer keeps commands consistent. One supported way lets the SDK and its test kit prove every extension. A user's script already works, so a rewrite in Go adds risk and no value.

**Rejected.** Python owning a separate display. A promise that an extension may use any language, which no test proves. Rewriting each wrapped script in Go.
