# D011 Check locally by default, test live on request

Decided 2026-10-05. `ncly doctor` stays on the machine by default. `--live` calls free endpoints only. Invalid config is a failed report finding on stdout, and independent checks still run. An install runs only after a yes in a terminal.

**Why.** Doctor must be safe to run at any time. A live check proves a key works without a bill.
