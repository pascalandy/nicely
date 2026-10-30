# D016 Run checks locally and on GitHub, and release by hand

Decided 2026-10-05. Changed 2026-10-29 by [#373](https://github.com/pascalandy/nicely/issues/373), and confirmed by Pascal in [#381](https://github.com/pascalandy/nicely/issues/381). Checks run through `just`, Lefthook, and `just signoff`, whose `signoff` status the ruleset of `main` requires. GitHub Actions runs `just check` again for each push to `main` and each pull request. That check is advisory: the ruleset does not require it. External pull requests go through the maintainer's agent and `just signoff`. A release runs only when an agent starts the release workflow.

**Why.** Local checks are faster, and `just signoff` proves the exact commit that reaches `main`. A push can skip the pre-push hook of Lefthook, so a second run on a clean runner shows a broken `main`. Nicely is public, so standard GitHub runners run that check for free.

`just signoff` uses `sdk/extensionrepo` in core and extension repositories.
The program freezes the full HEAD SHA and creates a detached scratch worktree at that SHA.
It clears `JUST_*` and `GIT_*` from its child environment before it resolves the repository or runs a check.
It removes caller `GOFLAGS`, whose overlays and module flags can replace committed source or dependencies.
It disables `GOWORK`, which can select external workspace modules, and `GOENV`, which can restore caller Go settings.
It runs that worktree's `just check` once, including its dependencies.
A failed check, a changed HEAD, or a changed tracked file prevents signoff.
Uncommitted changes in the caller cannot fix the checked commit.
After the checks pass, `gh signoff --commit <sha>` signs only the frozen SHA.
Success prints `Signed off on <sha>` with the full SHA.

Each run prints the path of its JSON receipt under Nicely's cache folder, in `signoff/`.
That folder must be outside the repository.
The receipt records the SHA, host OS, tool versions, and each invoked gate's exit code.
The current gate is `just check`.
The receipt exists before signoff and then records the exit code of `gh signoff`.
The run also keeps its check log, including failures.
The scratch worktree goes when the program returns.
Every run repeats the checks, even when a receipt already exists.

**Rejected.** Local checks alone, because a skipped hook then leaves a broken `main` unseen, as #373 found. GitHub Actions as the only gate, because Pascal often hits its limits, and local checks are faster. A required `check` status, because `signoff` already gates `main`.
