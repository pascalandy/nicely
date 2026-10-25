# D032 Keep small run records and resume explicitly

Decided 2026-10-05. [Operations](../contract/operations.md) separates domain steps and recovery from process lifecycles. Records keep evidence and output references.

**Scope update, 2026-10-17.** M10 and M13 first prove one item and one paid step, inspection, verified completion replay, and refusal of uncertain work. [COR-621](../../backlog/COR-621-resume-sdk.md#later) defers explicit resume and multi-step consumers until partial work needs reuse.

**Why.** A later session must distinguish verified output from an uncertain paid request. Intent precedes effects, and reuse requires verified evidence and the original resolved settings.

**Rejected.** Restarting every call, logs as the only evidence, keys or transcript bodies in records, and a daemon or workflow engine without a consumer.
