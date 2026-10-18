# M01 Extension host

Status: done
Version: v0.1.0

## Demo

`ncly hello run --json` runs a test extension written in shell from its manifest, and `ncly` checks its answer.

## Scope

Core finds executables named `ncly-<domain>`, reads their manifests and catalogs without running them, starts them with a validated call, and checks what they answer, as [Extensions](../north-star/contract/extensions.md) defines. This milestone uses fixture extensions only. The first bundled extension, `skill`, arrives in [M04](M04-skill.md), and the first real external one, `agent`, in [M09](M09-agent-profiles.md).

- The fixtures are shell scripts that live in the scenario archives, beside their manifests and their `en` catalogs. A fixture that writes a file when it runs proves that core did not start it
- Core validates an extension's call from its manifest, as it does for its own commands, then passes the canonical arguments of [Dispatch](../north-star/contract/dispatch.md)
- The explicit `reservedNames` of `internal/cli/root.go` stays, because it also reserves names whose command arrives later, and a test keeps it equal to the core and bundled rows of the [domain table](../north-star/architecture.md#domains-and-commands), plus `help`, `version`, and `config`
- An extension receives the environment that [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md) defines. No key reaches it until grants arrive in [M08](M08-grants.md)
- `--json` selects non-interactive mode, as [D046](../north-star/decisions/D046-machine-output-never-asks.md) decides, so core holds an extension's stdout without changing its mode
- The program runner arrives in `internal/runner/` and the process tree helpers in `internal/platform/`, which [M02](M02-sdk.md) moves to `sdk/`. An extension stays in the process group of `ncly`, as [D047](../north-star/decisions/D047-terminal-process-group.md) decides
- The cancellation scenarios run on macOS and Linux, so their fixtures use no `setsid`, which macOS lacks, and no `set -m`, which dash 0.5.12, the `sh` of Ubuntu 24.04, refuses without a terminal. Only the subreaper scenario is Linux only, and the test binary shortens the 15 seconds of cleanup

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M01-T1 | Manifest, help, and completion | agent | — | done |
| M01-T2 | Dispatch | agent | M01-T1 | done |
| M01-T3 | Answer check | agent | M01-T2 | done |
| M01-T4 | Cancellation | agent | M01-T3 | done |
| M01-T5 | Stubborn programs | agent | M01-T4 | done |

### M01-T1 Manifest, help, and completion

- **Read:** [Discovery](../north-star/contract/discovery.md), [Manifest](../north-star/contract/manifest.md), [Text of an extension](../north-star/contract/text-of-an-extension.md), [Command descriptions](../north-star/contract/command-descriptions.md), [Usage](../north-star/contract/usage.md), [ncly completion](../north-star/core-spec.md#ncly-completion)
- **Proves:** `testdata/script/extension_manifest.txtar`, `testdata/script/extension_help.txtar`

Everything core does with an extension without starting it. Core finds the fixture `ncly-hello` in the data folder and in `PATH`, reads `ncly-hello.toml` and `ncly-hello.locales/en.toml`, and builds its commands with the same code as its own. The manifest's types and its validation live in `internal/contract`, because [M02](M02-sdk.md) shares them with every Go extension.

The first scenario covers discovery and validation. The data folder wins over `PATH`, an earlier `PATH` folder wins over a later one, a relative `PATH` entry is skipped, and a fixture named after a reserved name is never listed. Each broken manifest answers `EXTENSION_INVALID` with exit 78 for its domain, and its fixture never runs: a missing manifest, a TOML syntax error, `format_version = 2`, `contract_version = 2`, a domain that differs from the file name, an undeclared parent path, a flag named `json`, a shorthand `é`, a flag named `a=b`, an own code without the domain prefix, an own code that maps to 75, and a summary ID that the `en` catalog lacks. Another fixture whose manifest is valid still shows in `ncly --help` beside the broken ones, so no broken declaration reaches the parser.

The second scenario covers the text. `ncly --help` lists extensions in their own section, in the existing help style, with the summary of each domain. `ncly hello --help`, `ncly help hello`, and `ncly hello` alone show the extension's help, and completion offers its commands and flags. The scenario runs once under the pseudo-locale, so the fixture's text is marked. A second fixture ships a partial `fr-CA` catalog: with `--lang fr-CA`, its help and completion show its French summary and the English text of the entry that the catalog lacks, never the ID, while the first fixture falls back to its `en` catalog. None of these calls starts a fixture.

### M01-T2 Dispatch

- **Read:** [Dispatch](../north-star/contract/dispatch.md), [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md), [Global flags](../north-star/contract/global-flags.md), [Configuration](../north-star/contract/configuration.md)
- **Proves:** `testdata/script/extension_dispatch.txtar`, `internal/cli/spec_test.go`

`ncly hello run` starts the fixture through the first version of the program runner, which builds its environment and returns its exit code. The fixture writes its arguments and its `NCLY_*` variables to a file, so the scenario checks exactly what it received:

- `ncly -v hello run -g Hi Ada --json` passes `run --greeting=Hi -- Ada`, with `NCLY_VERBOSE=1` and `NCLY_JSON=1`, and `--json=false` with `NCLY_JSON=1` in the environment removes `NCLY_JSON`
- An operand after `--` that starts with a dash reaches the fixture after its own `--`, and a repeated flag passes once with its last value
- An unknown flag, a flag without its value, and an extra argument fail with `USAGE_INVALID` before the fixture runs
- `NCLY_LANG` holds `fr-CA` for `LANG=fr_CA.UTF-8`, and `en` for `LANG=C`. `NCLY_BIN` holds the path of the running `ncly`, `NCLY_VERSION` and `NCLY_CONTRACT_VERSION` are set, and every `*_API_KEY` variable is gone
- A config file with a syntax error does not stop the fixture
- Without `--json`, the fixture's text reaches the caller unchanged with its exit code, and an exit code outside [Exit codes](../north-star/contract/exit-codes.md) becomes `RUNTIME` with exit 1
- Under the pseudo-terminal of `ttyin -stdin`, the fixture compares `ps -o pgid=` with `ps -o tpgid=` for its own process, finds that it belongs to the terminal's foreground group, and reads a line from the terminal, as [D047](../north-star/decisions/D047-terminal-process-group.md) requires

`internal/cli/spec_test.go` keeps `reservedNames` equal to the domain table.

### M01-T3 Answer check

- **Read:** [Answer check](../north-star/contract/answer-check.md), [Output](../north-star/contract/output.md), [Error codes](../north-star/contract/error-codes.md), [Modes](../north-star/contract/modes.md)
- **Proves:** `testdata/script/extension_protocol.txtar`, `sdk/contract/compat_test.go`

With `--json`, core holds the fixture's stdout and checks the answer against the rules of the spec. Fixtures answer with malformed JSON, two lines on stdout, an answer on stdout with exit 1, `contract_version` 2, a problem without `hint`, a lowercase code, an own code that the manifest does not declare, `USAGE_INVALID` with exit 1, a declared own code with the wrong exit, `USAGE_INVALID` leading `RUNTIME` with exit 2, the same with a code that `ncly` does not know added, exit 0 with a failed item, a failed item with exit 78 led by `AUTH_MISSING`, and exit 75 with a `RUNTIME` item error. Each one answers `RUNTIME` with exit 1, stdout empty, and the fixture's held output absent. Valid answers pass unchanged: a success, a declared own code, a code that this `ncly` does not know with exit 1, a warning with the fixture's prefix, a report answering `ok: false` on stdout with exit 78, the same report failing itself with `CONFIG_INVALID` on stderr and exit 78, and a command of kind `text`, whose output passes through under `--json`. Diagnostic lines that the fixture writes to stderr reach the caller in order, before its failure answer, and after its success.

A direct test of `contract.DetectMode` proves that `--json` selects non-interactive mode, and that `--json=false` with `NCLY_JSON=1` does not, because no command asks a question yet. A consumer case in `compat_test.go` reads an answer whose code it does not know and acts on its exit code, as the M00 [contract coverage](M00-foundation.md#contract-coverage) left for this milestone.

### M01-T4 Cancellation

- **Read:** [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md), [Answer check](../north-star/contract/answer-check.md)
- **Proves:** `testdata/script/extension_cancel.txtar`, `testdata/script/extension_orphan.txtar`

The program runner forwards a signal: Ctrl-C and SIGTERM reach the extension and every descendant that core can still reach, and core answers with the verdict of the first signal. The escalation to SIGKILL waits for M01-T5. Add the test commands that the scenarios need: `waitfile <path>`, which waits until a file exists, `gone <path>`, which checks that the process whose ID the file holds no longer runs, and `term`, which sends SIGTERM to the background commands, because the built-in `kill` of testscript sends only INT and KILL.

In the first scenario, the fixture writes a finished file and starts a grandchild in its own process group through `pgroup`, a program of the test binary. The grandchild writes its process ID once it is ready, and the name of the signal that ends it. After `kill -INT`, the answer is `INTERRUPTED` with exit 130, and after `term`, `TERMINATED` with exit 143. Either way, the grandchild is gone and the finished file is still there, which closes [#5](https://github.com/pascalandy/nicely/issues/5). The name it writes is `INT` after `kill -INT` and `TERM` after `term`, which proves that the signal reached the grandchild before any SIGKILL. When the fixture traps the signal and answers with an `output_dir` key, the answer keeps that key after `INTERRUPTED`. When it exits without an answer, core answers alone. The second scenario, with the `[linux]` condition, proves the subreaper: the fixture's child starts a grandchild in its own session and exits before the signal, and core still stops the orphan.

### M01-T5 Stubborn programs

- **Read:** [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md)
- **Proves:** `testdata/script/extension_cancel_stubborn.txtar`, `testdata/script/extension_stream_holder.txtar`

The program runner ends what a signal does not: SIGKILL after the 15 seconds of cleanup, and the 2 seconds of the drain limit, on their own clocks, as [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md) defines. The test binary shortens cleanup to 4 seconds and keeps the real 2-second drain limit, so the scenarios tell the two clocks apart.

The first scenario has two fixtures. One traps the signal and answers with its `output_dir` key 3 seconds later, and the answer keeps that key, which fails when the drain limit cuts cleanup short. The other traps Ctrl-C and SIGTERM, writes the signal it receives to a file, and keeps running. After `kill -INT`, the file holds `INT`, which proves that the signal came before SIGKILL, and the fixture is gone once the shortened cleanup ends. A second `kill -INT` during cleanup changes nothing: core still stops the fixture and answers `INTERRUPTED`. Check the answer and that the fixture is gone, because a core that died on the second signal would leave both wrong.

In the second scenario, the fixture starts a grandchild that inherits stdout and outlives it. With `--json`, core passes the fixture's valid answer on 2 seconds after the fixture exits, while the grandchild still runs. During those 2 seconds, once `gone` shows that the fixture has exited, the scenario sends `kill -INT` to core, and the answer stays the fixture's success with exit 0, because a signal after the extension exits changes nothing. Then the scenario stops the grandchild.
