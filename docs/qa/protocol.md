# UAT protocol

This file answers one question: how is a UAT of any extension run so that it can be repeated? The `qa.md` of each extension adds only its deltas.

## Fixed rules

- **Hosts.** Run on the macOS test host and on the Linux test host, never on the author's workstation. Name the host in the issue body, never in its title.
- **Install.** Install the extension from its release tag, with the commands of its `README.md`. Record the versions of `ncly`, the tag, its commit, and the wrapped program.
- **Fixtures.** Real data is a prerequisite that the issue names. Originals stay read-only and locked, as [fixtures.md](fixtures.md) says.
- **Deletion.** Move work folders to the trash. Never modify or unlock an original.

## One run

1. **Prerequisites.** Check that the fixture volume is mounted, its checksum is `OK`, and its lock holds. Otherwise stop.
2. **New folder.** Create `ncly-uat-<issue>-<YYYYMMDD-HHMM>/` and copy the fixture into it. Check the checksum of the copy.
3. **Discovery.** Give a fresh agent one sentence, without hints or `NCLY_*` variables. It must reach the command through `ncly skill view nicely` and `ncly describe`.
4. **Dry run.** It comes before the run and passes on valid input.
5. **Run.** Keep the `--json` answer, the exit code, stdout, and stderr.
6. **Independent check.** Check the output with a third-party tool, never with the extension.
7. **Error cases.** Run each invalid input of the extension's `qa.md`, as a dry run and as a run.
8. **Original.** Check that its checksum and its lock are unchanged.
9. **Evidence.** Keep it in `~/.cache/ncly-uat-<issue>-runs/<run>/`. Then move the work folder to the trash.

Make at least 3 runs, to measure discovery and determinism.

## Criteria families

| Family | Requires |
|---|---|
| Discovery | A fresh agent finds the command without hints |
| Preview | The dry run refuses what the run refuses, with the same code |
| Result | A third-party tool accepts the output |
| Fidelity | The output keeps the input properties that `## Acceptance` in the `spec.md` of the extension names |
| Errors | Invalid input gets its own code, the right exit code, and a message that names the file |
| Integrity | The input and the original are unchanged, and no partial file remains |
| Contract | With `--json`, a failure leaves stdout empty and the error object ends stderr, as [Output](../north-star/contract/output.md) says |

## Severity and verdict

- **Severity of a defect:** blocking, major, minor, or cosmetic.
- **Verdict of a UAT:** accepted, accepted with reservations, or refused.
- A blocking or major defect gives `refused` and a linked `2-type:bug` issue.
- A reservation is an unmeasured criterion or a minor defect that Pascal accepts. Each one gets an issue or a next test.

## Report

Post the report as a comment of the issue, in this order:

1. Rounds
2. Hosts and versions
3. Fixtures
4. Results by family
5. Defects, with severity and issue
6. Reservations
7. Verdict
8. Next tests

Only Pascal closes the issue.
