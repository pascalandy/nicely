# Extension repository

This part says how the Git repository of an extension outside Nicely is laid out, created, and checked. The extension template creates it, as [D062](../decisions/D062-copier-template.md) decides. [Release and install](extensions.md#release-and-install) says how an author releases a version, and how a user installs and removes it.

[Version pins](extension-repository/versions.md) defines the repository's Go module and dependency pins.

Its root holds the files of a first-party extension in `extensions/<name>/`, plus its own `README.md`, `vision.md`, `worker/`, and `AGENTS.md`:

| File | Holds |
|---|---|
| `README.md` | What the extension adds, and the commands that install a release |
| `vision.md` | The job of the extension in one sentence, the source and revision of each wrapped program, the view of its intake, and what it refuses |
| `intake.toml`, `review.toml` | The [build records](build-records.md): the facts of the intake, and the last readiness review with the ledger of the builds |
| `spec.md` | One section per command, titled `## ncly <path>`: what proves success, why each effect exists, each failure and its code, and the limits that the user accepted. Then `## Acceptance`: the real input, the properties that the output keeps, and each input that the wrapped program accepts but the domain refuses. Then `## Open lines`: the lines that no test proves yet. Each line of Failures and of Acceptance starts with its spec ID, as [Test kit](test-kit.md#spec-ids) says |
| `manifest.toml`, `locales/`, `skill/` | The manifest, the catalogs, and the skill, which the extension embeds as [SDK](sdk.md) defines |
| `worker/` | The scripts that it embeds as workers, one per command that wraps one |
| Go files | One handler per command, `extension_test.go`, which calls the test kit, and `cmd/ncly-<slug>/main.go`, which calls `extension.Main`. `internal/export/` calls `extensionrepo.Main` for repository tasks |
| `testdata/script/` | The scenarios, which the [test kit](test-kit.md) runs |
| `qa.md` | How a user accepts the extension: what its UAT adds to the UAT protocol of Nicely, with its error cases and its removal. Each criterion names the test that covers it, or why none can |
| `acceptance.sh` | The run of the command on the real input of `## Acceptance`, which checks each property and each refusal that it lists |
| `AGENTS.md` | The reading order, the change steps, the recipes, and the release steps, for an agent that changes the extension |
| `AGENTS/rule-index.md` | The check that enforces each rule of `AGENTS.md`, with its parts in `AGENTS/rule-index/` |
| `.agents/skills/nicely-extension/` | The skill that builds each command in rounds, from intake to a verified command, with a readiness review at the end of each round |
| `readiness.md` | The view of `review.toml`, which `just records` writes |
| `.copier-answers.yml`, `lefthook.yml` | The template version that `just update` reads, and the `just check` that Lefthook runs before each push |
| `.github/workflows/check.yml` | The `just check` that GitHub runs on Ubuntu, plus macOS for public repositories, for each push to `main` and each pull request |

`spec.md` holds only what the manifest cannot say, so it never repeats a flag or a type. [Extension identity](extension-identity.md) defines which commands it owns.

Each criterion has one normative place. A failure, a property, or a refusal is one line of `spec.md`, with its spec ID. `qa.md` names that ID instead of repeating the line, and a test claims it. A limit never contradicts a line of Failures or of Acceptance. An input that the wrapped program accepts and the domain refuses is a refusal of `## Acceptance`, never a limit.

An agent creates the repository without prompts, with the Copier that the template pins. It passes a Nicely tag or commit to `--vcs-ref`, then `--defaults`, and `--data` for `domain` and `module`, plus optional `name`.
An empty Copier name input omits the manifest `name` and keeps the whole-domain layout.
A named extension uses the slug for its executable, sidecars, skill, and config, as [Extension identity](extension-identity.md) defines. The bundled `nicely-extensions` skill shows the command. The agent then runs `go mod tidy` once, which writes `go.sum`, and commits the repository with Git. The root Go package keeps the domain as its name, including for named extensions.
The template refuses `main` and each Go keyword as a domain, such as `type`. It also refuses the domain `nicely`, with or without a name, because `ncly` bundles the skill `nicely` and skills named `nicely-<topic>`. These would hide the skill of the extension. The template has no tasks and no migrations, so Copier needs no `--trust`.
A push can skip Lefthook, so its GitHub workflow runs `just check` again on `ubuntu-24.04`.
Public repositories also run the same steps on `macos-15`, and both runners are free.

The repository tests itself through the [test kit](test-kit.md). Its `justfile` has twelve task recipes, plus `default`:

- `just check` runs every check: `go mod tidy -diff`, `gofmt`, `go vet`, the [analyzers](analyzers.md) of the pinned SDK, and `go test`, whose test kit installs the extension into scratch folders. With `verification/tlc.json`, it first reads the record of the [models](formal-models.md). It first prints how many rules of the rule index are `Prose only`
- `just finish` is the exit of the loop of the skill, and `just check` stays green for an incomplete round. [Finish gate](finish-gate.md) lists what it checks
- `just records` writes the views of the [build records](build-records.md#views): `## Intake` of `vision.md`, and `readiness.md`
- `just signoff` runs `just finish` and `just check` on the committed SHA in isolation, then signs that SHA, as [Local checks](../decisions/D016-local-ci.md) defines. After `just finish`, it runs the first failing run again, as [Finish gate](finish-gate.md#the-first-failing-run-again) says
- `just verify-models [jar]` checks the TLA+ models with the pinned TLC, and writes their record, as [Formal models](formal-models.md) says
- `just scenario <name>` runs the scenario `testdata/script/<name>.txtar` alone, and the test kit skips its coverage checks
- `just release <tag> [fixture]` publishes the signed SHA and proves the README install, as [Release and install](extensions.md#release-and-install) defines
- `just install` builds the executable and exports its sidecars from that build. It replaces the previous install of the slug in the extensions folder of [Discovery](discovery.md), the executable last
- `just uninstall` removes the install of the slug from that folder, the executable first. With nothing installed, it says so and exits 0
- `just acceptance [fixture]` runs `acceptance.sh` as an agent would, with `--json` and `--no-input`, in an [owned run](owned-runs.md) without the Nicely settings of the caller. It installs the extension and the `ncly` of the pinned Nicely version into scratch XDG folders. It works on a copy of the fixture and checks that the fixture keeps its checksum. Without a fixture, it makes an input
- `just update <version>` runs the target version's updater on a clean Git tree, with the template before the SDK. [Update an extension](update-an-extension.md) defines bootstrap and conflict recovery
- `just contract` prints the folder of this contract in the Go module of the pinned Nicely version, so an agent reads the exact rules offline

`just install` grants no key. A human grants one after reviewing the extension, as [Grants](grants.md) defines.

The SDK package `extensionrepo` owns install, uninstall, finish, records, signoff, and release.
The template's `internal/export/` program calls `extensionrepo.Main(Extension)`.
The recipes each call that program in one line: `go run ./internal/export install` or `go run ./internal/export uninstall`.
Signoff calls `go run ./internal/export signoff`.
Release calls `go run ./internal/export release <tag> [fixture]`, and acceptance calls `go run ./internal/export acceptance [fixture]`.
Each run of these programs builds its environment as [Owned runs](owned-runs.md) defines, and returns one result.
Its `export <folder>` subcommand writes the embedded sidecars of the staged build.

Install stages both builds beside the destination with one `go build`.
The staged export program writes the sidecars from that build.
After preparation succeeds, install moves the old catalog and skill folders aside.
It publishes the manifest, catalogs, and skill before the executable, with one rename per file or folder.
A failed sidecar publication keeps the previous executable, or leaves no executable on a first install.
The stage goes when the program returns.
This order does not make replacement of all sidecars atomic.

Uninstall moves the executable into a stage first, then the manifest, catalogs, and skill.
It removes dangling symlinks too.
Without an install, it exits 0.
Both tasks use the data folder of [Configuration](configuration.md), with an absolute `HOME`.

The test kit calls the same install and uninstall package on scratch folders.
It proves that a failed sidecar publication exposes no executable.
Each scenario ends with uninstall as cleanup.
Once per suite, a separate scratch install holds only the extension under test.
The kit installs and uninstalls that extension there.
Then `ncly describe <root> --json` must exit 2 with `NOT_FOUND`.
Extensions that a scenario adds cannot affect this lifecycle check.

## Update an extension

[Update an extension](update-an-extension.md) holds the bootstrap and the conflict recovery of `just update`.
