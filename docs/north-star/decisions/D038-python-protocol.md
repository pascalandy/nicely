# D038 Exchange acknowledged JSON Lines with the Python program

Decided 2026-10-07. Go and Python exchange versioned JSON Lines. An initial plan fixes the item set for both dry run and execution. Python announces an intent before user-side output or Deepgram upload. Go records the selected destination before reserving it, completes the persistence barrier, then acknowledges the final folder. [transcript's spec](../../../extensions/transcript/spec.md#protocol-with-the-python-program) holds the messages.

**Why.** One preparation avoids selecting a different Zoom meeting during execution. Recording a destination after creating it loses ownership on interruption, so Go records it first. No paid call starts before the durable intent; a lost acknowledgement stops Python before upload. Persisted verified artifacts can establish completion after a lost final message, without another request. This is the temporary Python boundary, not a protocol imposed on future extensions.

**Rejected.** Arguments plus one final JSON object, because Go could not record the intent before the upload. A separate file or socket, because stdin and stdout already connect the two processes.
