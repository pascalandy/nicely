# D013 Write core and the first-party extensions in Go

Decided 2026-10-05, revised 2026-10-09. Core and the first-party extensions aim to be fully Go. An extension may embed a Python program during its move, as transcript does, but that program returns JSON only, and the extension's Go code owns display, translation, and exit codes. An extension from a tap may use any language, as D005 allows.

**Why.** One binary and one display layer keep the experience identical across commands.

**Rejected.** Letting Python keep its own display.
