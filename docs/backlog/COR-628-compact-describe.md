---
id: COR-628
core: "compact describe"
state: "Idea from #375. Revisit when a new contract version is planned"
---

# COR-628 compact describe

`ncly describe <cmd> --json` would leave out `result.schema` and `dry_run_schema` by default, and give them on request.

The schemas take 60 to 80 percent of the bytes of a detailed entry.
On `052b3184`, `describe doctor --json` drops from 2,532 to 1,077 bytes without them, and `describe auth login --json` from 3,734 to 1,262.
The `nicely` skill reads one such entry before each new command, as [#375](https://github.com/pascalandy/nicely/issues/375) measured.

A compact default removes keys from an existing answer.
That is a breaking change under [Compatibility](../north-star/contract/compatibility.md), so it needs a new contract version and a decision.
