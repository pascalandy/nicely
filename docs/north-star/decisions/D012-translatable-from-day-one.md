# D012 Make every word of core translatable from day one

Decided 2026-10-05. Core text lives in go-i18n catalogs with a description per entry. A test compares every catalog with `en`, and a pseudo-locale scenario exposes text outside the catalog from M00. Plurals, numbers, sizes, and dates go through the i18n package. The catalog check requires a `one` form beside `other` in each entry whose `other` form reads a field named `Count` or `N`. English ships first, and Canadian French follows in M22. Each extension keeps its own catalog.

**Why.** Adding translation later means touching every string and every number format. AI makes catalogs cheap to keep complete, and a CLI in the user's language reaches people most CLIs ignore. The pseudo-locale catches a missing string in the milestone that adds it instead of in M22. When English writes `one` and `other` the same, as for a unit symbol, no scenario shows a missing plural form. The catalog check catches it instead.

**Rejected.** A static check for string literals, because the pseudo-locale also catches truncated layouts.
