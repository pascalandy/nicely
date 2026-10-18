# D038 Exchange acknowledged JSON Lines with the Python program

Decided 2026-10-07. Go and Python use versioned JSON Lines with a fixed plan and durable intent before acknowledgement. [Transcript's protocol](../../../extensions/transcript/spec.md#protocol-with-the-python-program) owns the messages and sequence.

**Why.** Recording the destination before reservation preserves ownership after interruption. An acknowledgement gates paid upload. Verified artifacts can resolve a lost final message without another request. This temporary Python boundary does not constrain future extensions.

**Rejected.** One final JSON answer cannot persist intent before upload. Extra files or sockets duplicate the existing stdin and stdout connection.
