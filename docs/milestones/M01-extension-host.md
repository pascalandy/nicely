# M01 Extension host

Status: planned
Version: v0.1.0

## Demo

`ncly hello` runs a test extension written in shell from its manifest, and `ncly` checks the answer.

## Scope

Core finds executables named `ncly-<domain>`, reads their manifests without running them, starts them, and checks what they answer. This milestone uses fixture extensions only. The first bundled extension, `skill`, arrives in [M03](M03-skill.md), and the first real external one, `agent`, in [M09](M09-agent-profiles.md).

- Discovery looks in `~/.local/share/nicely/extensions/` first, then in `PATH` order. Dispatch requires a manifest with a compatible protocol. A missing or incompatible manifest is reported without starting the executable
- Core commands and bundled extensions always win over an external extension with the same name. The explicit `reservedNames` of `internal/cli/root.go` stays, because it also reserves names whose command arrives later, and a test keeps it equal to the [domain table](../north-star/architecture.md#domains-and-commands)
- An extension receives the environment that [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs) defines. No key reaches it until grants arrive in [M08](M08-grants.md)
- With `--json`, core checks the answer: one JSON line, `ok`, a compatible `contract_version`, codes from core's registry or prefixed with the domain, and an exit code that agrees with them. Anything else is a protocol failure, never a success
- Ctrl-C and SIGTERM stop the extension's whole process tree. This brings the program runner, which doctor's probes, the SDK, and every extension reuse later
- `ncly --help` lists extensions in their own section, and the help and completion of an extension come from its manifest

The manifest so far, which the open questions complete:

```toml
format_version = 1
contract_version = 1
domain = "hello"
description = "Says hello"

[requires]
bins = []
keys = []
```

## Open questions

Settle each one in [Extensions](../north-star/contract.md#extensions), then delete this section.

- The complete manifest schema: command declarations, effects, modes, results, requirements, and the `version` of the extension. It must describe the bundled `skill` of M03 and the commands in [agent's spec](../../extensions/agent/spec.md) too, without duplicating help or completion definitions
- Whether the manifest carries the translations of its help text, or core asks the extension for its help in `NCLY_LANG`
- The error codes and exit codes for a missing manifest, an incompatible manifest, and a protocol failure
- Whether core captures the extension's stdout in JSON mode to check it, or passes the streams through and checks a copy
- The variable that gives an extension the path of the running `ncly`, so that it calls another domain through the same host
- How completion reaches values that only the extension knows, such as profile names
- The test commands that the cancellation scenarios need, such as `waitfile <path>`, which waits until a file exists, and `gone <path>`, which checks that the process whose ID the file holds no longer runs

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M01-T1 | Manifest and dispatch | agent | — | todo |
| M01-T2 | Answer check | agent | M01-T1 | todo |
| M01-T3 | Cancellation | agent | M01-T1 | todo |
| M01-T4 | Help and completion | agent | M01-T1 | todo |

### M01-T1 Manifest and dispatch

- **Read:** [Extensions](../north-star/contract.md#extensions), [Compatibility](../north-star/contract.md#compatibility), [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs)
- **Proves:** `testdata/script/extension_dispatch.txtar`, `testdata/script/extension_manifest.txtar`

`ncly hello` finds the fixture `ncly-hello` and reads its `ncly-hello.toml`, then runs it with its arguments and the `NCLY_*` variables and returns its exit code. The fixture and its manifest live in the scenario archive. A missing, malformed, or incompatible manifest is reported without starting the executable, which a fixture that writes a file when it runs proves. The data folder wins over `PATH`, an earlier `PATH` folder wins over a later one, and a reserved name never reaches an extension. Every `*_API_KEY` variable is removed from the extension's environment.

### M01-T2 Answer check

- **Read:** [Extensions](../north-star/contract.md#extensions), [Output](../north-star/contract.md#output), [Error codes](../north-star/contract.md#error-codes)
- **Proves:** `testdata/script/extension_protocol.txtar`

With `--json`, core checks the extension's answer. Fixtures answer with malformed JSON, two lines, another `contract_version`, an unknown code without the domain prefix, and an exit code that disagrees with `ok`. Each one fails as a protocol failure, and none counts as a success. Without `--json`, human output passes through unchanged.

### M01-T3 Cancellation

- **Read:** [Programs that ncly runs](../north-star/contract.md#programs-that-ncly-runs)
- **Proves:** `testdata/script/extension_cancel.txtar`, `testdata/script/extension_orphan.txtar`

The program runner's first version: Ctrl-C and SIGTERM reach the extension and every descendant that core can still reach, and SIGKILL follows 15 seconds later. Add the test commands that the open questions settle. In the first scenario, the fixture starts a grandchild in its own process group and writes its process ID. After `kill -INT`, the answer is `INTERRUPTED` with exit 130, and the grandchild is gone. The second scenario, with the `[linux]` condition, proves the subreaper: the fixture's child starts a grandchild in its own session and exits before the signal, and core still stops the orphan.

### M01-T4 Help and completion

- **Read:** [Extensions](../north-star/contract.md#extensions), [Usage](../north-star/contract.md#usage), [ncly completion](../north-star/core-spec.md#ncly-completion)
- **Proves:** `testdata/script/extension_help.txtar`

`ncly --help` shows extensions in their own section, in the existing help style, with the description from their manifest. `ncly hello --help` and `ncly help hello` show the extension's help without starting it, and completion offers its commands and flags. Run the scenario once under the pseudo-locale, as the open questions settle for extension text.
