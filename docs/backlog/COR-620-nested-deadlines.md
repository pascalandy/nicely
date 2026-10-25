---
id: COR-620
core: "Nested deadlines and cleanup"
state: "Idea. Revisit when An approved summary slice actually calls a nested recorded command"
---

# COR-620 Nested deadlines and cleanup

## Demo

| Idea | What it would do | Revisit when |
|---|---|---|
| Nested deadlines and cleanup | The program runner passes `NCLY_DEADLINE`, an absolute UTC time, to every program, and each `ncly` ends its work before that time minus its 15 seconds of cleanup. A nested `ncly` also gets a shorter cleanup than its parent. It then times out, or writes its record after the signal, before the parent's SIGKILL | An approved summary slice actually calls a nested recorded command |
