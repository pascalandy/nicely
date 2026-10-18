# D013 Write core and the first-party extensions in Go

Decided 2026-10-05, revised 2026-10-09. Core and first-party extensions target Go. During migration, an extension may embed Python that returns JSON only. Go owns display, translation, and exit codes. Third-party extensions may use any language under [D005](D005-extensions-and-taps.md).

**Why.** One display layer keeps commands consistent.

**Rejected.** Python owning a separate display.
