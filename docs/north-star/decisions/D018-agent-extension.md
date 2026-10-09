# D018 Run agents through the `agent` extension with one profile registry

Decided 2026-10-05, revised 2026-10-09. `agent` is a first-party external extension. `ncly agent run` runs a task through a harness, and the extension owns the profiles, the harness adapters, tools-off mode, and the depth limit, as [its spec](../../../extensions/agent/spec.md) defines. Profiles and dry run come in M09, execution in M10, and the other harnesses in M11. Transcript summarizes through `ncly agent run --json`, as any other extension would. A depth limit stops agents from launching agents without end. `ncly agent` absorbs the `headless` skill.

**Why.** Model IDs change every few months, so they belong in the config, not in code. One registry and one owner of harness behavior keep profiles, tool restrictions, cancellation, and summary recovery consistent. Composing through the CLI tests the same interface that agents use. A simple one-step command is the easiest first consumer of run records. "Harness" stays the internal term.

**Rejected.** "Infer", "chat", and "prompt", which mean little to most people or describe too little. Each project declaring its own inference settings. Python owning the summary. An internal agent service inside core that transcript calls in process, the first version of this entry, because core then grows a harness owner that it never needs to operate.
