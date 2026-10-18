# D012 Make every word of core translatable from day one

Decided 2026-10-05. Core uses go-i18n catalogs with descriptions. Extensions own their catalogs. English ships first; Canadian French follows in M22. The package comment of [sdk/i18n](../../../sdk/i18n/i18n.go) owns catalog use, and [the scenario rules](../../../testdata/script/README.md) own pseudo-locale scenarios from M00.

A catalog test compares entries with `en`. It requires `one` beside `other` when `other` reads `Count` or `N`.

**Why.** Early translation avoids revisiting every string and format. A pseudo-locale exposes missing strings and truncated layouts. The catalog test catches missing plural forms even when English renders both forms identically.

**Rejected.** A string-literal check alone cannot prove layout.
