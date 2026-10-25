---
id: COR-608
core: "Completion after --"
state: "Idea. Revisit when a user hits it, or a command group starts taking arguments"
---

# COR-608 Completion after --

`ncly -- <Tab>` offers nothing, because ncly rejects a command name after `--`. Cobra offers the commands, so the fix answers Cobra's completion protocol in `Main`.
