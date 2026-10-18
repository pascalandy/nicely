# D053 Give a problem its next command as words

Decided 2026-10-12. A problem may add `next`, an array of command words that runs as written, under [Output](../contract.md#output). The hint remains human text; the exit code still decides who may act or retry.

**Why.** Word arrays avoid parsing translated hints or shell quoting.

**Rejected.** Shell strings and typed action lists. Discovery and dry run already describe effects.
