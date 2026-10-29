# Extension repository

This part says how the Git repository of an extension outside Nicely is laid out, created, and checked. The extension template creates it, as [D062](../decisions/D062-copier-template.md) decides. [Release and install](extensions.md#release-and-install) says how an author releases a version, and how a user installs and removes it.

The repository is one Go module. Its `go.mod` pins one Nicely version, as [Compatibility](compatibility.md) defines. It also repeats the `replace` of Lip Gloss from Nicely's `go.mod`, because a `replace` never reaches another module. That fork fixes a `tmux info` call as the package loads, until a release of Lip Gloss ships the fix. Its root holds the files of a first-party extension in `extensions/<name>/`, plus its own `README.md`, `vision.md`, `worker/`, and `AGENTS.md`:

| File | Holds |
|---|---|
| `README.md` | What the extension adds, and the commands that install a release |
| `vision.md` | The job of the extension in one sentence, the source and revision of each program that it wraps, and what it refuses |
| `spec.md` | One section per command, titled `## ncly <path>`: what proves success, why each effect exists, each failure and its code, and the accepted limits. Then `## Acceptance`: the real input, the properties that the output keeps, and each input that the wrapped program accepts but the domain refuses |
| `manifest.toml`, `locales/`, `skill/` | The manifest, the catalogs, and the skill, which the extension embeds as [SDK](sdk.md) defines |
| `worker/` | The scripts that it embeds as workers, one per command that wraps one |
| Go files | One handler per command, `extension_test.go`, which calls the test kit, `cmd/ncly-<slug>/main.go`, which calls `extension.Main`, and `internal/export/`, which calls `extension.Export` for `just install` |
| `testdata/script/` | The scenarios, which the [test kit](test-kit.md) runs |
| `AGENTS.md` | The reading order, the change steps, the recipes, and the release steps, for an agent that changes the extension |
| `.agents/skills/nicely-extension/` | The skill that takes a script from intake to a verified command, and reviews readiness when the user asks |
| `.copier-answers.yml`, `lefthook.yml` | The template version that `just update` reads, and the `just check` that Lefthook runs before each push |

`spec.md` holds only what the manifest cannot say, so it never repeats a flag or a type. [Extension identity](extension-identity.md) defines which commands it owns.

An agent creates the repository without prompts, with the Copier that the template pins. It passes a Nicely tag to `--vcs-ref`, then `--defaults`, and `--data` for `domain` and `module`, plus optional `name`.
An empty Copier name input omits the manifest `name` and keeps the whole-domain layout.
A named extension uses the slug for its executable, sidecars, skill, and config, as [Extension identity](extension-identity.md) defines. The bundled `nicely` skill shows the command. The agent then runs `go mod tidy` once, which writes `go.sum`, and commits the repository with Git. The root Go package keeps the domain as its name, including for named extensions.
The template refuses `main` and each Go keyword as a domain, such as `type`. It also refuses any slug that equals `nicely`, because the bundled skill would hide that extension skill. The template has no tasks and no migrations, so Copier needs no `--trust`. It ships no GitHub workflow, since `just` and Lefthook run the checks.

The repository tests itself through the [test kit](test-kit.md). Its `justfile` has five recipes:

- `just check` runs every check: `go mod tidy -diff`, `gofmt`, `go vet`, and `go test`, whose test kit installs the extension into scratch folders
- `just install` builds the executable and exports its sidecars from that build. It replaces the previous install of the slug in the extensions folder of [Discovery](discovery.md), the executable last
- `just uninstall` removes the install of the slug from that folder, the executable first. With nothing installed, it says so and exits 0
- `just update <version>` moves the repository to another Nicely version: the template first, then the SDK, on a clean Git tree. Copier can leave a conflict as markers, a `.rej` file, or an unmerged index entry, and still exit 0. On a conflict, the recipe stops before `go get` and prints the commands that finish the update. It never resets, stashes, or commits
- `just contract` prints the folder of this contract in the Go module of the pinned Nicely version, so an agent reads the exact rules offline

`just install` grants no key. A human grants one after reviewing the extension, as [Grants](grants.md) defines.
