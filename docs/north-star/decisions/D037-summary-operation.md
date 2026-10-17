# D037 Summarize a saved transcript as a new operation

Decided 2026-10-07. The original plan proposed `ncly transcript summary run <source-run-id>` as a separate recorded operation. It would summarize verified saved transcripts without retranscribing them. That design also included item selection, profiles, and safe resume of never-started summary steps.

**Scope update, 2026-10-17.** Summary is outside the active milestone queue and belongs in [M99](../../milestones/M99-parking-lot.md#transcript-summary). The first candidate is one summary of one verified saved transcript, once [D028](D028-outside-text.md) has a proven tools-off path. Selection across multiple items, prompt and profile management, and explicit resume remain deferred. M13 performs transcription only.

**Why.** A summary can bill, so a lost answer may already be charged. A separate operation lets the caller explicitly choose a new bill without paying Deepgram again. Any later implementation must record the verified source references and preserve its source record. It must never repeat an unknown paid request automatically.

**Rejected.** Copying transcript content into the new record, which adds retention work without proving that the source is unchanged. Copying a completed `transcribe` step into a run that never transcribed. A resume flag that overrides an unknown effect. Treating every summary as free. Rerunning the whole transcript to obtain another summary.
