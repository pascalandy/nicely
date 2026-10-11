# D046 Never ask a human for machine output

Decided 2026-10-10. `--json` and `NCLY_JSON=1` select non-interactive mode, as [Modes](../contract.md#modes) defines, the way Terraform's [`-json` implies `-input=false`](https://developer.hashicorp.com/terraform/cli/commands/plan). A call that lacks a value or a confirmation fails with exit 2, and a step that only a human can do fails with exit 78, whatever the terminal.

**Why.** With `--json`, core holds an extension's stdout to check its answer, as [D050](D050-answer-check.md) decides, so the extension never sees a terminal there and never asks. Without this rule, the same call in a terminal would open a form for a core command and fail for an extension, which breaks principle 2. A caller that asks for JSON is a program, or a human who pipes the answer into one, so nobody is there to answer a form.

**Rejected.** Passing core's mode to the extension in a new variable, because an extension that asked while core held its stdout would draw its form into the answer. Letting `--json` ask for core's commands only, because a call would then behave differently by domain.
