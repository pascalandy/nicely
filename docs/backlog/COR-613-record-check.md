---
id: COR-613
core: "Record check in core"
state: "Idea. Revisit when a tap carries an extension that does not use the Go SDK"
---

# COR-613 Record check in core

Core reads the run record of an answer that exits 75.
When the record breaks [Retry safety](../north-star/contract/retry-safety.md), core turns the answer into a protocol failure, as [D045](../north-star/decisions/D045-step-effects.md) defers.
