# Decisions

Each entry records what was decided, why, and what was rejected. An entry states its decision briefly, and [principles.md](../principles.md), [contract.md](../contract.md), [core-spec.md](../core-spec.md), and the extension specs hold the full rules.

Until v0.0.1 ships, fix an entry in place, because git keeps the old text. From v0.0.1, add a new entry that names the entry it replaces, and mark the old entry `Replaced by Dxxx`.

| Decision | Title |
|---|---|
| [D001](D001-go-cobra-charm.md) | Build from scratch in Go with Cobra and Charm |
| [D002](D002-nicely-and-ncly.md) | Name the project Nicely and the command ncly |
| [D003](D003-domain-first.md) | Put the domain first and keep one verb per action |
| [D004](D004-minimal-core.md) | Keep core minimal and make every domain an extension |
| [D005](D005-extensions-and-taps.md) | Add extensions the Git way and carry them in taps |
| [D006](D006-agents-through-the-cli.md) | Serve agents through the CLI itself |
| [D007](D007-stable-agent-contract.md) | Keep one stable agent contract |
| [D008](D008-one-confirmation-flag.md) | Use one flag for every confirmation |
| [D009](D009-ask-in-a-terminal.md) | Ask in a terminal, fail elsewhere |
| [D010](D010-keychain-or-environment.md) | Keep keys in the keychain or the environment |
| [D011](D011-local-checks.md) | Check locally by default, test live on request |
| [D012](D012-translatable-from-day-one.md) | Make every word of core translatable from day one |
| [D013](D013-go-everywhere.md) | Write core and the first-party extensions in Go |
| [D014](D014-public-from-day-one.md) | Publish from the first commit |
| [D015](D015-macos-and-linux.md) | Target macOS and Linux with the same paths |
| [D016](D016-local-ci.md) | Run CI locally and release by hand |
| [D017](D017-testscript-scenarios.md) | Write acceptance criteria as testscript scenarios |
| [D018](D018-agent-extension.md) | Run agents through the `agent` extension with one profile registry |
| [D019](D019-transcript-extension.md) | Ship transcript as a first-party extension |
| [D020](D020-host-first.md) | Build the extension host before any domain |
| [D021](D021-skill-requirements-file.md) | Move skill requirements to a file |
| [D022](D022-private-docs.md) | Keep docs private by default |
| [D023](D023-shared-and-local-config.md) | Configure with a shared TOML file and a local one |
| [D024](D024-dynamic-completions.md) | Ship completions with dynamic values |
| [D025](D025-docs-layout.md) | Split the docs by question, and work in milestones |
| [D026](D026-english.md) | Write the repository in English |
| [D027](D027-json-answer.md) | Answer in one JSON line with `ok` and `errors` |
| [D028](D028-outside-text.md) | Keep outside text away from agent tools |
| [D029](D029-decide-early.md) | Decide early what is costly to change |
| [D030](D030-trusted-extensions.md) | Treat extensions as trusted code and grant keys one by one |
| [D031](D031-cancel-process-tree.md) | Stop the whole process tree on cancel and keep finished work |
| [D032](D032-run-records.md) | Keep small run records and resume explicitly |
| [D033](D033-one-declaration.md) | Describe a command once for humans and agents |
| [D034](D034-shared-preparation.md) | Use the same preparation for simulation and execution |
| [D035](D035-sync-ownership.md) | Track ownership when synchronizing files |
| [D036](D036-run-record-file.md) | Record each run in one JSON file that its executor locks |
| [D037](D037-summary-operation.md) | Summarize a saved transcript as a new operation |
| [D038](D038-python-protocol.md) | Exchange acknowledged JSON Lines with the Python program |
| [D039](D039-yt-dlp-updates.md) | Let yt-dlp update without a release |
| [D040](D040-wire-contract.md) | Serve extensions through a wire contract and a public Go SDK |
| [D041](D041-first-party-extensions.md) | Keep first-party extensions in the Nicely repository |
| [D042](D042-milestones-and-cards.md) | Plan work as milestones of small cards |
| [D043](D043-milestone-versions.md) | Number milestones by the release that ships them, and release separately |
| [D044](D044-additive-extensions.md) | Let extensions add commands, never change core |
