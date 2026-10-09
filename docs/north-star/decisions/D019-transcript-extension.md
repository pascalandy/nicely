# D019 Ship transcript as a first-party extension

Decided 2026-10-05, revised 2026-10-09. `transcript` is a first-party external extension: `ncly transcript run youtube` and `ncly transcript run zoom`, as [its spec](../../../extensions/transcript/spec.md) defines.

**Why.** It needs a Deepgram key and an agent for its summary, so it is no part of operating Nicely, and a first-time visitor should start elsewhere. It handles Zoom audio as well as YouTube.

**Rejected.** `ncly video transcribe`, the wrong domain for Zoom audio.
