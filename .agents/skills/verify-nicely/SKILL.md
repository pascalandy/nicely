---
name: verify-nicely
description: "Use when verifying that the real ncly CLI of the nicely repository works as a user runs it, such as the Verify step of a card in AGENTS.md, or when driving a feature of features/ with a scratch HOME."
---

# Verify nicely

`ncly` is a short-lived CLI: no server, no process that outlives a call. A run builds it once, then drives each feature through a wrapper that keeps every file in a scratch folder. `scripts/verify.sh` does each step, from the root of a nicely checkout, on macOS and Linux. It needs `bash`, `git`, and Go: `go version` must answer, or install Go from https://go.dev/dl/.

Copy this checklist and tick each step:

```
- [ ] 1. Launch: run="$(just verify)"
- [ ] 2. Doctor: --version names the build, and doctor passes
- [ ] 3. Drive: each feature that the card affects, from features/
- [ ] 4. Cleanup: just verify cleanup "$run"
- [ ] 5. Evidence: the transcripts outlive the cleanup, and the report quotes them
```

## Launch

```bash
run="$(just verify)"
```

`just verify` runs `scripts/verify.sh start`. It builds `ncly` from the checkout as it is, uncommitted changes included, into `$run/build/`. It makes a scratch home in `$run/scratch/home/`, with its own XDG folders, and installs the Go fixture `hello` there as an external extension, in `.local/share/nicely/extensions/`, with its manifest, catalogs, and skill beside it. It writes the wrapper `$run/scratch/bin/ncly`, which sets `HOME`, the XDG folders, and a `PATH` without any folder that holds an `ncly-<domain>` file, and unsets `NCLY_CONFIG`, which would point the config elsewhere, so every later command reaches only the scratch folder. On macOS, the auth feature points `HOME` back at the real home, as [auth](features/auth.md) says. Its last line on stdout is the run folder, under `${XDG_CACHE_HOME:-$HOME/.cache}/nicely/verify/`. Done when that folder holds `scratch/bin/ncly`.

## Doctor

```bash
just verify step "$run" doctor --version
just verify step "$run" doctor doctor
```

`step <run> <feature> <ncly args...>` runs the wrapper with those arguments and the caller's stdin. It prints the call, the time it started in ISO 8601 with its offset, its stdout, its stderr, and its exit code. It appends them to `$run/evidence/<feature>.txt`, and exits with the code of `ncly`. When it cannot write that transcript, it exits 1 and leaves the outputs of the call in `$run/scratch/`. Done when `--version` prints `ncly v<version>-<commit>`, where `<commit>` is the first line of `$run/evidence/commit.txt`, followed by `+dirty` when its second line says `dirty`, and `ncly doctor` exits 0 with a `pass` line for `extensions.hello` whose path lies under `<run>/scratch/home`, and an `auth.keychain` line: a `pass` when the keychain answers, or a `warn` with `KEYRING_ENV_ONLY` when it does not. When doctor fails, stop and report its output: a feature driven on a broken build proves nothing.

## Drive

Read [features/README.md](features/README.md), then the file of each feature that the card affects: [hello](features/hello.md), [auth](features/auth.md), [terminal](features/terminal.md), [agent](features/agent.md), [transcript](features/transcript.md), [transcript-run](features/transcript-run.md), or [template](features/template.md). Each file names the commands that reach the feature and the end state that proves it. A feature that needs a fixture beside `hello`, such as `auth`, `agent`, `transcript`, or `template`, installs it with `just verify prepare "$run" <feature>`. Run each command through `just verify step "$run" <feature> ...`, or through `just verify tmux "$run" <feature> ...` for a step that needs a terminal inside tmux, such as those of `terminal` and the steps of `auth` in a terminal, never through another `ncly`, so the transcript holds every call. Done when every command of the feature ran and its end state holds, or the report names the command whose result differs, with the expected and the actual result.

Read [shared domains](features/shared-domains.md) when verifying named extensions in a shared domain.

A card that changes the steps or the end state of a feature updates its file first, and a card that adds a feature adds its file and its line in the index.

## Cleanup

```bash
just verify cleanup "$run"
```

`cleanup` removes `$run/scratch/` and `$run/build/`, and keeps `$run/evidence/`. It refuses a folder that `start` did not create. When the run prepared the `auth` feature, it first logs the fake key out of the real keychain with `ncly auth logout`, then checks with `ncly auth status` that the keychain answers without the key, with `AUTH_MISSING`, both in `$run/evidence/cleanup.txt`. Its stderr says `logged the fake key of verifynicely out of the keychain` only when the logout removed a key, and `found no fake key` otherwise. When the key may remain, it keeps the scratch folder and exits 1, so a second cleanup can try again. Run it after a failed run too, so no scratch folder and no fake key stay. Done when it lists the evidence files and `$run/scratch` is gone.

```bash
just verify prune --keep 5
```

`prune --keep <n>` keeps the newest `<n>` runs, and removes the others with their evidence. It keeps a run whose scratch folder remains, because its cleanup has not ended. Done when it names each run that it removed or kept, or says that it found no run.

## Evidence

```bash
just verify quote "$run" hello
```

`quote <run> [<feature>]` prints the transcripts with `<run>` for the run folder and `/Users/me` for the home folder, ready to paste in a pull request. After auth, also run `just verify quote "$run" cleanup`, whose `cleanup.txt` proves that the fake key left the keychain. Proof standards:

- Drive the path a user takes, through installed extensions and real commands, never a test hook
- Capture the action and the resulting state, such as a second command that reads what the first one wrote
- Check side effects, such as files or a keychain entry, beside the answer
- Quote the lines that prove each end state, after the cleanup, from the quoted transcript

Done when, after the cleanup, `quote` prints each transcript, `cleanup.txt` included after auth, and the report quotes the lines that prove each end state.

## In a card

Step 6 of a card in `AGENTS.md` runs this skill.

- An agent that did not write the change runs the skill, on a small model of the other family than the implementer's, at effort `xhigh`, in the same checkout, so it sees uncommitted changes.
- The pull request quotes the lines that prove each end state. It names the commit that `evidence/commit.txt` records, followed by `dirty` when the file says so. A dirty run proves the checkout as it was, not that commit alone.
- A fix that changes a verified behavior needs a run that covers it. A run at the top of the stack covers it when it drives the features that the fix affects. A restack keeps the recorded commit as it is: the pull request names the new head beside it, and the run that covers each later fix.
- A card that reaches nothing the scenarios replace, such as the real keychain, a real terminal, a harness, `uv`, or a paid service, marks the step `n/a`.

Done when the quoted evidence names its commit and proves the end state of each feature that the card affects.
