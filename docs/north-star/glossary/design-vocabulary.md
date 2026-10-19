# Design vocabulary

## Relations

- Primitives compose into commands and extensions
- Layers build abstractions from the bottom up: foundation → runtime layer → application layer
- An interface hides its implementation
- An additive extension adds to its host; a behavioral extension changes it

## Definitions

- Abstraction
  - Def.: something simpler that hides complexity behind it
  - Project note: each layer is an abstraction over the layers below it
- Layer
  - Def.: a level of abstraction that depends only on the levels below it
  - Project note: platform, contract, services, core, and extensions, drawn in [architecture.md](../architecture.md#layers)
  - Distinct from: a layer of a [stack](../glossary.md#planning), which `just check-layers` checks
- Foundation
  - Def.: the lowest layer, which hides the outside world behind one interface
  - Project note: the platform
  - Distinct from: Pi's `pi-ai`, its foundation over LLM providers
- Runtime layer
  - Def.: the machinery that executes work, apart from what the work is
  - Project note: the services
  - Distinct from: a runtime such as Node, which runs code
- Application layer
  - Def.: the layer that users and agents talk to
  - Project note: core
- Host
  - Def.: a program that finds, loads, and runs extensions
  - Project note: core
- Primitive
  - Def.: a piece that a layer exposes as fundamental, which nobody decomposes at that level
  - Project note: [Primitives](../contract.md#primitives) lists Nicely's
  - Distinct from: a construct, a piece of syntax such as `if`, and a built-in, which ships with a system whether or not it is primitive
- Interface
  - Def.: what a part promises, apart from how it works
  - Project note: the wire contract, for extensions, and the agent contract, for every caller of the CLI
- Implementation
  - Def.: how a part keeps the promise of its interface
  - Project note: the SDK implements the contract in Go
- Composition
  - Def.: combining small pieces into a bigger one. A piece is composable when it combines without special cases
  - Project note: extensions compose through the CLI, as transcript calls `ncly agent run --json`
- Modular
  - Def.: split into parts with clear boundaries, so each part changes alone
  - Project note: each part has one owner, one package, and one spec section, in the [code map](../architecture.md#code-map)
- Distribution unit
  - Def.: what people install and share
  - Project note: a tap
- Additive extension
  - Def.: an extension that adds capabilities without changing how its host behaves
  - Project note: every Nicely extension, as principle 13 says
- Behavioral extension
  - Def.: an extension that changes how its host behaves through hooks, such as blocking a call
  - Project note: Pi's extensions. Nicely has none
- Dogfooding
  - Def.: building your own features with the public interface that outsiders get
  - Project note: first-party domains are extensions, and core uses the SDK
