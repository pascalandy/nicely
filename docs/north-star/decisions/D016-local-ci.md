# D016 Run checks locally and on GitHub, and release by hand

Decided 2026-10-05. Checks run through `just`, Lefthook, and `gh signoff`. External pull requests go through the maintainer's agent and `just signoff`. A release runs only when an agent starts the release workflow.

**Scope update, 2026-10-29.** GitHub Actions also runs `just check` for each push to `main` and each pull request, as [#373](https://github.com/pascalandy/nicely/issues/373) asks. Local checks and `just signoff` stay, and releases stay manual.

**Why.** Pascal often hits GitHub Actions limits, and local checks are faster. But a push can skip the pre-push hook of Lefthook, so nothing else protected `main`. Nicely is public, so standard GitHub runners run its checks for free.
