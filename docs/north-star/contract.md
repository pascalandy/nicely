# Contract

This file and its parts are the contract that every command follows, core's and each extension's. The contract covers what humans and agents type, what core and an extension exchange, and what scripts parse. Core's commands live in [core-spec.md](core-spec.md), and each extension's in its own spec, such as [transcript's](../../extensions/transcript/spec.md). The code tables name the milestone that brings each entry.

## Primitives

Every command and every extension composes these primitives. Each belongs to one or both scopes of the contract: the agent contract, what a caller of the CLI types and parses, and the wire contract, what core and an extension exchange, as [D040](decisions/D040-wire-contract.md) decides. Each primitive is defined once, in the section that the table links.

| Primitive | Scope | Defined in |
|---|---|---|
| Domain, resource, and verb | Agent contract | [Usage](contract/usage.md) |
| Flag | Agent contract | [Global flags](contract/global-flags.md) |
| Manifest | Wire contract | [Manifest](contract/manifest.md) and [Command descriptions](contract/command-descriptions.md) |
| Environment | Wire contract | [Programs that ncly runs](contract/programs-that-ncly-runs.md) |
| Answer | Both | [Output](contract/output.md) |
| Exit code | Both | [Exit codes](contract/exit-codes.md) |
| Run record | Wire contract | [Record format](contract/record-format.md) |

## Parts

Each part is one file, in reading order.

- [Usage](contract/usage.md)
- [Global flags](contract/global-flags.md)
- [Modes](contract/modes.md)
- [Output](contract/output.md)
- [Compatibility](contract/compatibility.md)
- [Retry safety](contract/retry-safety.md)
- [Exit codes](contract/exit-codes.md)
- [Error codes](contract/error-codes.md)
- [Configuration](contract/configuration.md)
- [Command descriptions](contract/command-descriptions.md)
- [Extensions](contract/extensions.md)
  - [Discovery](contract/discovery.md)
  - [Manifest](contract/manifest.md)
  - [Valid manifest](contract/valid-manifest.md)
  - [Text of an extension](contract/text-of-an-extension.md)
  - [Dispatch](contract/dispatch.md)
  - [Commands that core calls](contract/commands-that-core-calls.md)
  - [Completion values](contract/completion-values.md)
  - [Status command](contract/status-command.md)
  - [Answer check](contract/answer-check.md)
  - [SDK](contract/sdk.md)
    - [Wrapped programs](contract/wrap.md)
    - [Test kit](contract/test-kit.md)
- [Operations](contract/operations.md)
  - [Records and evidence](contract/records-and-evidence.md)
  - [Record format](contract/record-format.md)
  - [Inspection and explicit resume](contract/inspection-and-explicit-resume.md)
  - [Run IDs](contract/run-ids.md)
  - [Dry-run answer](contract/dry-run-answer.md)
- [Keys](contract/keys.md)
- [Grants](contract/grants.md)
- [Programs that ncly runs](contract/programs-that-ncly-runs.md)
