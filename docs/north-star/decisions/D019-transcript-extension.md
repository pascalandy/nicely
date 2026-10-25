# D019 Ship transcript as a first-party extension

Decided 2026-10-05, revised 2026-10-09. `transcript` is a first-party external extension.

**Scope update, 2026-10-17.** [M13](../../milestones/M13-transcript.md) delivers one explicit YouTube URL with a granted Deepgram key and no agent. [Its spec](../../../extensions/transcript/spec.md) owns the workflow. [XT-102](../../backlog/XT-102-transcript.md#transcript-extras) defers Zoom and [summary](../../backlog/XT-102-transcript.md#transcript-summary).

**Why.** Transcription is a domain task. Its service key makes a simpler everyday tool a better first experience. The domain can accept future audio sources.

**Rejected.** `ncly video transcribe`, because transcription is not inherently a video operation.
