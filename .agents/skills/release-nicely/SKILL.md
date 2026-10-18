---
name: release-nicely
description: "Use when Pascal asks to release Nicely: check, tag, and publish a version of the nicely repository through release.yml and the Homebrew tap, then verify it."
---

# Release nicely

Release only when Pascal asks. A release is never a card, so it never blocks a milestone. Milestone `M<n>` ships in `v0.<n>.0`, M00 in v0.0.1, and a fix between two milestones takes the next patch number, as [D043](../../../docs/north-star/decisions/D043-milestone-versions.md) decides.

Copy this checklist and tick each step:

```
- [ ] 1. Check: `just release-check vX.Y.Z` passes in a clean checkout of `origin/main`
- [ ] 2. Tag: the tag vX.Y.Z on origin names the SHA of step 1
- [ ] 3. Start: one workflow run of `release.yml` names the tag and that SHA
- [ ] 4. Wait: the workflow run succeeds
- [ ] 5. Verify: `just release-verify vX.Y.Z` passes, and the report gives the tag, the SHA, the workflow run, and each result
```

1. Run `git fetch origin`. Check out `origin/main` in a clean checkout, such as `git worktree add --detach <folder> origin/main`. Make sure that `git status --porcelain` prints nothing there. Keep the SHA that `git rev-parse HEAD` prints. Run `just release-check vX.Y.Z` in that checkout. If it fails, fix the cause on `main`. Then start again at step 1.
2. Tag that SHA: `git tag vX.Y.Z <sha>`. Push the tag: `git push origin vX.Y.Z`. Never move a tag that exists.
3. Start the release once: `gh workflow run release.yml --repo pascalandy/nicely --ref vX.Y.Z -f tag=vX.Y.Z`. The workflow publishes the archives, then updates the formula in the Homebrew tap. Keep the URL of the workflow run that the command prints. Its last number is the `<id>` of the workflow run. If it prints no URL, wait one minute. Then list the workflow runs of the tag: `gh run list --repo pascalandy/nicely --workflow release.yml --event workflow_dispatch --branch vX.Y.Z --json databaseId,createdAt`. Take the one workflow run created after your command. Read its identity: `gh run view <id> --repo pascalandy/nicely --json headBranch,headSha`. `headBranch` must be the tag, and `headSha` must be the SHA of step 1. If no workflow run or several match, or a value differs, stop and report. Never start the release again.
4. Wait for the workflow run: `gh run watch <id> --repo pascalandy/nicely --exit-status`. If it fails, report its evidence, such as `gh run view <id> --repo pascalandy/nicely --log-failed`. Then stop. Never start the release again, and never retry a publication.
5. Run `just release-verify vX.Y.Z`. Report the tag, the SHA, the workflow run, its result, and the results for the assets and the Homebrew tap.

The tag and the `release.yml` that the workflow run uses stay on the checked commit, even when `main` moves on. A new commit needs its own check. GitHub Actions runs only when started this way. After the first release, Pascal checks on his Mac that `brew install pascalandy/tap/ncly` builds `ncly`, and that `ncly <Tab>` completes in zsh.
