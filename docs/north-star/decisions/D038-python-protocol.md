# D038 Exchange acknowledged JSON Lines with the Python program

Decided 2026-10-07. Go and Python use versioned JSON Lines with a fixed plan and durable intent before acknowledgement. [Transcript's protocol](../../../extensions/transcript/spec/protocol.md) owns the messages and sequence.

**Why.** Recording the destination before reservation preserves ownership after interruption. An acknowledgement gates paid upload. This temporary Python boundary does not constrain future extensions.

**Cut.** On 2026-10-18, Pascal cut the recovery of a lost final message, and the `artifact` message with it. Such a loss leaves the step unknown, so a repeat refuses. [XT-102](../../backlog/XT-102-transcript.md#later) parks the recovery.

**Rejected.** One final JSON answer cannot persist intent before upload. Extra files or sockets duplicate the existing stdin and stdout connection.
