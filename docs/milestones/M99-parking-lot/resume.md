# Resume

Parked from M15 and the broader M10 fixtures: `ncly run resume`, partial multi-step continuation,
nested child-run recovery, and compatibility tests for changing a step's effect policy.
This is a deferred capability, not a prerequisite for recording one paid step safely.

M10 and M13 retain locks, durable intent before a paid request, verified completion, run
inspection, same-ID completed-result replay, and refusal of an unknown paid effect.
Retry before any persisted non-repeatable intent remains governed by the shared contract.
Parked from M13 on 2026-10-18, when Pascal cut it: the local recovery of a lost final protocol message.
Complete recorded artifacts would let Go record completion without another upload.
Until then, such a loss leaves the transcript step unknown, and a repeat refuses, as
[D038](../../north-star/decisions/D038-python-protocol.md) records.

Revisit when a selected command has useful partial work that the user actually needs to
continue. Design and prove that one continuation before generalizing it to other domains.
Revisit the transcript recovery when a lost final message strands a paid transcript.
M15 is removed from the delivery queue.
