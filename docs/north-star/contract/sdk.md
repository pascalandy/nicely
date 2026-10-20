# SDK

The SDK is the Go implementation of this contract, in the public packages of `sdk/` that the [code map](../architecture.md#code-map) lists, as [D040](../decisions/D040-wire-contract.md) decides. Core uses it too, so each shared layer has one implementation. A Go extension imports `sdk/` only, never core's `internal/`. Go refuses that import from another module, and `just lint` refuses it inside Nicely's.

A Go extension is one package that exports one `extension.Extension` value, built from `manifest.toml` and `locales/` in its folder with `go:embed`. The API of `sdk/extension` comes down to this:

```go
type Extension struct {
	Manifest  []byte             // manifest.toml
	Locales   fs.FS              // holds locales/<lang>.toml
	Skill     fs.FS              // holds skill/SKILL.md and the files it names
	Resources fs.FS              // holds worker/, the workers that the extension embeds, from M18-T4
	Handlers  map[string]Handler // one per command that runs, keyed by path, such as "hello run"
}

// Handler runs one validated call. ctx ends on the first Ctrl-C or SIGTERM.
type Handler func(ctx context.Context, call Call) contract.Outcome

type Call struct {
	Path    string              // such as "hello run"
	Args    []string            // the positional arguments
	Flags   map[string]string   // each flag that the caller set, by long name, except a variadic one; a switch holds "true" or "false"
	Lists   map[string][]string // each variadic flag that the caller set, by long name, with its values in the caller's order
	JSON    bool
	Mode    contract.Mode
	Printer *i18n.Printer // in the language of NCLY_LANG
	Stdin   io.Reader
	Stdout  io.Writer // human text only; the SDK prints the answer
	Stderr  io.Writer
}

// Key returns the key of a service that the command declares, or the problem that NCLY_MISSING_KEYS names for it
func (c Call) Key(service string) (string, *contract.Problem)

func Main(e Extension) // runs the extension as this process, then exits with its code
func Check(e Extension) error
func Export(e Extension, dir string) error // writes the sidecars of e into dir, from M18-T3

// ParseSkill reads the frontmatter of a SKILL.md: its name and description, both required
func ParseSkill(data []byte) (Frontmatter, error)
```

`Main` parses the call with the manifest, prints the help of a group, and calls the handler of the path. A variadic flag reaches the handler in `Lists` alone, with its values in the caller's order, so `Flags` keeps one string for each other flag, and no handler reads one value of a variadic flag as its only one. `Call.Key` takes a service that the command declares, and returns its key or the problem that `NCLY_MISSING_KEYS` names for it. That problem has the hint and `next` of [Grants](grants.md), so a handler returns it and never builds one. A service whose variable is empty and that `NCLY_MISSING_KEYS` does not name, as when `ncly-<domain>` runs directly, gets `AUTH_MISSING`, and a service that the command does not declare gets `RUNTIME`, because core never passes its key. `Main` prints the result by its kind, from the command's `result`:

- `object` and `batch`: in human mode, the handler writes its text to `Stdout`. With `--json`, it writes nothing there, and the SDK prints the answer from the outcome
- `report`: in human mode, the handler writes its report to `Stdout`. With `--json`, the SDK prints the findings from the outcome's data on stdout, in an answer without `errors`, as [Output](output.md) says. In both modes, the handler returns its findings in that data, and sets the outcome's `HumanNeeded` when a finding needs a human, which makes the exit code 78 instead of 0, and `ok` false with `--json`. Core's doctor uses the same outcome
- `text`: the handler writes its text to `Stdout`, whatever `--json` says

A failure of the command itself prints as an ordinary failure for every kind: the error block, or the answer on stderr with `--json`. On Ctrl-C or SIGTERM, the SDK ends `ctx`, stops the extension's programs as [Programs that ncly runs](programs-that-ncly-runs.md) defines, and sets the outcome's signal, so a handler returns what it has. A handler that has not returned once the 15 seconds of cleanup and the 2 seconds of the drain limit have passed, the time that the SIGKILL of cleanup gives its programs to end, gets no more time: the SDK answers `INTERRUPTED` or `TERMINATED` alone and exits. Under `ncly`, core's own 15 seconds govern, and either path gives the same answer, because both carry the signal's code and no data key. Run directly, `ncly-<domain>` also parses the global flags and the ordinary syntax that core accepts, because canonical arguments are a subset of it, and `--version` prints the version of the module that built it. `Check` reports a manifest that breaks the rules of [Manifest](manifest.md), a command that runs without a handler, a handler for no such command, a catalog that breaks the catalog rules below, and a skill whose `name` is not the domain. Every Go extension has a test that calls it. `Check` accepts an extension without a skill, because the bundled `skill` extension ships none of its own. From M18-T3, the [test kit](test-kit.md) requires a skill named after the domain of every external extension. `Main` checks the manifest and the handlers when it starts and answers `RUNTIME` for those faults, while a missing translation falls back to `en` at run time, as [Text of an extension](text-of-an-extension.md) says.

From M18-T4, `Resources` holds the extension's workers, which [Programs that ncly runs](programs-that-ncly-runs.md) defines. For each run that uses a worker, the SDK extracts it into a private temporary folder, and deletes that folder once the handler returns. The handler runs the worker by its absolute path, as [Wrapped programs](wrap.md) defines.

The same value serves both transports, so they cannot drift, as [D051](../decisions/D051-bundled-child-process.md) decides:

- **External:** the extension's `main` calls `extension.Main`. Installing it writes the embedded manifest, catalogs, and skill beside the executable, as `ncly-<domain>.toml`, `ncly-<domain>.locales/`, and `ncly-<domain>.skill/`. From M18-T3, `Export` writes them from the same value, so an install that uses it never copies them from the source tree
- **Bundled:** `cmd/ncly` passes the value to core in an explicit list, never through an `init` side effect. Core reads its embedded manifest and catalogs, validates its call, then starts the `ncly` binary itself as the extension's program, with `ncly-<domain>` as its first argument, `argv[0]`, through the program runner. Started under the name of a bundled extension, `ncly` runs that extension's `Main` instead of core, as busybox picks its applet from its name. A bundled extension therefore gets the canonical arguments, the environment, the process group, the cancellation, and the answer check of an external one

The SDK accepts the error codes that the manifest declares, and every warning code that starts with the domain and an underscore, beside core's codes, so the verdict and the answer follow [Output](output.md). The SDK's catalog holds the text that the SDK renders: help pages, global flags, usage errors, error labels, outcome summaries, size units, and the error of an extension that does not match its own manifest. An extension's printer takes an ID that starts with its domain and a dot from the extension's catalogs, and any other ID from the SDK's, so an extension never overrides the SDK's text. Core's printer takes an ID from core's catalog when that catalog holds it, and from the SDK's otherwise. The catalog rules are these: every entry has a description, every language holds exactly the IDs of `en`, an extension's IDs start with its domain and a dot, and core's IDs differ from the SDK's. In each language, an entry whose `other` form reads a field named `Count` or `N`, which holds a number, also has a `one` form that is not empty. `Check` applies them to an extension. In the Nicely repository, one parity test also applies them to every catalog folder: core's, the SDK's, and each extension's.

## Rules that the types cannot express

The Go types of `sdk/extension` cannot enforce these rules, so each extension keeps them:

- `main` and `init` start nothing before `extension.Main`: no program, no request, and no write. `--help` and `--version` therefore stay safe
- When `ctx` ends, a handler returns, and leaves no program running, as [Programs that ncly runs](programs-that-ncly-runs.md) defines
- A handler reports a failure through its outcome. It never prints an answer, exits, or panics
- With `--json`, a handler of an `object` or a `batch` command writes nothing to `Stdout`, as `Main` above says
- A handler starts each program through `sdk/runner`, or through `sdk/wrap` from M18-T4, never through `os/exec` alone
- A handler calls `Call.Key` only on the path that needs the key, never in a dry run, which receives no key, as [Grants](grants.md) defines. It passes the key only to the program that its spec names
- A dry run shares preparation with execution and starts no effect, as [Global flags](global-flags.md) defines
