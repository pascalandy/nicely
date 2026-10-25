---
id: XT-102
extension: "transcript"
state: "Ready. Next card: the dry-run lock fix. Shipped by M13 and M16. Later: summary, Go port, Omarchy, other inputs, live Deepgram check, recovery of a lost final message"
---

# XT-102 transcript

## Demo

A transcript dry run under a busy run ID answers like execution and starts no work.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| XT-102-T1 | Probe the run lock during dry run | agent | - | todo |

### XT-102-T1 Probe the run lock during dry run

- **Read:** [Record format](../north-star/contract/record-format.md), [Records and replay](../../extensions/transcript/spec/records-and-replay.md)
- **Proves:** `extensions/transcript/testdata/script/transcript_contenders.txtar`

Use the SDK's nonblocking lock probe before the transcript dry run reads its run record.
A held lock answers exit 75 with `TEMPORARY`, as execution does, and starts no program or paid request.
Prove the dry run while a real contender holds the run lock.
A free probe creates no lock file and releases the lock at once.
Keep verified replay and refusal of uncertain work unchanged.
The dry run writes no record.
Update the transcript spec's dry-run lock sentence with the fix.


## Later

### Transcript in Go

Ports the Python program of the transcript extension to Go.

Revisit condition: Python blocks a release or a translation.

### Structured summaries

A bundled prompt writes a versioned JSON file beside its Markdown.
The file holds decisions, actions, and open questions, each citing a transcript segment that a local check verifies.

Revisit condition: Pascal wants to search decisions or actions across summaries.

### Transcript on Omarchy

The paid end-to-end check of `ncly transcript` passes on Omarchy, with YouTube cookies from a Linux browser. [#23](https://github.com/pascalandy/nicely/issues/23) tracks it.

Revisit condition: Pascal needs transcripts on Omarchy.

### Transcript summary

Parked from M14 and the implicit summary in M13: summarizing transcripts, prompt selection,
agent invocation, and the evidence needed to associate each summary with its source.
M13 produces a transcript from one YouTube URL without invoking an agent.

The first candidate is one explicit summary of one verified saved transcript, with one prompt.
It requires a proven tools-off path for outside text under D028. The selected Codex path does
not supply that capability. Do not add a second harness merely to keep the old roadmap moving.
M14 is removed from the delivery queue; a concrete need and that capability must precede a
replacement slice. [Structured summaries](#structured-summaries) remain a separate option.

### Transcript extras

Parked from M13 and M16: multiple URLs and batch orchestration, Zoom folder discovery and
`--latest`, browser matrices, configurable prompts, output opening, spinners, and automatic
migration or retirement of the old transcript CLI and skills.
M13 keeps one YouTube URL and its three verified artifacts. M16 checks its prerequisites,
teaches the command, and proves that an independent fresh session can use it.

Revisit another input source for a real transcript need. Add browser support only when the
selected machine needs it. Retire the old CLI only after Pascal's required uses have parity;
a partial replacement does not justify deleting a working workflow. Pascal picked the spinner's
look in M07-T1; [the transcript spec at 29b311f](https://github.com/pascalandy/nicely/blob/29b311fed17791eb8d91d043a70751c3c2ce3d3e/extensions/transcript/spec.md)
records it, so a spinner needs no new mockup.

#### Live Deepgram check

Parked from M16 on 2026-10-19: `--live` on `ncly transcript status`, which would ask Deepgram whether it accepts the granted key.
No Deepgram document says that `GET /v1/auth/token` bills nothing.
[Authenticating](https://developers.deepgram.com/docs/authenticating) suggests the endpoint, and says nothing of its cost:

> A quick test to see if your key is validating correctly, is to make a request to the `/auth/token` endpoint on our API.

Neither that page, [the pricing page](https://deepgram.com/pricing), nor [the index of the docs](https://developers.deepgram.com/llms.txt) says whether that request bills.

Revisit it when a Deepgram document says that the request bills nothing, and quote that sentence in the spec.
Status then declares `live` and `timeout`, and a server of the test binary sees exactly one request.
Only a 401 claims that Deepgram rejected the key.
A timeout or a 5xx says that the check reached no verdict, and never blames the key.

### Recovery of a lost final message

Parked from M13 on 2026-10-18, when Pascal cut it: the local recovery of a lost final protocol message.
Complete recorded artifacts would let Go record completion without another upload.
Until then, such a loss leaves the transcript step unknown, and a repeat refuses, as
[D038](../north-star/decisions/D038-python-protocol.md) records.

Revisit the transcript recovery when a lost final message strands a paid transcript.
