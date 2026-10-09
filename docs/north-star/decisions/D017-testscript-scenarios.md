# D017 Write acceptance criteria as testscript scenarios

Decided 2026-10-05. Each card names the testscript scenarios that prove it. End-to-end scenarios come before unit tests. M00 verifies the real parser, streams, modes, and shared outcome logic, including partial results. A later domain proves its effects, simulation, and recovery on its first real implementation.

**Why.** Pascal writes the specifications and agents write the Go. A scenario reads like a terminal session, so Pascal checks behavior without reading Go. Parser errors and misplaced output can make a command unusable by agents even when its success path works.

**Rejected.** Treating a contract fixture as proof that a future command avoids writes or duplicate paid requests. Stubbed services keep tests repeatable, but they do not prove that the real Python program and harness support the required protocol and restrictions.
