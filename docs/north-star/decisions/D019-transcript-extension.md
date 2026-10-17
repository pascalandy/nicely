# D019 Ship transcript as a first-party extension

Decided 2026-10-05, revised 2026-10-09. `transcript` is a first-party external extension.
The original plan combined YouTube, Zoom, and summaries.

**Scope update, 2026-10-17.** [M13](../../milestones/M13-transcript.md) starts with one
`ncly transcript run youtube` call for one explicit URL, as [its spec](../../../extensions/transcript/spec.md)
defines. It requires a granted Deepgram key and no agent. Zoom belongs to
[transcript extras](../../milestones/M99-parking-lot.md#transcript-extras), and summary to
[its own deferred capability](../../milestones/M99-parking-lot.md#transcript-summary).

**Why.** Transcription is a domain task, not part of operating Nicely. Its required service key
makes a simpler everyday tool a better first experience for a new visitor. The domain name
can serve future audio sources without requiring them in its first milestone.

**Rejected.** `ncly video transcribe`, because transcription is not inherently a video operation.
