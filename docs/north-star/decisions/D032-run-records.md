# D032 Keep small run records and resume explicitly

Decided 2026-10-05. The [operation contract](../contract.md#operations) covers preparation, execution, inspection, and explicit resume. Domains own their steps and recovery rules. The program runner owns subprocess lifecycles. Records preserve the evidence needed to reuse finished work, with references to outputs rather than copies of their contents. M00 fixed the contract. `ncly agent run` in M10 and transcript from M13 implement it, and core exposes the [run commands](../core-spec.md#ncly-run).

**Why.** A new agent session must be able to tell whether Deepgram finished before a summary failed. Writing the intended non-repeatable effect before starting it leaves evidence even if the process dies before saving the result. Resume validates inputs, relevant configuration, and saved outputs, and keeps the profile that the run recorded instead of reading it again. An uncertain paid request needs reconciliation or a human decision, because a missing response does not prove that nothing happened.

**Rejected.** Restarting every invocation from scratch. Logs as the only record of completed work. Recording keys or duplicating all transcript content in history. A daemon, scheduler, or general workflow engine before a command needs one.
