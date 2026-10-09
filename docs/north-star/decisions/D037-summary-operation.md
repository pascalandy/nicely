# D037 Summarize a saved transcript as a new operation

Decided 2026-10-07. `ncly transcript summary run <run-id>` summarizes the verified transcripts of a saved run as a new recorded operation. `ncly run resume` never repeats a `summarize` step whose outcome is `unknown`, and the silent summary retries of the transcript CLI are gone. [transcript's spec](../../../extensions/transcript/spec.md#ncly-transcript-summary-run) holds the rules.

**Why.** A summary can bill, so a lost answer may already be charged. Resume never repeats that unknown request. A new operation copies verified transcript references, keeps its own prompt and profile, and writes a file named with its run ID. It leaves the source record valid and supports safe resume of its own never-started summary. Holding the source lock prevents a new summary racing its source executor. The caller explicitly chooses any new bill.

**Rejected.** A resume flag that overrides an unknown step, because `--force` never makes an unknown effect safe. Treating every summary as free, because some profiles bill per call. Rerunning the whole transcript, which bills Deepgram again.
