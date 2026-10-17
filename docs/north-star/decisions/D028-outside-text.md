# D028 Keep outside text away from agent tools

Decided 2026-10-05. Text from outside Nicely, such as a transcript or a web page, reaches an agent only with its tools turned off. The adapter must enforce this mode or refuse the call. A call must explicitly request `--tools` before a trusted task receives them. A post-run change check never promises to prevent writes.

**Scope update, 2026-10-17.** The first [Codex path](../../../extensions/agent/spec.md#ncly-agent) cannot enforce tools-off mode. It therefore requires explicit `--tools` and accepts only trusted task instructions. Omitting that flag returns `CAPABILITY_UNSUPPORTED`, including during dry run. It exposes no `--input`, and `--task-file` is not a route for outside text.

[Transcript summary](../../milestones/M99-parking-lot.md#transcript-summary) stays deferred until a tools-off path is verified. The proposed `--check-unchanged` detector remains with [agent extras](../../milestones/M99-parking-lot.md#agent-extras); it cannot satisfy this rule.

**Why.** A video can carry instructions. At the original decision, the transcript CLI used a tool-free `claude --print` for summaries. Refusal when tools-off cannot be enforced keeps a forgotten flag from silently granting tools. A dry run checks the requested mode before execution.

**Rejected.** Calling a post-run check `--read-only`. Relying on that check to prevent writes, or silently allowing tools when a harness cannot turn them off.
