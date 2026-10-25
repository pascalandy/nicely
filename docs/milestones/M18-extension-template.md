# M18 Extension template

Status: done
Version: v0.18.0

## Demo

An agent wraps a script from a fresh template repository, runs its tests there, and installs it from a tag with `just install`.

## Scope

Pascal settled the requirements in [#273](https://github.com/pascalandy/nicely/issues/273) on 2026-10-20.
[Extensions](../north-star/contract/extensions.md) holds the rules of an extension repository and the bar, "What a great extension does".
Core gains no command: an extension installs from its own repository, and its repository tests itself.
Taps, other languages for an extension, and a cache for workers stay in [backlog](../backlog/README.md).
A small personal script becomes the first extension built from the released template, before the author guide.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M18-T1 | Settle the extension requirements | Pascal | — | done |
| M18-T2 | Make the extension template ready: the external author contract, the bar, and cards T3 to T5 | agent | M18-T1 | done |
| M18-T3 | Publish the SDK test kit | agent | M18-T2 | done |
| M18-T4 | Wrap programs and embed workers through the SDK | agent | M18-T2 | done |
| M18-T5 | Release the Copier template and the skill entry | agent | M18-T3, M18-T4 | done |

### M18-T1 Settle the extension requirements

Pascal settled the requirements in [#273](https://github.com/pascalandy/nicely/issues/273) on 2026-10-20.

### M18-T2 Make the extension template ready: the external author contract, the bar, and cards T3 to T5

- **Read:** [Scope](../backlog/README.md), [#273](https://github.com/pascalandy/nicely/issues/273), [Make a item ready](../backlog/README.md#make-an-item-ready), [Architecture](../north-star/architecture.md)
- **Proves:** `docs/north-star/contract/extensions.md` holds the bar and the external author rules. `docs/north-star/contract/programs-that-ncly-runs.md` defines a worker and the new session. `docs/north-star/contract/sdk.md` lists the rules that its types cannot express. Principle 8 and D013 are scoped, and D047 is revised. `docs/north-star/decisions/D062-copier-template.md` and `docs/north-star/decisions/D063-sdk-test-kit.md` exist

Write the specs that the cards below read, and move the milestone and contributor instructions out of the contract parts.

### M18-T3 Publish the SDK test kit

- **Read:** [Test kit](../north-star/contract/test-kit.md), [Extension repository](../north-star/contract/extensions.md#extension-repository), [SDK](../north-star/contract/sdk.md), [Compatibility](../north-star/contract/compatibility.md), [D063](../north-star/decisions/D063-sdk-test-kit.md)
- **Proves:** `sdk/extensiontest/`, through `testdata/extensions/standalone/`, a fixture module with its own go.mod whose go test runs the kit against go tool ncly. It covers the link checks, the version check, and key cases that never read the keychain, and it settles Lip Gloss

Add `sdk/extensiontest`, and `Export` in `sdk/extension`, which writes the sidecars of an `extension.Extension` value. The scratch install of the kit uses `Export`, so the sidecars come from the build under test.

Lay out the fixture module as an extension repository. Its `go.mod` replaces the Nicely module with this checkout, and just test runs it against the local SDK. Prove a granted key and a missing key with sentinel values, and that no case reads the keychain. Give each check of the kit a case that fails and names its fix:

- A command without a scenario
- A skill that names an undeclared command
- An answer that breaks its schema
- A Nicely version in `.copier-answers.yml` that differs from `go.mod`

Nicely's `go.mod` replaces `charm.land/lipgloss/v2` with a fork that fixes a `tmux info` call at package init. A `replace` never reaches a consumer module, and both the extension and the `go tool ncly` host import Lip Gloss. Use the upstream release once it carries the fix. Otherwise, put the same `replace` in the fixture module, and prove with a test that its build gets the fix.

Once this card ships, remove from the docs each marker that names M18-T3, such as "from M18-T3".

### M18-T4 Wrap programs and embed workers through the SDK

- **Read:** [Wrapped programs](../north-star/contract/wrap.md), [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md), [Commands that core calls](../north-star/contract/commands-that-core-calls.md), [Required programs](../north-star/core-spec.md#required-programs), [SDK](../north-star/contract/sdk.md), [D047](../north-star/decisions/D047-terminal-process-group.md)
- **Proves:** `sdk/wrap/`, `testdata/extensions/hello/testdata/script/wrap.txtar`, and `testdata/script/session.txtar`. They cover a missing program, a version probe that times out, an exit other than 0, malformed output, and a timeout with exit 1. They also cover Ctrl-C with exit 130, an empty stdin, and a child environment with only the declared keys. A worker reads ./input.md from the caller's folder, and an open of /dev/tty fails at once under a real PTY, on Linux and macOS

Add `sdk/wrap`, as [Wrapped programs](../north-star/contract/wrap.md) defines, and state its default size limit there. Add `Resources` to `extension.Extension`, with the extraction that [SDK](../north-star/contract/sdk.md) defines. Give the `hello` fixture one wrapped command with a worker, which the scenarios run.

In `sdk/runner`, capture output up to a size limit. Start each program without the terminal in a new session, instead of its own process group. Setting both would fail, since `setpgid` refuses a session leader. Doctor's version probes and the commands that core calls then get the new session through the runner. Prove the open of `/dev/tty` under the PTY that core's `ttyin` test helper gives, for a worker, a version probe, and a status command.

Once this card ships, remove from the docs each marker that names M18-T4, such as "from M18-T4".

### M18-T5 Release the Copier template and the skill entry

- **Read:** [Extension repository](../north-star/contract/extensions.md#extension-repository), [What a great extension does](../north-star/contract/extensions.md#what-a-great-extension-does), [Test kit](../north-star/contract/test-kit.md), [Wrapped programs](../north-star/contract/wrap.md), [D062](../north-star/decisions/D062-copier-template.md), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `copier.yml`, `template/`, and `scripts/template-test.sh`. Nicely's just check renders the template, and the rendered repository passes go mod tidy, just check, and just install into scratch folders. just update carries a customized repository to the next template revision, and keeps the author's edits. On a conflict, it stops before go get

Add `copier.yml` at the root, whose `_subdirectory` points at `template/`. The template has no tasks and no migrations, so Copier needs no `--trust`. Give each Go file of the template the `.jinja` suffix, so the `./...` of Nicely's module never builds it. An agent creates a repository without prompts, with a pinned Copier: `copier copy --vcs-ref v<nicely-version> --defaults`, and `--data` for the domain and the module.

The template renders the layout of [Extension repository](../north-star/contract/extensions.md#extension-repository), with `vision.md` and `spec.md` written as forms to fill. It adds `extension_test.go`, which calls the test kit, `lefthook.yml`, which runs just check before each push, and `.copier-answers.yml`. Its `go.mod` holds `tool github.com/pascalandy/nicely/cmd/ncly`, and it ships no GitHub workflow. Its example command, `inspect`, wraps a short standard-library Python worker that returns the size and the line count of a file. Seed its scenarios with the cases of [Wrapped programs](../north-star/contract/wrap.md).

just check of the rendered repository adds go mod tidy -diff, formatting, and lint to go test. just update needs a clean Git tree. It runs copier update, then go get of that Nicely version and go mod tidy, then just check. Copier writes a conflict as markers and can still exit 0. The recipe therefore looks for markers, `.rej` files, and unmerged index entries before it touches `go.mod`. On a conflict, it stops and prints the three commands that finish the update. When Copier fails or stops, the files can be incomplete without a marker. The recipe then tells the author to return to the commit before the update, remove the files that it added, and run it again. It never resets, stashes, or commits.

The template's authoring skill, in `.agents/skills/nicely-extension/`, covers the intake, the choice of operations, the readiness review, the error mapping, and the verification. The readiness review runs only when the user asks for it. Each finding names its evidence, the command that it affects, its consequence, its severity, who can fix it, and the smallest fix. The review ends with one verdict: ready, ready with stated limits, or needs work. Accepted limits go into `spec.md`. Declining the review never waives a check of the contract. The review never rejects a script for being small, young, or private.

The bundled `nicely` skill gains an entry that gives the template command and links [Extensions](../north-star/contract/extensions.md). Carry the Lip Gloss choice of M18-T3 into the template. Once this card ships, remove from the docs each marker that names M18-T5, such as "from M18-T5".

## After this milestone

The release check waits for the tag of M18 and blocks no card. On the mbp and on om1, install a tagged extension with git clone and just install. Then run just contract and go test from an empty module cache, because a local `replace` would hide a file missing from the module zip. M18 counts as released once the check passes. A failure gets a patch release before the first extension built from the template.
