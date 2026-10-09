# D016 Run CI locally and release by hand

Decided 2026-10-05. Checks run through `just`, Lefthook, and `gh signoff`. GitHub Actions runs only when an agent starts the release workflow. External pull requests go through the maintainer's agent and `just signoff`.

**Why.** Pascal often hits GitHub Actions limits, and local checks are faster.
