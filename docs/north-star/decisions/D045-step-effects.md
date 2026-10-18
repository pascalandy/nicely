# D045 Derive retry safety from declared step effects

Decided 2026-10-10, revised 2026-10-16. Record support derives [retry safety](../contract.md#retry-safety) from declared step effects. [Record format](../contract.md#record-format) defines their order and preserves the more cautious of stored and current declarations.

**Why.** A second, manual effect report can drift and authorize another charge. Stored restrictions survive rollbacks; current restrictions prevent reuse of outdated permissions. [pi-durable](https://github.com/earendil-works/pi/blob/4dd2b149d155d7369de16ea4dc77c84e9f5ffb24/packages/durable/docs/spec.md#L3385-L3388) makes the same stored-and-current policy choice.

**Rejected.** Domain-computed retry verdicts and trust in either policy alone. Core verification of an extension's record waits for a non-Go consumer in [M99](../../milestones/M99-parking-lot.md). [Resume fixtures](../../milestones/M99-parking-lot.md#resume) wait for a selected multi-step consumer; M10 proves one paid step.
