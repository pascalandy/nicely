# Extensions

Core is a small host, as [D004](../decisions/D004-minimal-core.md) decides.
[Extension identity](extension-identity.md) defines how extensions share a domain, their slugs, and their command roots. An extension comes in one of two transports, with one manifest format and one contract:

- A **bundled** extension is a Go package compiled into `ncly`, with its manifest and catalogs embedded. Only what an agent needs to operate Nicely is bundled: `skill`
- An **external** extension is an executable with adjacent sidecars, named by [Extension identity](extension-identity.md). Core finds it as [Discovery](discovery.md) defines. Taps, which would bring more, wait in [M99](../../milestones/M99-parking-lot/taps.md). The first-party ones, such as `agent` and `transcript`, live in `extensions/<name>/` in the Nicely repository, which is their official tap

Core commands and bundled extensions always win over an external extension with the same name. [ncly describe](../core-spec.md#ncly-describe) reports the source of each command that runs. [ncly doctor](../core-spec.md#ncly-doctor) adds where each extension comes from, and every extension that another one hides.

An extension receives the environment that [Programs that ncly runs](programs-that-ncly-runs.md) defines, which carries the global flags as variables. It answers with the envelope of [Output](output.md) and the exit codes of [Exit codes](exit-codes.md). It may use core's [error codes](error-codes.md), and its manifest declares its own. It translates its own text into the language of `NCLY_LANG`. It is written in Go with the SDK, the one supported way for now, as [D013](../decisions/D013-go-everywhere.md) decides. The wire contract stays open to any language, as [D040](../decisions/D040-wire-contract.md) decides. A program that the extension runs keeps its own language, as a worker of [Programs that ncly runs](programs-that-ncly-runs.md) does.

When a selected workflow composes extensions, it uses the CLI as an agent would. The shipped `skill` extension reads installed extension skill paths through `"$NCLY_BIN" describe --json`. Transcript summary remains in [M99](../../milestones/M99-parking-lot/transcript-summary.md). It writes its run records in the shared format of [Record format](record-format.md), through the SDK for a Go extension, so `ncly run` inspects every recorded operation alike. A Go extension imports only the public `sdk/` packages, so a bundled extension could leave the binary without a rewrite.

Each extension keeps its spec in `spec.md`: in `extensions/<name>/` for a first-party one, and at the root of its repository otherwise. It ships a skill named after its slug, which teaches an agent to use it. [ncly describe](../core-spec.md#ncly-describe) gives the folder of each external extension's skill, and the bundled `skill` extension ships none of its own, because the `nicely` skill that it serves teaches core and the skill domain.

## Extension repository

An author outside Nicely keeps an extension in a Git repository of its own, which the extension template creates, as [D062](../decisions/D062-copier-template.md) decides. The repository is one Go module. Its `go.mod` pins one Nicely version, as [Compatibility](compatibility.md) defines. It also repeats the `replace` of Lip Gloss from Nicely's `go.mod`, because a `replace` never reaches another module. That fork fixes a `tmux info` call as the package loads, until a release of Lip Gloss ships the fix. Its root holds the files of a first-party extension in `extensions/<name>/`, plus its own `vision.md`, `worker/`, and `AGENTS.md`:

| File | Holds |
|---|---|
| `vision.md` | The job of the extension in one sentence, the source and revision of each program that it wraps, and what it refuses |
| `spec.md` | One section per command, titled `## ncly <path>`: what proves success, why each effect exists, each failure and its code, and the accepted limits |
| `manifest.toml`, `locales/`, `skill/` | The manifest, the catalogs, and the skill, which the extension embeds as [SDK](sdk.md) defines |
| `worker/` | The scripts that it embeds as workers, one per command that wraps one |
| Go files | One handler per command, `extension_test.go`, which calls the test kit, `cmd/ncly-<slug>/main.go`, which calls `extension.Main`, and `internal/export/`, which calls `extension.Export` for `just install` |
| `testdata/script/` | The scenarios, which the [test kit](test-kit.md) runs |
| `AGENTS.md` | The reading order, the change steps, and the recipes, for an agent that changes the extension |
| `.agents/skills/nicely-extension/` | The skill that takes a script from intake to a verified command, and reviews readiness when the user asks |
| `.copier-answers.yml`, `lefthook.yml` | The template version that `just update` reads, and the `just check` that Lefthook runs before each push |

`spec.md` holds only what the manifest cannot say, so it never repeats a flag or a type. [Extension identity](extension-identity.md) defines which commands it owns.

An agent creates the repository without prompts, with the Copier that the template pins. It passes a Nicely tag to `--vcs-ref`, then `--defaults`, and `--data` for `domain` and `module`, plus optional `name`.
An empty Copier name input omits the manifest `name` and keeps the whole-domain layout.
A named extension uses the slug for its executable, sidecars, skill, and config, as [Extension identity](extension-identity.md) defines. The bundled `nicely` skill shows the command. The agent then runs `go mod tidy` once, which writes `go.sum`, and commits the repository with Git. The root Go package keeps the domain as its name, including for named extensions.
The template refuses `main` and each Go keyword as a domain, such as `type`. It also refuses any slug that equals `nicely`, because the bundled skill would hide that extension skill. The template has no tasks and no migrations, so Copier needs no `--trust`. It ships no GitHub workflow, since `just` and Lefthook run the checks.

The repository tests itself through the [test kit](test-kit.md). Its `justfile` has four recipes:

- `just check` runs every check: `go mod tidy -diff`, `gofmt`, `go vet`, and `go test`, whose test kit installs the extension into scratch folders
- `just install` builds the executable and exports its sidecars from that build. It replaces the previous install of the slug in the extensions folder of [Discovery](discovery.md), the executable last
- `just update <version>` moves the repository to another Nicely version: the template first, then the SDK, on a clean Git tree. Copier can leave a conflict as markers, a `.rej` file, or an unmerged index entry, and still exit 0. On a conflict, the recipe stops before `go get` and prints the commands that finish the update. It never resets, stashes, or commits
- `just contract` prints the folder of this contract in the Go module of the pinned Nicely version, so an agent reads the exact rules offline

`just install` grants no key. A human grants one after reviewing the extension, as [Grants](grants.md) defines. A user installs a fixed version with `git clone --branch <tag>`, then `just install`, which need Go and `just`. Taps, which would install without Go, wait in [M99](../../milestones/M99-parking-lot/taps.md).

## What a great extension does

Every extension meets this bar, a first-party one included. Each line links the parts that hold its rules, and names what checks it. A machine checks what it can. The author and the reviewer judge the rest.

1. **One job.** One sentence states it, in `vision.md` or at the start of a first-party spec. Each command is one bounded operation, never a relay of arbitrary arguments. Judgment
2. **`ncly describe` is enough to call it right.** The manifest declares every effect, program, and key, as [Manifest](manifest.md) defines. Checked by `extension.Check` and the scenarios
3. **A machine caller never waits and is never misled.** Machine output never asks, as [D046](../decisions/D046-machine-output-never-asks.md) decides, and success means that the work is proven ([Output](output.md), [Exit codes](exit-codes.md)). Checked by the scenarios, and by the test kit
4. **A dry run is a faithful preview** ([Global flags](global-flags.md), [Operations](operations.md)). Checked by the scenarios
5. **A failure says what to do next, and a repeat is safe or refused** ([Output](output.md), [Retry safety](retry-safety.md)). Checked by the scenarios
6. **It travels whole.** One build gives the executable, its sidecars, and its embedded workers ([SDK](sdk.md)). Checked by the scratch install of the test kit
7. **Its skill teaches when, not how.** `ncly describe` already lists the flags, as [D054](../decisions/D054-extension-skills-through-describe.md) decides. The test kit checks that the skill names only declared commands. The rest is judgment
8. **Each promise of its spec has a scenario.** The test kit checks that each command appears in a scenario. The rest is judgment
9. **It keeps the rules that the Go types cannot express**, which [SDK](sdk.md#rules-that-the-types-cannot-express) lists. The scenarios catch a broken answer or exit code. The rest is judgment
