# D037 Summarize a saved transcript as a new operation

Decided 2026-10-07. The proposed `ncly transcript summary run <source-run-id>` is a separate recorded operation over a verified saved transcript, without retranscription. It must record verified source references, preserve the source record, and never automatically repeat an unknown paid request.

**Scope update, 2026-10-17.** [M99](../../milestones/M99-parking-lot/transcript-summary.md) defers summary until [D028](D028-outside-text.md) has a verified tools-off path. The first candidate is one source and one summary. Item selection, prompt and profile management, and resume remain deferred.

**Why.** A new summary can bill independently without another Deepgram charge.

**Rejected.** Transcript copies in records, invented completed transcription steps, resume overriding uncertainty, and retranscription to obtain another summary.
