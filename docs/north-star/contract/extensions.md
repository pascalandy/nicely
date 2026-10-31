# Extensions

Core is a small host, as [D004](../decisions/D004-minimal-core.md) decides.
[Extension identity](extension-identity.md) defines how extensions share a domain, their slugs, and their command roots. An extension comes in one of two transports, with one manifest format and one contract:

- A **bundled** extension is a Go package compiled into `ncly`, with its manifest and catalogs embedded. Only what an agent needs to operate Nicely is bundled: `skill`
- An **external** extension is an executable with adjacent sidecars, named by [Extension identity](extension-identity.md). Core finds it as [Discovery](discovery.md) defines. Taps, which would bring more, wait in [COR-622](../../backlog/COR-622-taps-offer.md#later). The first-party ones, such as `agent` and `transcript`, live in `extensions/<name>/` in the Nicely repository, which is their official tap

Core commands and bundled extensions always win over an external extension with the same name. [ncly describe](../core-spec.md#ncly-describe) reports the source of each command that runs. [ncly doctor](../core-spec.md#ncly-doctor) adds where each extension comes from, and every extension that another one hides.

An extension receives the environment that [Programs that ncly runs](programs-that-ncly-runs.md) defines, which carries the global flags as variables. It answers with the envelope of [Output](output.md) and the exit codes of [Exit codes](exit-codes.md). It may use core's [error codes](error-codes.md), and its manifest declares its own. It translates its own text into the language of `NCLY_LANG`. It is written in Go with the SDK, the one supported way for now, as [D013](../decisions/D013-go-everywhere.md) decides. The wire contract stays open to any language, as [D040](../decisions/D040-wire-contract.md) decides. A program that the extension runs keeps its own language, as a worker of [Programs that ncly runs](programs-that-ncly-runs.md) does.

When a selected workflow composes extensions, it uses the CLI as an agent would. The shipped `skill` extension reads installed extension skill paths through `"$NCLY_BIN" describe --json`. Transcript summary remains in [XT-102](../../backlog/XT-102-transcript.md#transcript-summary). It writes its run records in the shared format of [Record format](record-format.md), through the SDK for a Go extension, so `ncly run` inspects every recorded operation alike. A Go extension imports only the public `sdk/` packages, so a bundled extension could leave the binary without a rewrite.

Each extension keeps its spec in `spec.md`: in `extensions/<name>/` for a first-party one, and at the root of its repository otherwise. It ships a skill named after its slug, which teaches an agent to use it. [ncly describe](../core-spec.md#ncly-describe) gives the folder of each external extension's skill, and the bundled `skill` extension ships none of its own, because the `nicely` skill that it serves teaches core and the skill domain.

## Extension repository

An author outside Nicely keeps an extension in a Git repository of its own, which the extension template creates, as [D062](../decisions/D062-copier-template.md) decides. [Extension repository](extension-repository.md) holds its layout, its creation, and its recipes.

### Release and install

An author releases a version with `just release vX.Y.Z [fixture]`.
The SDK requires a clean HEAD equal to the fetched default branch of `origin`.
A successful signoff receipt must name that SHA, with exit 0 for every check and signoff.
Its checks must hold `just finish`, so a receipt with only `just check` never releases.
Release runs `just acceptance [fixture]`, then creates and pushes an annotated tag whose message holds that signoff receipt.
It executes the fenced `sh` block under `## Install` in `README.md`, with `TAG` replaced by the new tag.
The install uses scratch HOME and XDG folders, without the caller's keys or installed extensions.
The clone's HEAD must equal the tagged SHA, even when another README paragraph names the new tag.
Release then checks discovery, measures [Economy](../../qa/north-star.md#economy), uninstalls, and requires `NOT_FOUND`.
Economy runs literal manifest examples in both output formats, with a dry run when supported.
It selects these modes even when an example already sets them.
Each extension invocation extracts the files of `testdata/examples.txtar` from the tagged scratch clone.
It ignores archive setup commands and keeps extracted files off PATH.
An optional fixture replaces the first example argument that names an archive file, under that argument's filename.
Each measurement uses a fresh working folder, so previous output cannot change the next measurement.
Each measurement records the command, exit, elapsed milliseconds, and separate stdout and stderr byte counts.
Failed Economy calls and exceeded budgets are reservations, not release failures.
When available, `gh` reads the SHA's hosted check runs. Missing or failed hosted evidence becomes a note.
Release stores its receipt and command logs outside the repository, under the Nicely cache's `release/` folder.
It prints the receipt path, and its last success line names the tag and SHA.
A failed published install keeps the pushed tag and records the failure. Release never replaces an existing tag.

The receipt uses `tag`, `sha`, `os`, `signoff`, `pushed`, `cloned_sha`, `steps`, `economy`, `hosted_checks`, `notes`, `reservations`, and optional `failure`.
`signoff` holds the original signoff receipt. `pushed` records tag publication, and `cloned_sha` records the published clone's HEAD.
`hosted_checks` holds paginated check-run responses. `notes` records unavailable hosted evidence, and `reservations` records Economy findings.
`failure` holds the fatal diagnostic.
Each row in `steps` or `economy` holds `name`, `command`, `exit`, `elapsed_ms`, `stdout_bytes`, `stderr_bytes`, `stdout`, and `stderr`.
`stdout` and `stderr` name log files relative to the receipt. Scratch installs and measurement folders are removed after use.

A user installs a fixed version with `git clone --branch <tag> <repository>`, then `just install` in the clone, which need Go and `just`.
For an annotated tag, Git can warn that the tag `is not a commit!`, then note a detached HEAD: the clone holds the tagged commit.
`ncly describe <root> --json` then confirms the install.
The generated `README.md` gives these commands for its repository, and `AGENTS.md` gives the release steps.
Taps, which would install without Go, wait in [COR-622](../../backlog/COR-622-taps-offer.md#later).
A user removes an extension by moving its four files out of the extensions folder of [Discovery](discovery.md): `ncly-<slug>`, `ncly-<slug>.toml`, `ncly-<slug>.locales/`, and `ncly-<slug>.skill/`.
`just uninstall` in a clone does it, the executable first.
`ncly describe <root> --json` then answers `NOT_FOUND`, unless a folder of `PATH` holds another `ncly-<slug>`.
Removal keeps the keys and the grants, so an install at the same path has its grants again.
`ncly auth revoke <service> <slug>` takes a grant back.

## What a great extension does

Every extension meets this bar, a first-party one included. Each line links the parts that hold its rules, and names what checks it. A machine checks what it can. The author and the reviewer judge the rest.

1. **One job.** One sentence states it, in `vision.md` or at the start of a first-party spec. Each command is one bounded operation, never a relay of arbitrary arguments. Judgment
2. **`ncly describe` is enough to call it right.** The manifest declares every effect, program, and key, as [Manifest](manifest.md) defines. Checked by `extension.Check` and the scenarios
3. **A machine caller never waits and is never misled.** Machine output never asks, as [D046](../decisions/D046-machine-output-never-asks.md) decides, and success means that the work is proven ([Output](output.md), [Exit codes](exit-codes.md)). Checked by the scenarios, and by the test kit
4. **A dry run is a faithful preview** ([Global flags](global-flags.md), [Operations](operations.md)). Checked by the scenarios
5. **A failure says what to do next, and a repeat is safe or refused** ([Output](output.md), [Retry safety](retry-safety.md)). Checked by the scenarios
6. **It travels whole.** One build gives the executable, its sidecars, and its embedded workers ([SDK](sdk.md)). Checked by the scratch install of the test kit
7. **Its skill teaches when, not how.** `ncly describe` already lists the flags, as [D054](../decisions/D054-extension-skills-through-describe.md) decides. The test kit checks that the skill names only declared commands. The rest is judgment
8. **Each promise of its spec has a test.** A command that reads or writes user files starts its promises from the [generic cases](generic-cases.md). The test kit checks that each command appears in a scenario, and that a test claims each line of Failures and Acceptance ([Spec IDs](test-kit.md#spec-ids)). Whether that test proves its line is judgment
9. **It keeps the rules that the Go types cannot express**, which [SDK](sdk.md#rules-that-the-types-cannot-express) lists. The scenarios catch a broken answer or exit code. The rest is judgment
