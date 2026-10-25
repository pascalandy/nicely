# D029 Decide early what is costly to change

Decided 2026-10-05. Specify a rule before the code that it constrains if a later item would otherwise require a rewrite. Implement it only when an item needs it.

**Why.** Extension boundaries, translation, and shared configuration constrain earlier code.

**Rejected.** Waiting until each rule's own item to decide it.
