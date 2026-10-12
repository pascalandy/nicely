# D028 Keep outside text away from agent tools

Decided 2026-10-05. Text from outside Nicely, such as a transcript or a web page, reaches an agent only with the agent's tools turned off. The adapter must support and enforce this mode or refuse the call. `ncly agent run` keeps tools off unless a call without material asks for them with `--tools`. The planned `--check-unchanged` flag detects changes after a run and makes no promise to prevent writes.

**Why.** A video can carry instructions. The transcript CLI already summarizes with a tool-free `claude --print` for that reason. With tools off by default, a forgotten flag never hands tools to an agent, and a caller can check the tools-off mode in a dry run before its material exists.

**Rejected.** Calling a post-run check `--read-only`. Relying on that check to prevent writes, or silently allowing tools when a harness cannot turn them off.
