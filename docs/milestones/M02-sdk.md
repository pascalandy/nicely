# M02 SDK

Status: done
Version: v0.2.0

## Demo

A Go test extension built only on the public `sdk/` packages answers exactly like a core command, in English and in the pseudo-locale, both as `ncly-<domain>` and compiled into `ncly`.

## Scope

The SDK is the Go implementation of the contract between core and its extensions, as [SDK](../north-star/contract.md#sdk) defines. Core uses it too, so there is one implementation of each shared layer.

- Move the packages that extensions share from `internal/` to `sdk/`, under the names of the [code map](../north-star/architecture.md#code-map): the contract, i18n, config, tui, the platform helpers, the program runner of M01, and the command line of `internal/cli`. Core imports them from there. The host, core's commands and catalog, and later the keychain stay in `internal/`
- One entry point in `sdk/extension` serves both transports: a bundled extension runs as a child of the same `ncly` binary, as [D051](../north-star/decisions/D051-bundled-child-process.md) decides
- The SDK reads an extension's own codes from its manifest and its text from its catalogs, beside the SDK's catalog
- A lint rule forbids `sdk/`, `extensions/`, and the fixture to import `internal/`, so a bundled extension can leave the binary without a rewrite
- A first-party extension keeps everything it owns in `extensions/<name>/`: `manifest.toml`, `locales/`, `spec.md`, `SKILL.md`, `testdata/script/`, and its code. Its first consumer is `skill` in [M04](M04-skill.md)
- The Go fixture extension `hello` lives in `testdata/extensions/hello/`, outside `extensions/`, which is the official tap, and follows the same layout. It mirrors the shell fixture of M01, with an own code, an own warning, and a command that blocks until it is cancelled

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M02-T1 | Shared packages in sdk | agent | — | done |
| M02-T2 | Import boundary | agent | M02-T1 | done |
| M02-T3 | Extension entry point | agent | M02-T1 | done |
| M02-T4 | Bundled transport | agent | M02-T3 | done |

### M02-T1 Shared packages in sdk

- **Read:** [SDK](../north-star/contract.md#sdk), [code map](../north-star/architecture.md#code-map)
- **Proves:** `testdata/script/`, every existing scenario unchanged, and the catalog test of `sdk/i18n`

Move the shared packages to `sdk/` without changing behavior, and move the code of `internal/cli` that turns declarations into commands, reads the global flags, prints help, and prints answers into `sdk/command`. Split the catalog: the text that the SDK renders goes to `sdk/i18n/locales/`, and the text of core's commands to `internal/cli/locales/`, with the same IDs. Core's printer reads its own catalog beside the SDK's. The parity test walks every catalog folder and checks that core's catalog and the SDK's share no ID. Every existing scenario and test passes unchanged, and the **Proves** lines of done cards that name a moved test, such as M00-T2, M00-T3, M00-T4, and M01-T3, point to its new path in `sdk/`, so the plan check still finds every proof.

### M02-T2 Import boundary

- **Read:** [SDK](../north-star/contract.md#sdk), [AGENTS.md](../../AGENTS.md#code)
- **Proves:** `.golangci.yml`, and a temporary import of `internal/` from `sdk/` that fails the lint, shown in the pull request

Add a `depguard` rule that keeps the code of `sdk/`, `extensions/`, and `testdata/extensions/` away from `internal/`. The rule skips test files, with `!$test`, so the tests that compare code and spec keep reading the docs through `internal/specdoc` after M02-T1 moves them with their packages.

### M02-T3 Extension entry point

- **Read:** [SDK](../north-star/contract.md#sdk), [Manifest](../north-star/contract.md#manifest), [Text of an extension](../north-star/contract.md#text-of-an-extension), [Error codes](../north-star/contract.md#error-codes)
- **Proves:** `testdata/extensions/hello/testdata/script/sdk_extension.txtar`, and the test of `sdk/extension`

Build `sdk/extension` with the API of the spec, and the fixture `hello` on it, with its `manifest.toml`, its `locales/en.toml`, the own code `HELLO_GRUMPY`, the own warning `HELLO_UNNAMED`, a `report` command that fails itself with `CONFIG_INVALID` under `--fail`, a `text` command, and `hello wait`. That command starts a program through `sdk/runner`, which writes a finished file and starts a grandchild in its own process group, then waits until its context ends and returns a data key. `--stubborn` makes the handler ignore its context, and `--stubborn-child` makes the program ignore SIGTERM. `just lint` names `./testdata/extensions/...`. The test binary of `cmd/ncly` registers the fixture's external `main` as `hello-external`, never as `ncly-hello`, and the setup writes the script `ncly-hello` that runs it into the data folder of `$WORK`, with the manifest and catalog beside it. It runs the scenarios of `testdata/script/`, of each `extensions/*/testdata/script/`, and of each `testdata/extensions/*/testdata/script/`, with the same commands, so an extension's folder runs without editing the test.

Through `ncly`, the fixture answers a success, a success with `HELLO_UNNAMED` and exit 0, a usage error, `HELLO_GRUMPY`, a failed report on stdout with exit 78, the same report failing itself with `CONFIG_INVALID` on stderr, stdout empty, and exit 78, and its text under `--json`, each matching a core command's envelope, exit code, and error block, under `--json`, as text, and under the pseudo-locale, where the `pseudo` command accepts the fixture's codes. Run directly from the data folder, `ncly-hello run --json -- Ada` gives the same answer as `ncly hello run --json Ada`. Run directly too, with the signal sent to the extension alone, `hello wait` answers `TERMINATED` with its data key after `term`, which only the SDK's handling can give, the finished file stays, and the program and its grandchild are gone. `hello wait --stubborn-child` gives the same once the SDK's SIGKILL ends the program. `hello wait --stubborn` answers `TERMINATED` alone once the shortened cleanup ends, while a second signal changes nothing. The test of `sdk/extension` shows that `Check` rejects an own code without the domain prefix, a command that runs without a handler, a handler for no command, a catalog ID without the domain prefix, an entry without a description, and an `fr-CA` catalog that lacks an ID of `en`.

### M02-T4 Bundled transport

- **Read:** [SDK](../north-star/contract.md#sdk), [Discovery](../north-star/contract.md#discovery), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs)
- **Proves:** `testdata/script/bundled_extension.txtar`

`cmd/ncly` passes core an explicit list of bundled extensions, empty until `skill` joins it in M04, and its entry function runs a bundled extension's `Main` when started under that extension's name. The test binary registers `ncly-bundled` and `ncly-hello`, which both call that same entry function with a list that holds the fixture, so the bundled child reaches the fixture through the production routing of `argv[0]`, and the scenario fails when that routing is removed. The scenario runs the calls of M02-T3 through `ncly` and `ncly-bundled` and compares their stdout, stderr, and exit codes byte for byte. It sends `term` to each build in turn, one background call at a time, while `hello wait` runs, and both answer `TERMINATED` with exit 143 and the data key of `hello wait`, with the fixture gone. Under `ncly-bundled`, the bundled fixture wins over the external one, whose script writes a marker file that never appears, and a bundled name is reserved. A test of `cmd/ncly` calls `Check` on every bundled extension, so a broken one fails before release.
