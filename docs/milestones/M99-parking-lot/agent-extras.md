# Agent extras

Parked from M09 and M11: Claude, Pi, Grok, and OpenCode adapters; a Codex API key; named
profiles; provider selection; model catalogs; profile status screens; configurable system prompts; untrusted
`--input`; built-in code review; `--check-unchanged` snapshots and guards; and migration of the existing `headless` skill.
M09 and M10 keep one trusted task through the installed Codex CLI, native login, model and
effort settings, dry run, durable execution, inspection, and cancellation.

Revisit each extra for a demonstrated workflow that the current path cannot serve. A second
adapter must solve that need on its own. Model names in a brainstorm do not create an adapter
matrix. A tools-off consumer must first prove that its chosen runtime disables tools, as
[D028](../../north-star/decisions/D028-outside-text.md) requires; a filesystem check is no substitute.
The old multi-adapter M11 is removed from the delivery queue.

A Codex API key needs no new Nicely machinery: M08 already ships optional keys. PR #156
recorded that `codex exec` reads `CODEX_API_KEY`, so the adapter would map the granted `openai`
key to it. Reverify that flag before a card uses it.
