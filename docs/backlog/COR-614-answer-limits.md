---
id: COR-614
core: "Answer limits"
state: "Idea. Revisit when a tap carries an extension that does not use the Go SDK"
---

# COR-614 Answer limits

The answer check bounds the size of a held answer and of a held stderr line, its JSON depth, and rejects invalid UTF-8 and duplicate keys, then stops the extension and answers `RUNTIME`. An extension built on the SDK cannot produce these answers.
