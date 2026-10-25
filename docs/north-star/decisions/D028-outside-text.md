# D028 Keep outside text away from agent tools

Decided 2026-10-05. Outside text reaches agents only with tools disabled. An adapter must enforce that mode or refuse. Trusted tasks receive tools only through explicit `--tools`.

**Scope update, 2026-10-17.** [The Codex path](../../../extensions/agent/spec.md#ncly-agent) requires `--tools` because it cannot enforce tools-off. It accepts no outside text, including through `--task-file`. [Summary](../../backlog/XT-102-transcript.md#transcript-summary) waits for verified tools-off support.

**Why.** A transcript or web page can contain instructions. The original transcript CLI used tool-free `claude --print` for this reason.

**Rejected.** Silent tools permission and post-run detection presented as write prevention or `--read-only`. The deferred `--check-unchanged` detector cannot satisfy this rule.
