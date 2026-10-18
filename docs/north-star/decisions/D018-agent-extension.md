# D018 Run agents through the `agent` extension

Decided 2026-10-05, revised 2026-10-09. A first-party external extension owns `ncly agent run`.

**Scope update, 2026-10-17.** [M09](../../milestones/M09-agent-profiles.md) and [M10](../../milestones/M10-agent-run.md) deliver one trusted Codex task. [The spec](../../../extensions/agent/spec.md) owns invocation, tools permission, cancellation, and depth. [M99](../../milestones/M99-parking-lot.md#agent-extras) defers other harnesses, profiles, and `headless` migration. [Transcript summary](../../milestones/M99-parking-lot.md#transcript-summary) requires a verified tools-off path under [D028](D028-outside-text.md).

**Why.** One owner keeps harness behavior consistent. Model IDs belong in configuration because they change. CLI composition preserves the interface agents use. One task first proves run records.

**Rejected.** A core agent service, Python-owned summary, and project-specific inference settings. The names "infer", "chat", and "prompt" obscure the task. "Harness" stays internal.
