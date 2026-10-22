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

See [Architecture](glossary/architecture.md).

## Contract

See [Contract](glossary/contract.md).

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
- Run
  - Def.: one recorded operation, from its first call to its last, which one run ID names and one run record describes
  - Project note: each call that safely starts its work again is one more attempt, as `attempts` in [Record format](contract/record-format.md) counts
  - Distinct from: the verb `run`, as in `ncly agent run`
- Run record
  - Def.: the durable evidence of one operation, identified by `run_id`, including its step outcomes and artifact references
- Run ID
  - Def.: the identifier of one run, chosen by the caller with `--run-id`, or generated
  - Project note: [Run IDs](contract/run-ids.md) defines what a call with an existing run ID does
- Resolved value
  - Def.: a value that preparation picked for a run from the config or the machine, such as its model and effort
  - Project note: a run record keeps it under `resolved`, and every later attempt reads it from there, as [Records and evidence](contract/records-and-evidence.md) defines
  - Distinct from: an input, which the call states, and which a call with the same run ID compares
- Effect
  - Def.: what an operation does outside Nicely's cache and state and the run's temporary files, such as a user write, network access, or a paid request
  - Project note: each step of a run record declares the strongest effect that it can make, as [Record format](contract/record-format.md) lists
- Item
  - Def.: one input of an operation, such as one URL, with its key, its output folder, and its steps
  - Project note: a run record and a dry-run plan hold one item per input, in input order, and a batch answers one result per item
  - Distinct from: an entry that a `list` command shows, which [Usage](contract/usage.md) also calls an item
- Step
  - Def.: one unit of work of an item, such as `transcribe`, which the domain names and whose effect it declares
  - Project note: record support persists its intent and its final status, as [Record format](contract/record-format.md) defines
  - Distinct from: a step of a card's checklist in AGENTS.md
- Step boundary
  - Def.: the moment just before or just after record support persists a change to a step: its intent, the evidence of one artifact, or its final status
  - Project note: the boundary scenario of a recorded operation kills it at every step boundary, as [the scenario rules](../../testdata/script/README.md) ask
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

See [Design vocabulary](glossary/design-vocabulary.md).

## Terms to avoid as canonical forms

- plugin, add-on, module, for an extension
  - Use instead: extension
  - Reason: one word names a domain added to core, bundled or external
- headless mode
  - Use instead: non-interactive mode
  - Reason: [Modes](contract/modes.md) names the two modes
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
