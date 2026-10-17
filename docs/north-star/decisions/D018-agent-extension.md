# D018 Run agents through the `agent` extension

Decided 2026-10-05, revised 2026-10-09. Agent execution belongs in a first-party external extension, reached through `ncly agent run`. The original plan combined named profiles, several harness adapters, tools-off mode, and transcript summary.

**Scope update, 2026-10-17.** [M09](../../milestones/M09-agent-profiles.md) prepares one trusted Codex task; [M10](../../milestones/M10-agent-run.md) executes and records it. The extension uses native Codex login and model and effort settings, without named profiles or provider selection. [Its spec](../../../extensions/agent/spec.md) defines the invocation, explicit tools permission, cancellation, and depth limit.

Additional harnesses, profiles, and migration of the `headless` skill remain in [M99](../../milestones/M99-parking-lot.md#agent-extras). [Transcript summary](../../milestones/M99-parking-lot.md#transcript-summary) also remains deferred. It must have a verified tools-off path under [D028](D028-outside-text.md) before it uses this CLI. The extension boundary remains available for future composition without making those consumers part of the first agent capability.

**Why.** Model IDs change, so they belong in configuration, not code. One owner keeps harness behavior, cancellation, and depth limits consistent. Composing through the CLI uses the same interface that agents use. A one-step task is the first consumer of run records. "Harness" stays the internal term.

**Rejected.** "Infer", "chat", and "prompt", which mean little to most people or describe too little. Each project declaring its own inference settings. Python owning the summary. An internal agent service inside core that transcript calls in process, the first version of this entry, because core needs no harness owner to operate.
