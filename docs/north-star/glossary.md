# Nicely glossary

This file fixes the vocabulary of Nicely: what each term means here, how the terms relate, and which words to avoid.

## Reading conventions

- Relations come before definitions. An arrow reads as a path, from the first term to the last
- In a definition, Def. gives the local meaning, Project note the Nicely instance or nuance, and Distinct from a term often confused with it
- [Terms to avoid](#terms-to-avoid-as-canonical-forms) names the words to replace

## Global conceptual model

### Relations

- Core hosts extensions; an extension adds a domain and never changes core
- An extension reaches core only through the wire contract: its manifest, its environment, its answer, its exit code, and its run records
- The agent contract and the wire contract are two overlapping scopes of one contract: both hold the answer and the exit code
- An extension reaches another extension only through the CLI, as an agent does
- Platform → contract → services → core → extensions: each layer uses only the layers below it
- A tap carries skills and extensions to a machine
- A principle links the decisions behind it; a decision explains a rule that a spec holds

### Definitions

- Core
  - Def.: the host that every install gets. It dispatches each command and owns the commands of [core-spec.md](core-spec.md)
  - Project note: layer 3. It knows no domain
  - Distinct from: Pi's `pi-agent-core`, which is a runtime layer
- Extension
  - Def.: a domain added to core through a manifest, bundled or external
  - Distinct from: a Pi extension, which runs inside Pi's process and hooks its lifecycle
- Contract
  - Def.: the rules that every command and every extension follows, in [contract.md](contract.md)
  - Project note: layer 1. It defines the primitives

## Architecture

### Relations

- An extension is bundled or external; a first-party extension can be either
- The SDK implements the contract and the services in Go, for core and every Go extension
- The program runner starts every program that core or an extension runs
- A Python program lives inside an extension and talks to its Go code in JSON only

### Definitions

- Bundled extension
  - Def.: an extension compiled into `ncly`, with the same manifest and contract as an external one
  - Project note: only `skill` is bundled
- External extension
  - Def.: an executable named `ncly-<domain>` with a compatible manifest beside it
- First-party extension
  - Def.: an extension in `extensions/` of this repository, such as `agent` or `transcript`
- Platform
  - Def.: layer 0: the helpers that hide the operating system from the layers above
  - Project note: the [code map](architecture.md#code-map) lists them
- Services
  - Def.: layer 2: what core and extensions share while work runs
  - Project note: the [code map](architecture.md#code-map) lists them
- SDK
  - Def.: the public Go packages in `sdk/` that implement the contract and the shared services, for core and for every Go extension
  - Distinct from: services, the one layer of the SDK that runs work
- Program runner
  - Def.: the service that runs every other program, with its environment, timeout, signals, and cleanup
- Python program
  - Def.: Python code that an extension embeds while it moves to Go, such as transcript's

## Contract

### Relations

- A command reads domain → resource → verb
- A manifest holds one command declaration per command
- An answer and its exit code always agree; exit 75 alone permits an automatic retry
- A grant lets one key reach one extension
- A report command keeps its report on stdout, even when a check fails

### Definitions

- Agent contract
  - Def.: the scope of the contract that a caller of the CLI relies on: command names, flags, answers, error codes, and exit codes
  - Project note: [contract.md](contract.md) and the extension specs define it. Translation never changes it
  - Distinct from: the wire contract, which it overlaps on the answer and the exit code
- Wire contract
  - Def.: the scope of the contract that core and an extension exchange: the manifest, the environment, the answer, the exit code, and the run records, as [D040](decisions/D040-wire-contract.md) decides
  - Distinct from: an RPC protocol through which core serves extensions, which D040 rejects
- Domain
  - Def.: a top-level noun that groups actions, such as `video`
- Resource
  - Def.: an optional second noun inside a domain, such as `prompt` in `ncly transcript prompt list`
- Verb
  - Def.: the action of a command, one per action across domains, such as `list` or `run`
  - Project note: [Usage](contract.md#usage) lists them
- Command declaration
  - Def.: the one description of a command that supplies its help, completion, validation, and discovery
- Manifest
  - Def.: the file that declares an extension's commands, effects, modes, and requirements, read without starting the extension
- Environment
  - Def.: the variables that an extension receives: the resolved global flags, the versions, and the keys granted to it
- Answer
  - Def.: the one JSON object that a command prints under `--json`, with `ok` and, on failure, `errors`
- Exit code
  - Def.: the number a command exits with. Each error code maps to one
- Report command
  - Def.: a command whose result is a report, such as `ncly doctor`
- Grant
  - Def.: the permission, given by a human in a terminal, for one extension to receive one key
- Interactive mode
  - Def.: the mode in which `ncly` may ask questions. [Modes](contract.md#modes) defines when it applies
  - Distinct from: non-interactive mode
- Non-interactive mode
  - Def.: every case outside interactive mode. A missing value fails instead of opening a form

## Operations

### Relations

- A command performs an operation; an operation that bills or leaves reusable work writes a run record
- A run record references its artifacts and the evidence of each step
- Resume continues a run record after checking its evidence, and never repeats an unknown effect
- Dry run and execution share one preparation

### Definitions

- Operation
  - Def.: the work prepared and performed by a command. The domain owns its steps
- Dry run
  - Def.: a simulation that uses the same preparation as execution and changes nothing outside Nicely's cache
- Run record
  - Def.: the durable evidence of one operation, identified by `run_id`, including its step outcomes and artifact references
- Artifact
  - Def.: a result produced by an operation, referenced by its path and the evidence needed to verify it
- Resume
  - Def.: an explicit request to continue a recorded run after validating its inputs, artifacts, and remaining steps

## Agents

### Relations

- An agent operates `ncly` through a shell, and `ncly` runs an agent through the `agent` extension
- A profile names a harness, a provider, a model, and an effort
- A skill teaches an agent a task; each extension ships one

### Definitions

- Agent
  - Def.: an AI program that works through a shell. The `agent` extension launches one
- Harness
  - Def.: the program an agent runs in: Claude Code, Codex, Pi, OpenCode, or Grok
- Profile
  - Def.: a named combination of harness, provider, model, and effort, stored in the config
- Skill
  - Def.: a folder with a `SKILL.md` that teaches an agent a task

## Configuration and text

### Relations

- The local config overrides the shared config on one machine
- The pseudo-locale marks every catalog string, so text outside a catalog stands out

### Definitions

- Shared config
  - Def.: `config.toml`, the setup the user wants, which can travel between machines
- Local config
  - Def.: `config.local.toml`, which holds what belongs to one machine and overrides the shared config
- Pseudo-locale
  - Def.: a generated test language that marks every catalog string

## Distribution

### Relations

- A tap carries skills and extensions; this repository is the official tap of the first-party extensions
- The Homebrew tap ships `ncly` itself

### Definitions

- Tap
  - Def.: a Git repository that holds skills and extensions, like a Pi package
  - Distinct from: the Homebrew tap
- Homebrew tap
  - Def.: the repository `pascalandy/homebrew-tap`, whose formula builds `ncly`

## Planning

### Relations

- A milestone holds cards, at most five of them for agents; a card is one pull request
- A principle links its decisions; a decision records what was decided, why, and what was rejected

### Definitions

- Milestone
  - Def.: a numbered stage of work in `docs/milestones/`, one capability that a sentence can demonstrate
  - Project note: its status is `planned`, `ready`, `active`, or `done`. M99 is the parking lot, and its status stays `open`
- Card
  - Def.: one pull request of a milestone, with its owner, its dependencies, what to read, and what proves it
- Principle
  - Def.: a rule that shapes every design choice, in [principles.md](principles.md)
- Decision
  - Def.: one file in [decisions/](decisions/README.md) that records what was decided, why, and what was rejected

## Design vocabulary

### Relations

- Primitives compose into commands and extensions
- Layers stack abstractions: foundation → runtime layer → application layer
- An interface hides its implementation
- An additive extension adds to its host; a behavioral extension changes it

### Definitions

- Abstraction
  - Def.: something simpler that hides complexity behind it
  - Project note: each layer is an abstraction over the layers below it
- Layer
  - Def.: a level of abstraction that uses only the levels below it
  - Project note: platform, contract, services, core, and extensions, drawn in [architecture.md](architecture.md#layers)
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
  - Def.: the smallest piece that a layer exposes, which nobody decomposes at that level
  - Project note: [Primitives](contract.md#primitives) lists Nicely's
  - Distinct from: a construct, a piece of syntax such as `if`, and a built-in, which ships with a system without being primitive
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
  - Project note: each part has one owner, one package, and one spec section, in the [code map](architecture.md#code-map)
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

## Terms to avoid as canonical forms

- plugin, add-on, module, for an extension
  - Use instead: extension
  - Reason: one word names a domain added to core, bundled or external
- headless mode
  - Use instead: non-interactive mode
  - Reason: [Modes](contract.md#modes) names the two modes
- guide
  - Use instead: the file that answers the question: vision.md, principles.md, architecture.md, or glossary.md
  - Reason: the guide was split by question
- cli-spec
  - Use instead: contract.md for the rules of every command, core-spec.md for core's commands
- the spec, alone
  - Use instead: contract.md, core-spec.md, or the extension's spec.md
  - Reason: three kinds of spec exist
- tap, for the Homebrew tap
  - Use instead: Homebrew tap
- task, for one pull request of a milestone
  - Use instead: card

## Terminology decisions kept

- Harness stays the internal term for the program an agent runs in, as [D018](decisions/D018-agent-extension.md) decides
- Core means the host at layer 3, never everything that Nicely provides
- A tap is a repository of skills and extensions; the Homebrew tap keeps its full name
- Card, not task, names one pull request of a milestone, as [D042](decisions/D042-milestones-and-cards.md) decides
- Contract names all of contract.md. Agent contract and wire contract name its two overlapping scopes, never separate protocols
