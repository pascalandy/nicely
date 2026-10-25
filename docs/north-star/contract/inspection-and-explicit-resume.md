# Inspection and explicit resume

Inspection is read-only and reports saved evidence, not an assumption that a process marked running is still alive. `ncly agent run` and transcript support inspection and verified-result replay, not `ncly run resume`. Partial multi-step continuation and child-run recovery remain in [COR-621](../../backlog/COR-621-resume-sdk.md#later).

A repeated call can resolve a lost final message from a complete set of independently verified local artifacts only when its domain specifies the evidence. Hold the run lock, validate the record version and stated inputs, verify the artifacts, and durably record completion before answering. This starts no network request and executes no unfinished step. Incomplete or changed evidence returns `RESUME_UNSAFE`, with the artifacts still available. `--force` never overrides uncertainty or makes an unknown paid request safe.

Nicely has no daemon or scheduler. A record makes work inspectable after a session ends; it does not keep the process alive. A user must inspect unresolved evidence and explicitly choose any new operation that might repeat a charge.
