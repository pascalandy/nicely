# UAT protocol

This file answers one question: how is a UAT of any extension run so that it can be repeated? The `qa.md` of each extension adds only its deltas. [north-star.md](north-star.md) says what every UAT aims for.

## Fixed rules

- **Hosts.** Run on the macOS test host and on the Linux test host, never on the author's workstation. Name the host in the issue body, never in its title.
- **Paid calls.** A demo call to Codex uses `gpt-6-luna` at effort `low`. It was the cheapest model that Codex listed on 2026-10-29.
- **Runner.** An agent runs each UAT by default, on macOS too, with screen control such as `cua`. It calls a human only for a case of [Human escalation](#human-escalation).
- **Install.** Install the extension from its release tag, with the commands of its `README.md`. Record the versions of `ncly`, the tag, its commit, and the wrapped program. To keep the host's own install, or to test `ncly` from `main`, install as [Isolated install](#isolated-install) says.
- **Fixtures.** Real data is a prerequisite that the issue names. Originals stay read-only and locked, as [fixtures.md](fixtures.md) says.
- **Deletion.** Move work folders to the trash. Never modify or unlock an original.

## One run

1. **Prerequisites.** In a Nicely checkout, run `just preflight` with the fixture folders that the issue names. It checks the host, then the checksum and the lock of each fixture. Run it as [Desktop session](#desktop-session) says. The offline Go probe may populate Go caches; it leaves source files and fixtures unchanged. Stop when it exits nonzero.
2. **New folder.** Create `ncly-uat-<issue>-<YYYYMMDD-HHMM>/` and copy the fixture into it. Check the checksum of the copy.
3. **Discovery.** Give a fresh agent one sentence, without hints or `NCLY_*` variables. Its instructions hold only the agent line of the [README](../../README.md). It must reach the command through `ncly skill view nicely` and `ncly describe`.
4. **Dry run.** When the command has a dry run, it comes before the run and passes on valid input.
5. **Run.** Keep the `--json` answer, the exit code, stdout, and stderr.
6. **Independent check.** Check the output with a third-party tool, never with the extension.
7. **Error cases.** Run each invalid input of the extension's `qa.md`, as a run, and as a dry run when the command has one.
8. **Original.** Check that its checksum and its lock are unchanged.
9. **Evidence.** Keep it in `~/.cache/ncly-uat-<issue>-runs/<run>/`. Then move the work folder to the trash.

Make at least 3 runs, to measure discovery and determinism.

## Isolated install

An isolated install leaves the host's `ncly`, extensions, config, and keys unchanged. `just acceptance` in an extension repository does these steps itself.

1. **Tools first.** Before you change a variable, find the path of `just` and `go`: `mise which just` for a tool that mise installed, else `command -v just`. Call those paths. Once `XDG_CONFIG_HOME` moves, a shim of mise outside a folder that pins the tool fails with `No version is set for shim`.
2. **Scratch folders.** Create one folder with `mktemp -d`. Point `XDG_BIN_HOME`, `XDG_DATA_HOME`, `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, and `XDG_CACHE_HOME` to subfolders of it.
3. **`ncly` from `main`.** In a Nicely checkout of `origin/main`, run `just install ncly`, and put the folder of `XDG_BIN_HOME` first in `PATH`. Add the first-party extensions that the UAT needs, such as `just install ncly agent`. Record `ncly --version`.
4. **The extension.** Run `just install` in its clone, under the same variables. It installs into the scratch data folder.
5. **End.** Move the scratch folder to the trash.

## Human escalation

The agent stops and calls Pascal only in these cases. Each UAT issue names its own in its Human escalation section.

- The OS or the keychain asks for a password.
- A window takes the keyboard, and no action of the agent reaches it. `cua do click` cannot reach a SecurityAgent window, such as one about the keychain. Escape presses its Cancel button.
- The OS asks for a permission, such as access to files on a removable volume. A desktop-session process that reads an external volume, such as a Go cache, opens this prompt. `just preflight` warns about it first, see [Desktop session](#desktop-session).
- A step stays blocked after one retry.

Never press **Reset To Defaults** in a keychain window: it can replace the real keychain.

## Desktop session

On macOS, some steps run in the desktop session, such as the steps that use the login keychain. The agent reaches it through `cua do shell`.

- Run `just preflight` over SSH or in a terminal of the desktop session, such as Ghostty. Through `cua do shell`, a checkout on an external volume can open the permission prompt before preflight warns.
- On macOS, preflight warns when the checkout, a Go cache, or a fixture folder is on an external volume. It reads whether Pascal allowed Cua Spaces to read removable volumes. That read needs Full Disk Access. Without it, preflight warns that it cannot check the grant, so run it over SSH.
- When Cua Spaces has no grant, ask Pascal to allow it once. You can also run the steps in a terminal of the desktop session.
- Give `cua do shell` one script file at a path without spaces, such as `cua do shell /bin/bash /tmp/step.sh`. It joins its arguments into one `/bin/sh -c` line and drops their quotes and backslashes.
- In that script, set `PATH` and change to the folder of the step. `cua do shell` starts in the home folder, with only the system folders in `PATH`.
- A permission prompt or a keychain window is a case of [Human escalation](#human-escalation).

## No-spawn probe

This probe proves that a replay starts no harness, such as Codex.

1. Count the harness processes with `pgrep -x codex | wc -l`, and its sessions with `find ~/.codex/sessions -type f | wc -l`.
2. Repeat the call under its run ID, with a `PATH` that holds no folder with the harness. Under that `PATH`, `command -v codex` prints nothing, and `ncly` cannot start the harness, because it finds each program only in `PATH`.
3. Count both again. Equal counts and a replay that succeeds prove that nothing started.

## Criteria families

| Family | Requires |
|---|---|
| Discovery | A fresh agent finds the command without hints |
| Preview | When the command has a dry run, it refuses each error of the shared preparation with the code of the run |
| Result | A third-party tool accepts the output |
| Fidelity | The output keeps the input properties that the `spec.md` of the extension names in `## Acceptance` |
| Errors | Invalid input gets its own code, the right exit code, and a message that names the file |
| Integrity | The input and the original are unchanged, and no partial file remains |
| Contract | With `--json`, a failure leaves stdout empty and the error object ends stderr, as [Output](../north-star/contract/output.md) says |
| Automation | A test as code covers each acceptance criterion, as [Automation](north-star.md#automation) says |
| Economy | The output and the skills are short and correct, and the default flags suit the common case, as [Economy](north-star.md#economy) says |

## Severity and verdict

- **Severity of a defect:** blocking, major, minor, or cosmetic.
- **Verdict of a UAT:** accepted, accepted with reservations, or refused.
- A blocking or major defect gives `refused` and a linked `2-type:bug` issue.
- A reservation is an unmeasured criterion or a minor defect that Pascal accepts. Each one gets an issue or a next test.
- An output over its [budget](north-star.md#budgets) is a minor defect.
- Each [Automation](north-star.md#automation) finding is a reservation. Examples are a criterion that no test as code covers, and a manual step that a test could replace.

## Report

Post the report as a comment of the issue, in this order:

1. Rounds
2. Hosts and versions
3. Fixtures
4. Results by family
5. Automation: the findings that [Automation](north-star.md#automation) lists
6. Economy: the findings and the measurements of [Economy](north-star.md#economy)
7. Defects, with severity and issue
8. Reservations
9. Verdict
10. Next tests

Only Pascal closes the issue.
