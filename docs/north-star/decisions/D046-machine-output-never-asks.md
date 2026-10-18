# D046 Never ask a human for machine output

Decided 2026-10-10. `--json` and `NCLY_JSON=1` select non-interactive mode under [Modes](../contract/modes.md), as [Terraform's JSON mode](https://developer.hashicorp.com/terraform/cli/commands/plan) does.

**Why.** Core captures extension stdout to validate JSON. Prompts would mix into that answer. Core commands must behave the same way.

**Rejected.** A separate mode variable or prompts only for core commands. Neither preserves one behavior across domains.
