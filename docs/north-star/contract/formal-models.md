# Formal models

This part says how a repository checks its TLA+ models with TLC, and how `just check` reads their outcomes.
[D068](../decisions/D068-verification-map.md) puts these checks at level 2.
Core owns the TLC runner, the package `sdk/verification`. Each repository owns its models, their mapping to its code, and their expected outcomes.

## The suite

`verification/tlc.json` of the repository pins the tools and lists the checks:

| Key | Holds |
|---|---|
| `java` | The Java of mise that runs TLC, such as `java@temurin-21.0.13+9.0.LTS` |
| `tlc_url` | The URL of the TLC jar |
| `tlc_sha256` | The SHA-256 of the jar |
| `checks` | Each check: its `name`, its `model` (`.tla`), its `config` (`.cfg`), and an optional `violation` |

A check without `violation` must explore every state with no error, and TLC exits 0.
A check with `violation` must find that invariant violated, and TLC exits 12. It proves that the invariant can fail, such as on a past mistake.
The paths go from the root, and stay in the repository.
A key that the suite does not define fails, so no pin hides in a key that nothing reads.

The pin of core is TLC v1.7.4, a release whose jar has not changed since 2024-08-08.
On 2026-11-02, the URL of TLC v1.8.0, a prerelease, served other bytes than the SHA-256 that video-convert pinned.
A new pin needs a review of the bytes that it names.

## `just verify-models`

`just verify-models [jar]` runs the suite in a host [owned run](owned-runs.md).
It needs mise, which installs the pinned Java when it is missing.

1. It reads the suite, and stops on an error before any download.
2. It reads the local jar, or downloads `tlc_url`, and checks its SHA-256. A local jar needs the pinned SHA-256 too. Changed bytes stop the run before Java starts.
3. It runs `java -version`, then each check with one worker, so a rerun gives the same record.
4. It accepts each check only with its exit, the message of its outcome, and the statistics of the graph. A pass also needs an empty queue.
5. It writes `verification/tlc-record.json` only when each check gives its expected outcome, and no input changed during the run.

The folder of the run stays, with the jar, the log of each program, and the states of TLC.
Its answer gives each outcome, the record, and that folder.

## The record

`verification/tlc-record.json` holds the pins, the SHA-256 of each input, and the outcome of each check.
The inputs are the suite, and each `.tla` and `.cfg` file of the folder of each model and each configuration.
So a module that a model extends is an input too.
A complete exploration records its distinct states and its depth. A violation records only its invariant, because TLC can stop at another state.

`just check` of core and `extensionrepo check` read only the suite, its inputs, and the record, so they need no Java and no network.
A repository without `verification/tlc.json` has no record to check.
They fail when the record is missing, holds another pin, is of other inputs, or holds an outcome that the suite does not expect.

The record binds the model, not the code that it models.
A change of that code needs a review of the mapping of the model.

## The model of core

Core's first model is its [finish gate](finish-gate.md), in `verification/finish/`, whose `README.md` maps it to the code.
Its configuration `PreFix` is the gate at the parent of 59d33f38, where a finding of this round with no state passed.
TLC must find `NoOpenPass` violated there, and `Current` must explore every state with no error.
