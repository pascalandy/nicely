# D017 Write acceptance criteria as testscript scenarios

Decided 2026-10-05. Prefer end-to-end testscript scenarios. Each card names its proof under [Work on a card](../../../AGENTS.md#work-on-a-card). Recorded operations include a kill at every step boundary through a test-only hook.

**Why.** Pascal can review terminal behavior without reading Go. Parser and stream failures matter beyond the success path. Systematic boundaries catch crashes that hand-picked points miss, as cancellation injection does in [Asupersync](https://github.com/Dicklesworthstone/asupersync).

**Rejected.** Fault injection in release binaries. Shared fixtures cannot prove a later command's effects or retry safety. Stubs cannot prove the real Python or harness protocol.
