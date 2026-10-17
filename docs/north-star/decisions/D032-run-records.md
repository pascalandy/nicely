# D032 Keep small run records and resume explicitly

Decided 2026-10-05. The [operation contract](../contract.md#operations) defines preparation, execution, inspection, and recovery. Domains own their steps and recovery rules. The program runner owns subprocess lifecycles. Records preserve evidence with references to outputs rather than copies of their contents. M00 fixed the contract; M10 and M13 are its first planned consumers.

**Scope update, 2026-10-17.** Each first consumer has one item and one paid step. M10 implements the needed record support and core's [inspection commands](../core-spec.md#ncly-run). Verified completion replay and refusal of uncertain paid work remain required. [Explicit resume](../../milestones/M99-parking-lot.md#resume), multi-step consumers, and their fixtures wait for a workflow with useful partial work. The record format is not a requirement to build every future recovery path now.

**Why.** A new session must distinguish verified output from an uncertain paid request. Persisting intent before an effect preserves evidence if the process dies before saving its result. Replay verifies saved outputs and keeps the run's resolved settings. A missing response never proves that nothing happened. Future resume must validate the evidence needed for each step it reuses.

**Rejected.** Restarting every invocation from scratch. Logs as the only record of completed work. Recording keys or duplicating all transcript content in history. A daemon, scheduler, or general workflow engine before a command needs one.
