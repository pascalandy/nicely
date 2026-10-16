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

`just verify` runs `scripts/verify.sh start`. It builds `ncly` from the checkout as it is, uncommitted changes included, into `$run/build/`. It makes a scratch home in `$run/scratch/home/`, with its own XDG folders, and installs the Go fixture `hello` there as an external extension, in `.local/share/nicely/extensions/`, with its manifest, catalogs, and skill beside it. It writes the wrapper `$run/scratch/bin/ncly`, which sets `HOME`, the XDG folders, and a `PATH` without any folder that holds an `ncly-<domain>` file, so every later command reaches only the scratch folder. Its last line on stdout is the run folder, under `${XDG_CACHE_HOME:-$HOME/.cache}/nicely/verify/`. Done when that folder holds `scratch/bin/ncly`.

## Doctor

```bash
just verify step "$run" doctor --version
just verify step "$run" doctor doctor
```

`step <run> <feature> <ncly args...>` runs the wrapper with those arguments and the caller's stdin, prints the call, its stdout, its stderr, and its exit code, appends them to `$run/evidence/<feature>.txt`, and exits with the code of `ncly`. Done when `--version` prints `ncly v<version>-<commit>`, where `<commit>` is the first line of `$run/evidence/commit.txt`, followed by `+dirty` when its second line says `dirty`, and `ncly doctor` exits 0 with a `pass` line for `extensions.hello` whose path lies under `<run>/scratch/home`. When doctor fails, stop and report its output: a feature driven on a broken build proves nothing.

## Drive

Read [features/README.md](features/README.md), then the file of each feature that the card affects: [hello](features/hello.md). Each file names the commands that reach the feature and the end state that proves it. Run each command through `just verify step "$run" <feature> ...`, never through another `ncly`, so the transcript holds every call. Done when every command of the feature ran and its end state holds, or the report names the command whose result differs, with the expected and the actual result.

A card that changes the steps or the end state of a feature updates its file first, and a card that adds a feature adds its file and its line in the index.

## Cleanup

```bash
just verify cleanup "$run"
```

`cleanup` removes `$run/scratch/` and `$run/build/`, and keeps `$run/evidence/`. It refuses a folder that `start` did not create. Run it after a failed run too, so no scratch folder stays. Done when it lists the evidence files and `$run/scratch` is gone.

## Evidence

```bash
just verify quote "$run" hello
```

`quote <run> [<feature>]` prints the transcripts with `<run>` for the run folder and `/Users/me` for the home folder, ready to paste in a pull request. Proof standards:

- Drive the path a user takes, through installed extensions and real commands, never a test hook
- Capture the action and the resulting state, such as a second command that reads what the first one wrote
- Check side effects, such as files or a keychain entry, beside the answer
- Quote the lines that prove each end state, after the cleanup, from the quoted transcript

Done when `quote` prints the transcript of every feature driven after the cleanup, and the report quotes the lines that prove each end state.
