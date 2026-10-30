# D016 Run checks locally and on GitHub, and release by hand

Decided 2026-10-05. Changed 2026-10-29 by [#373](https://github.com/pascalandy/nicely/issues/373), and confirmed by Pascal in [#381](https://github.com/pascalandy/nicely/issues/381). Checks run through `just`, Lefthook, and `just signoff`, whose `signoff` status the ruleset of `main` requires. GitHub Actions runs `just check` again for each push to `main` and each pull request. That check is advisory: the ruleset does not require it. External pull requests go through the maintainer's agent and `just signoff`. A release runs only when an agent starts the release workflow.

**Why.** Local checks are faster, and `just signoff` proves the exact commit that reaches `main`. A push can skip the pre-push hook of Lefthook, so a second run on a clean runner shows a broken `main`. Nicely is public, so standard GitHub runners run that check for free.

**Rejected.** Local checks alone, because a skipped hook then leaves a broken `main` unseen, as #373 found. GitHub Actions as the only gate, because Pascal often hits its limits, and local checks are faster. A required `check` status, because `signoff` already gates `main`.
