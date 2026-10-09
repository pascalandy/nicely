# D034 Use the same preparation for simulation and execution

Decided 2026-10-05. `--dry-run` and execution use the same preparation of inputs, profiles, destinations, and effects. A simulation distinguishes verified conditions from checks deferred until execution. The [dry-run contract](../contract.md#global-flags) allows only documented, bounded cache preparation; it reads no key and creates no durable run record. Execution revalidates conditions that may have changed.

**Why.** A separately written simulation can approve a destination that execution resolves differently. A dry run also cannot prove that a key works or reserve files against another process. Reporting these limits gives agents enough evidence to proceed without overstating what was checked.

**Rejected.** Parallel implementations of planning and execution. Calling a simulation successful while silently skipping a required check. Treating a prior dry run as permission to overwrite a changed destination.
