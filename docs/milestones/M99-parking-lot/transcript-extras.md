# Transcript extras

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

## Live Deepgram check

Parked from M16 on 2026-10-19: `--live` on `ncly transcript status`, which would ask Deepgram whether it accepts the granted key.
No Deepgram document says that `GET /v1/auth/token` bills nothing.
[Authenticating](https://developers.deepgram.com/docs/authenticating) suggests the endpoint, and says nothing of its cost:

> A quick test to see if your key is validating correctly, is to make a request to the `/auth/token` endpoint on our API.

Neither that page, [the pricing page](https://deepgram.com/pricing), nor [the index of the docs](https://developers.deepgram.com/llms.txt) says whether that request bills.

Revisit it when a Deepgram document says that the request bills nothing, and quote that sentence in the spec.
Status then declares `live` and `timeout`, and a server of the test binary sees exactly one request.
Only a 401 claims that Deepgram rejected the key.
A timeout or a 5xx says that the check reached no verdict, and never blames the key.
