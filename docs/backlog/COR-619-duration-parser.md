---
id: COR-619
core: "One duration parser"
state: "Idea. Revisit when a change touches `ncly doctor --timeout`, or a third command takes `--timeout`"
---

# COR-619 One duration parser

`ncly doctor` reads its `--timeout` through `sdk/command.Duration`, as `ncly agent run` does, instead of its own `parseTimeout` in `internal/cli/doctor.go`, so both commands share one rule.
