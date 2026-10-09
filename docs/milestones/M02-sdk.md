# M02 SDK

Status: planned
Version: v0.2.0

## Demo

A Go test extension built only on the public `sdk/` packages answers exactly like a core command, in English and in the pseudo-locale, both as `ncly-<domain>` and compiled into `ncly`.

## Scope

The SDK is the Go implementation of the contract between core and its extensions. Core uses it too, so there is one implementation of each shared layer.

- Move the packages that extensions share from `internal/` to `sdk/`: the contract, i18n, config, and the platform helpers they need. Core imports them from there. The keychain stays in core, because an extension receives its keys in its environment
- An extension registers its own codes, each prefixed with its domain and mapped to one exit code, and its own catalog
- One entry point serves both transports: the same Go package runs as `ncly-<domain>` or compiles into `ncly` as a bundled extension, with the same manifest
- A lint rule forbids `sdk/` and `extensions/` to import `internal/`, so a bundled extension can leave the binary without a rewrite
- A first-party extension keeps everything it owns in `extensions/<name>/`: its manifest, `spec.md`, `SKILL.md`, catalogs, scenarios, and code

## Open questions

Settle each one in [Extensions](../north-star/cli-spec.md#extensions) and [AGENTS.md](../../AGENTS.md#code), then delete this section.

- The package names and the public API of `sdk/`, and which helpers stay inside core
- How `cmd/ncly` registers the bundled extensions, such as an explicit list rather than `init` side effects
- Where an extension's scenarios and fixtures live, such as `extensions/<name>/testdata/script/`, and how `just test` builds and runs them
- How an extension's catalog joins the parity test against `en`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M02-T1 | Shared packages in sdk | agent | — | todo |
| M02-T2 | Import boundary | agent | M02-T1 | todo |
| M02-T3 | Extension codes and catalogs | agent | M02-T1 | todo |
| M02-T4 | Bundled transport | agent | M02-T3 | todo |

### M02-T1 Shared packages in sdk

- **Read:** [Extensions](../north-star/cli-spec.md#extensions), [AGENTS.md](../../AGENTS.md#code)
- **Proves:** `testdata/script/`, every existing scenario unchanged

Move the shared packages to `sdk/` without changing behavior. Core imports them from there, and every existing scenario and test passes unchanged.

### M02-T2 Import boundary

- **Read:** [AGENTS.md](../../AGENTS.md#code)
- **Proves:** `.golangci.yml`, and a temporary import from `extensions/` into `internal/` that fails `just lint`, shown in the pull request

Add the lint rule that keeps `sdk/` and `extensions/` away from `internal/`.

### M02-T3 Extension codes and catalogs

- **Read:** [Error codes](../north-star/cli-spec.md#error-codes), [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/sdk_extension.txtar`

A Go fixture extension built on `sdk/` registers a prefixed code and its own catalog. Its answers match a core command's envelope, exit codes, and error block, and its text is marked under the pseudo-locale. A code without the domain prefix fails at registration.

### M02-T4 Bundled transport

- **Read:** [Extensions](../north-star/cli-spec.md#extensions)
- **Proves:** `testdata/script/bundled_extension.txtar`

The same fixture, compiled into a test build of `ncly`, answers exactly as its external build does. A bundled extension wins over an external one with the same name, and its name joins the reserved names.
