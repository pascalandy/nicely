# D034 Use the same preparation for simulation and execution

Decided 2026-10-05. Dry run and execution share [preparation](../contract.md#operations). Dry run checks one invocation; doctor checks the machine. Execution revalidates mutable conditions.

**Why.** Separate simulations can approve different destinations. A dry run cannot prove credentials or reserve files. Its [cache exception](../contract.md#global-flags) and pending checks make those limits explicit. Doctor belongs at setup or unexplained failures, not before every call.

**Rejected.** Separate planning logic, silently skipped checks, and a prior dry run treated as overwrite permission.
