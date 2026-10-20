# Architecture

## Relations

- An extension is bundled or external; a first-party extension can be either
- A default extension is external and ships with Nicely
- The SDK implements the contract and the services in Go, for core and every Go extension
- A Go extension calls the SDK, and the SDK calls the extension's handlers
- The program runner starts the programs that core or an extension runs, except the one that [Programs that ncly runs](../contract/programs-that-ncly-runs.md) names
- A Python program lives inside an extension and talks to its Go code in JSON only

## Definitions

- Bundled extension
  - Def.: an extension compiled into `ncly`, with the same manifest and contract as an external one
  - Project note: only `skill` is bundled
- External extension
  - Def.: an executable named `ncly-<domain>` with a compatible manifest and its catalogs beside it
- Default extension
  - Def.: an external extension that ships with Nicely as a separate executable, through the same path as any external extension
  - Project note: [D004](../decisions/D004-minimal-core.md) permits everyday domains in the default install
- First-party extension
  - Def.: an extension in `extensions/` of this repository, such as `agent` or `transcript`
- Platform
  - Def.: layer 0: the helpers that hide the operating system from the layers above
  - Project note: the [code map](../architecture.md#code-map) lists them
- Services
  - Def.: layer 2: what core and extensions share while work runs
  - Project note: the [code map](../architecture.md#code-map) lists them
  - Distinct from: a network service such as Deepgram, whose key `ncly auth` manages
- SDK
  - Def.: the public Go packages in `sdk/` that implement the contract and the shared services, for core and for every Go extension
  - Distinct from: services, the one layer of the SDK that runs work
- Handler
  - Def.: the Go function that runs one command of an extension
  - Project note: the author writes it. `extension.Main` calls it with an `extension.Call`, its context object, and a Go `ctx`, which only carries cancellation, as [SDK](../contract/sdk.md) defines. The SDK calls the author's code, not the reverse, which is inversion of control
- Program runner
  - Def.: the service that runs programs for core and extensions, with their environment, timeout, signals, and cleanup
  - Project note: [Programs that ncly runs](../contract/programs-that-ncly-runs.md) defines its scope, including the one program that starts outside it
- Python program
  - Def.: Python code that an extension embeds while it moves to Go, such as transcript's
- Engine
  - Def.: transcript's Python program, which it copied from `pascalandy/skills`
  - Project note: [the transcript spec](../../../extensions/transcript/spec/protocol.md#the-engine) uses this name
