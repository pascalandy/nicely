# Nicely architecture

This file says how Nicely is built: its layers, who owns each part, and where each domain lives.

## Core and extensions

Nicely is a small core, and everything it does for a user is an extension, in the spirit of Pi v1: a minimal host that is easy to extend and to compose.

- **Core** is the host: dispatch, help, completion, `describe`, `doctor`, `auth` with grants, `run`, and `tap`. It knows no domain.
- **Extensions** add domains, such as `transcript`. All of them share one manifest format and one contract, and they run with the rights of the user, because Nicely is not a sandbox. A **bundled** extension is compiled into `ncly`. Only what an agent needs to operate Nicely is bundled: `skill`, with the `nicely` skill. An **external** extension is an executable named `ncly-<domain>` beside its manifest. The first-party ones, such as `agent` and `transcript`, live in `extensions/` of this repository, and Pascal's personal tools, such as fleet sync, live in his private tap.

The parts form a stack of layers:

```
4  Extensions  bundled: skill, with the nicely skill
               external: agent, transcript, markdown, video, image, docs; fleet in a private tap
3  Core        dispatch, help, completion, describe, doctor, auth and grants, run, tap
2  Services    config, i18n, program runner, run records
1  Contract    answer envelope, code to exit, modes, declaration and manifest
0  Platform    durable writes, locks, keychain, trash, process tree
```

- A layer uses only the layers below it.
- Each part has one owner: one package and one spec section.
- Layers 0 to 2 that extensions need form the public SDK in `sdk/`, which core uses too. The keychain stays in core, and keys reach an extension only through grants.
- An extension reaches core through its environment and the CLI, and another extension only through the CLI, the same interface that agents use.
- Adding a domain never edits core.

Core says where a feature lives, not when it ships. The milestones decide timing.

## Domains and commands

| Command | What it does | Kind | Milestone |
|---|---|---|---|
| `completion` | Prints shell completion scripts | core | M00 |
| `describe` | Describes installed commands and their capabilities | core | M04 |
| `doctor` | Checks tools, keys, extensions, and components | core | M05 |
| `auth` | Stores keys and grants them to extensions | core | M06, grants in M08 |
| `run` | Lists, inspects, and explicitly resumes recorded operations | core | M10, resume in M15 |
| `tap` | Adds and syncs repositories of skills and extensions | core | M12 |
| `skill` | Lists and views skills for agents, including the `nicely` skill | bundled | M03 |
| `agent` | Runs a task through a harness and a profile | external | M09 to M11 |
| `transcript` | Transcribes YouTube and Zoom audio, then summarizes it | external | M13 to M16 |
| `markdown`, `video`, `image` | Everyday tasks that need no configuration | external | M18 to M20, after the inventory of M17 |
| `docs` | Gathers project docs into one private site | external | M21 |

Core reserves the names of the core and bundled rows. `transcript` is not the first thing a visitor should try, because it needs a Deepgram key. First impressions come from the everyday domains.

## Where the details live

- [contract.md](contract.md) holds the agent contract and the commands of core. Each extension's spec lives in `extensions/<name>/spec.md`, such as [transcript's](../../extensions/transcript/spec.md). Read the sections that your card links before changing a command.
- [decision-records.md](decision-records.md) explains why each rule exists. Read it before you propose to change a rule.
- [dev-preferences.md](dev-preferences.md) records how Pascal wants changes made, such as deciding the look with a published mockup.
- The milestone files in `docs/milestones/` list the work as cards, in order. `just next` prints the card to do, `just status` shows the progress, and [M99](../milestones/M99-parking-lot.md) parks every other idea.

## Boundaries to keep through the milestones

| Owner | Responsibility |
|---|---|
| Command declaration and manifest | Arguments, flags, result contract, effects, prerequisites, and capabilities used by discovery, help, completion, and validation |
| Extension host | Discovery, manifests, dispatch, answer checks, and grants |
| Extension | Its domain's preparation, business steps, result verification, the evidence that permits resume, its catalog, and its skill |
| Record support | Run IDs, durable records, execution locks, and publication of the extension's result |
| Program runner | Child environment, timeouts, signals, and process cleanup |
| Platform support | Atomic file replacement, locks, keychain, and trash |

M00 fixed these responsibilities and the contract. Each implementation arrives with its first consumer: the host and the program runner in M01, the SDK in M02, and records with `ncly agent run` in M10.
