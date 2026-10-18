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
- Platform → contract → services → core → extensions: each layer depends only on the layers below it
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
- The program runner starts the programs that core or an extension runs, except the one that [Programs that ncly runs](contract.md#programs-that-ncly-runs) names
- A Python program lives inside an extension and talks to its Go code in JSON only

### Definitions

- Bundled extension
  - Def.: an extension compiled into `ncly`, with the same manifest and contract as an external one
  - Project note: only `skill` is bundled
- External extension
  - Def.: an executable named `ncly-<domain>` with a compatible manifest and its catalogs beside it
- First-party extension
  - Def.: an extension in `extensions/` of this repository, such as `agent` or `transcript`
- Platform
  - Def.: layer 0: the helpers that hide the operating system from the layers above
  - Project note: the [code map](architecture.md#code-map) lists them
- Services
  - Def.: layer 2: what core and extensions share while work runs
  - Project note: the [code map](architecture.md#code-map) lists them
  - Distinct from: a network service such as Deepgram, whose key `ncly auth` manages
- SDK
  - Def.: the public Go packages in `sdk/` that implement the contract and the shared services, for core and for every Go extension
  - Distinct from: services, the one layer of the SDK that runs work
- Program runner
  - Def.: the service that runs programs for core and extensions, with their environment, timeout, signals, and cleanup
  - Project note: [Programs that ncly runs](contract.md#programs-that-ncly-runs) defines its scope, including the one program that starts outside it
- Python program
  - Def.: Python code that an extension embeds while it moves to Go, such as transcript's

## Contract

### Relations

- A command reads domain → resource → verb
- A manifest holds one command declaration per command
- Core starts an extension with canonical arguments, and checks its answer; an answer that breaks the contract is a protocol failure
- An answer and its exit code always agree; exit 75 alone permits an automatic retry
- A problem names what went wrong with its code, and may name the next command as words
- A command declares the services whose keys it may need
- A grant lets one key reach one extension at one source
- A report command keeps its report on stdout, even when a check fails
- Doctor lists the checks of an extension's status command under the extension's component
- Completion takes the values that only an extension knows from a list command
- Completion and doctor each call a command of an extension themselves: a list command and a status command, which keep their own conditions

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
  - Project note: `ncly-<domain>.toml`, whose schema [Manifest](contract.md#manifest) defines
- Conditional program
  - Def.: a program in a manifest's `programs` that a command needs only for some configs or invocations, such as one harness among several
  - Project note: [Required programs](core-spec.md#required-programs) defines it. Doctor never checks it, and preparation checks it when an invocation selects it
- Canonical arguments
  - Def.: the one form in which core passes a validated call to an extension: the command path after the domain, each flag as `--<name>=<value>`, then `--` and the positional arguments
  - Project note: [Dispatch](contract.md#dispatch) defines it. `--dry-run` and `--run-id` pass as words after the own flags; a flag that has a variable, such as `--json`, reaches the extension as that variable instead
- Protocol failure
  - Def.: an answer or an exit code from an extension or an embedded program that breaks the contract
  - Project note: core answers `RUNTIME` in its place, as [Answer check](contract.md#answer-check) defines
  - Distinct from: a failure that the extension reports through a valid answer
- Environment
  - Def.: the variables that an extension receives: the resolved global flags, the versions, and the keys granted to it
- Answer
  - Def.: the one JSON object that a command prints under `--json`, with `ok` and, on failure, `errors`
- Exit code
  - Def.: the number a command exits with. Each error code maps to one
- Problem
  - Def.: one entry of `errors` or `warnings`: a `code`, a `message`, a `hint`, and an optional `next`
  - Project note: [Output](contract.md#output) defines it. Agents act on `code` and `next`, never on `message` or `hint`
- Next command
  - Def.: the command that a problem's hint names, as an array of words in `next`, present only when it runs as is
  - Distinct from: a hint, the text for a human, which may wrap the same command in a sentence
- Report command
  - Def.: a command whose result is a report, such as `ncly doctor`
- Status command
  - Def.: the report command `ncly <domain> status` that an extension's manifest marks for doctor, which runs it and lists its checks under the extension's component
  - Project note: [Status command](contract.md#status-command) defines it. An agent may also call it alone
  - Distinct from: `ncly auth status`, a command of core that reports keys
- List command
  - Def.: a command whose verb is `list`, such as `ncly skill list`
  - Project note: [Completion values](contract.md#completion-values) defines when completion may run one: never one that writes a user file or bills
- Command that core calls
  - Def.: a command of an installed extension that core starts itself for one of its own commands: completion's list command or doctor's status command
  - Project note: [Commands that core calls](contract.md#commands-that-core-calls) defines the rules that both share. [Completion values](contract.md#completion-values) and [Status command](contract.md#status-command) keep the conditions, the limit, and the reading of the answer of each
  - Distinct from: an ordinary dispatch, in which core starts a command for its caller, as when an agent calls a status command alone
- Grant
  - Def.: the permission, given by a human in a terminal, for one extension to receive one key
  - Project note: [Grants](contract.md#grants) binds it to the extension's domain and source, the path where discovery found the extension until taps
- Optional key
  - Def.: a service that a command declares but runs without, such as a harness key, whose harness then uses its own login
  - Project note: the `keys` table of a [manifest](contract.md#manifest) marks it, and a missing one never fails `ncly auth status` or `ncly doctor`
  - Distinct from: a conditional program, which a command needs only for some configs
- Service
  - Def.: a network service whose key `ncly auth` manages, such as Deepgram, named `deepgram`
  - Project note: [Keys](contract.md#keys) defines its name and the variable of its key, such as `DEEPGRAM_API_KEY`. A command declares its services in the `requires.keys` of its manifest
  - Distinct from: services, layer 2, and the service field of a keychain entry
- Interactive mode
  - Def.: the mode in which `ncly` may ask questions. [Modes](contract.md#modes) defines when it applies
  - Distinct from: non-interactive mode
- Non-interactive mode
  - Def.: every case outside interactive mode. A missing value fails instead of opening a form
- Plain prompt
  - Def.: a question in the terminal that the human answers by typing a line, such as a `[y/N]` confirmation, whose look [Modes](contract.md#modes) describes
  - Project note: [ncly auth](core-spec.md#ncly-auth) defines it. [M07](../milestones/M07-forms.md) replaced the key prompt of M06, which turned echo off, with a masked field, and styled the confirmation, which stays a prompt, as [D059](decisions/D059-no-buttons.md) decides
  - Distinct from: a form, including the masked field of a key, which the shared tui package builds and [Modes](contract.md#modes) describes

## Operations

### Relations

- A command performs an operation; an operation that bills or leaves reusable work writes a run record
- A run record references its artifacts and the evidence of each step
- Record support persists a step's intent before an effect that cannot safely be repeated, and derives retry safety from the effects that the steps declare
- Resume continues a run record after checking its evidence, and never repeats an unknown effect
- A run ID names one run; a call with a run ID that a record holds continues that run instead of starting another
- A run record keeps what the call states as inputs, and what preparation picked as resolved values; a call with the same run ID compares only the inputs
- Dry run and execution share one preparation

### Definitions

- Operation
  - Def.: the work prepared and performed by a command. The domain owns its steps
- Dry run
  - Def.: a simulation that uses the same preparation as execution and changes nothing outside Nicely's cache
- Run record
  - Def.: the durable evidence of one operation, identified by `run_id`, including its step outcomes and artifact references
- Run ID
  - Def.: the identifier of one run, chosen by the caller with `--run-id`, or generated
  - Project note: [Run IDs](contract.md#run-ids) defines what a call with an existing run ID does
- Resolved value
  - Def.: a value that preparation picked for a run from the config or the machine, such as its model and effort
  - Project note: a run record keeps it under `resolved`, and every later attempt reads it from there, as [Records and evidence](contract.md#records-and-evidence) defines
  - Distinct from: an input, which the call states, and which a call with the same run ID compares
- Effect
  - Def.: what an operation does outside Nicely's cache and state and the run's temporary files, such as a user write, network access, or a paid request
  - Project note: each step of a run record declares the strongest effect that it can make, as [Record format](contract.md#record-format) lists
- Step boundary
  - Def.: the moment just before or just after record support persists a change to a step: its intent, the evidence of one artifact, or its final status
  - Project note: the boundary scenario of a recorded operation kills it at every step boundary, as [Work on a card](../../AGENTS.md#work-on-a-card) asks
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
  - Project note: an extension's skill is named after its domain. The bundled `nicely` skill teaches core

## Configuration and text

### Relations

- The local config overrides the shared config on one machine
- The pseudo-locale marks every catalog string, so text outside a catalog stands out

### Definitions

- Shared config
  - Def.: `config.toml`, the setup the user wants, which can travel between machines
- Local config
  - Def.: `config.local.toml`, which holds what belongs to one machine and overrides the shared config
- Catalog
  - Def.: one TOML file of translated text for one language, whose entries each have an ID, a description for the translator, and plural forms
  - Project note: core, the SDK, and each extension keep their own. An external extension's catalogs sit in `ncly-<domain>.locales/` beside its manifest
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
- A stack holds layers, bottom to top; a layer is one branch and its pull request, based on the layer below
- A principle links its decisions; a decision records what was decided, why, and what was rejected

### Definitions

- Milestone
  - Def.: a numbered stage of work in `docs/milestones/`, one capability that a sentence can demonstrate
  - Project note: its status is `planned`, `ready`, `active`, or `done`. M99 is the parking lot, and its status stays `open`
- Card
  - Def.: one pull request of a milestone, with its owner, its dependencies, what to read, and what proves it
- Stack
  - Def.: a chain of pull requests that build on each other, as `gh stack` manages them
- Principle
  - Def.: a rule that shapes every design choice, in [principles.md](principles.md)
- Decision
  - Def.: one file in [decisions/](decisions/README.md) that records what was decided, why, and what was rejected

## Design vocabulary

### Relations

- Primitives compose into commands and extensions
- Layers build abstractions from the bottom up: foundation → runtime layer → application layer
- An interface hides its implementation
- An additive extension adds to its host; a behavioral extension changes it

### Definitions

- Abstraction
  - Def.: something simpler that hides complexity behind it
  - Project note: each layer is an abstraction over the layers below it
- Layer
  - Def.: a level of abstraction that depends only on the levels below it
  - Project note: platform, contract, services, core, and extensions, drawn in [architecture.md](architecture.md#layers)
  - Distinct from: a layer of a [stack](#planning), which `just check-layers` checks
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
  - Project note: [Primitives](contract.md#primitives) lists Nicely's
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
