## Planned commands

Nicely is a small core that extensions add commands to, in the spirit of Pi.
Core brings `describe`, `doctor`, `auth`, and `run`. It bundles `skill`, which agents use to learn Nicely.
External extensions supply other commands under the [extension contract](../docs/north-star/contract/extensions.md).

The first `agent` command runs one trusted task through Codex. The first `transcript` command transcribes one explicitly selected YouTube video. A checkout installs the extensions with `just install`. Each capability starts with its smallest useful path; additional harnesses, transcript summaries, and taps remain in the [backlog](../docs/backlog/README.md).

Named extensions share domain help, completion, and describe, such as `video convert` and `video subtitle`.
Read their skills by slug, such as `ncly skill view video-convert`.
Configuration, doctor components, and key grants use each extension's slug independently.
For example, `ncly doctor video-subtitle` checks that extension, and `ncly auth grant deepgram video-subtitle` grants only its key access.
Existing whole-domain grants keep working without migration.
The template creates named repositories with `--data domain=video --data name=convert`, such as `ncly-video-convert`.
Omit `name` for a whole-domain repository.
Both use the SDK test kit and the same update recipe.
The kit runs manifest examples with scratch fixtures and checks each answer.
It uses dry runs where available and prints skip reasons in plain `just check` output.
`just check` also runs the [analyzers](../docs/north-star/contract/analyzers.md) of the pinned SDK, such as the refusal of `io.Discard` in a `runner.Program`.
The kit, `just signoff`, `just release`, and `just acceptance` run in [owned runs](../docs/north-star/contract/owned-runs.md), without settings of the caller such as `NCLY_CONFIG`.
`just verify-models` checks the TLA+ models of `verification/tlc.json` with a pinned TLC, whose SHA-256 it checks before Java starts. `just check` reads only their [record](../docs/north-star/contract/formal-models.md), with no Java and no network.
`just update VERSION` runs the target version's updater.
It removes the caller's `GIT_*` settings from update programs, including Copier's Git calls, so a hook cannot redirect the update to another repository.
Older repositories use the bootstrap in [Update an extension](../docs/north-star/contract/extension-repository.md#update-an-extension).
Further work lives in [docs hub](../docs/backlog/XT-104-docs-hub.md), [Canadian French](../docs/backlog/COR-102-canadian-french.md),
[Markdown view](../docs/backlog/XT-103-markdown-view.md), and [video and image jobs](../docs/backlog/XT-602-video-image-jobs.md).

In a terminal, a form asks a human for each missing value. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

