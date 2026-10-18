# D009 Ask in a terminal, fail elsewhere

Decided 2026-10-05. Ask for missing values in interactive mode and exit 2 otherwise. After a form, print the equivalent command after `Next time:`. [Modes](../contract.md#modes) defines when to ask.

**Why.** Forms guide beginners without blocking agents. A pseudo-terminal alone does not prove a human is present, so `CI` and `NCLY_NO_INPUT` also disable prompts.
