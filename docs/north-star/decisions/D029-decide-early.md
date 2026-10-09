# D029 Decide early what is costly to change

Decided 2026-10-05. A rule that a later milestone would otherwise rewrite goes into the north-star files or an extension's spec before the code that it constrains. Its code waits for the milestone that uses it.

**Why.** Extensions, Canadian French, the shared config, and agent profiles all put constraints on the contract, the config, and the catalogs. A rule written early costs a paragraph. The same rule found later costs a rewrite of code already written.

**Rejected.** Deciding each rule only when its milestone starts.
