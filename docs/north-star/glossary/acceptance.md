# Acceptance

## Relations

- A UAT follows the UAT protocol, runs on a copy of a fixture, and adds its verdict to the ledger of [docs/qa/](../../qa/README.md)
- An agent runs a UAT after a preflight, and calls a human only for a case of human escalation. The human accepts the delivery

## Definitions

- UAT
  - Def.: user acceptance testing, in which a human accepts a delivery of Nicely on a real host, with real data
  - Project note: each UAT is a GitHub issue with the label `2-type:uat`
  - Distinct from: verify-nicely, which proves one card on the real `ncly`
- UAT protocol
  - Def.: the steps that every UAT repeats, in [protocol.md](../../qa/protocol.md). The `qa.md` of an extension adds only its deltas
  - Distinct from: the contract, whose two scopes are never called protocols
- Fixture
  - Def.: real data that a UAT runs on, read-only and locked, listed in [fixtures.md](../../qa/fixtures.md)
  - Project note: a made fixture, such as an empty file, comes from the run itself, in its work folder
  - Distinct from: a fixture extension under `testdata/extensions/`, which the tests use to prove the SDK
- Preflight
  - Def.: `just preflight`, which reads the host of a UAT and its fixture folders, and changes nothing
  - Distinct from: `ncly doctor`, which checks what Nicely needs, not what a UAT needs
- Human escalation
  - Def.: a case where the agent of a UAT stops and calls a human, as [protocol.md](../../qa/protocol.md#human-escalation) lists them
  - Project note: each UAT issue names its cases in a section of that name
