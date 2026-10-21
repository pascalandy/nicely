---
name: release-nicely
description: "Use after Pascal accepts a completed Nicely milestone, or when he requests a release: generate notes, check, tag, publish, and verify."
---

# Release nicely

Run this cycle after each completed milestone that Pascal accepts, without a separate release request.
For an explicitly requested patch, use the same cycle.
[D043](../../../docs/north-star/decisions/D043-milestone-versions.md) owns acceptance and version numbering.
A release is never a card and never blocks the next milestone.

Copy this checklist and tick each step:

```
- [ ] 1. Ready: the milestone's cards are merged on origin/main, and Pascal accepted it as D043 says
- [ ] 2. Changelog: `just changelog vX.Y.Z` writes this version's notes, reviewed against its commits
- [ ] 3. Check: `just release-check vX.Y.Z` passes in a clean checkout of origin/main
- [ ] 4. Tag: vX.Y.Z on origin names the SHA of step 3
- [ ] 5. Start: one workflow_dispatch run of release.yml names the tag and that SHA
- [ ] 6. Wait: the workflow run succeeds
- [ ] 7. Verify: `just release-verify vX.Y.Z` passes, and the report gives each result
- [ ] 8. Mac: `brew-qa.sh vX.Y.Z` passes on the mini, or on Pascal's Mac when the mini is offline
```

## Generate the changelog and check

1. Fetch origin and its tags with `git fetch origin --tags`.
2. Use a clean checkout of `origin/main`, such as `git worktree add --detach <folder> origin/main`.
3. Require empty output from `git status --porcelain`.
4. Keep the SHA from `git rev-parse HEAD` and choose the version under D043.
5. Require that the chosen tag and GitHub release do not already exist.
   If either exists, inspect the existing publication and report its state instead of starting another cycle.
6. Run `just changelog vX.Y.Z` in that checkout.
7. Review `.release/changelog.md` against the commits since the previous release tag.
   The recipe uses pinned git-cliff through `uv` and [cliff.toml](../../../cliff.toml) to parse Nicely's commit format.
   An explicit previous tag is supported, such as `just changelog v0.16.0 v0.2.1`.
   Pass the last tag that has a GitHub release when a newer tag never got one.
   Notes live in the GitHub release, with this ignored file as their publication input.
   No changelog commit changes the checked SHA.
8. Run `just release-check vX.Y.Z` in the same checkout.
   If it fails, fix the cause through a merged pull request and restart the cycle.
9. Require empty output from `git status --porcelain` and the same SHA from `git rev-parse HEAD`.

## Tag and publish once

1. Tag the checked SHA with `git tag vX.Y.Z <sha>`.
2. Push it with `git push origin vX.Y.Z`.
   Keep existing tags unchanged.
3. Start publication once with `gh workflow run release.yml --repo pascalandy/nicely --ref vX.Y.Z -f tag=vX.Y.Z`.
   The workflow regenerates the same notes from the tag, publishes the versioned archives, and updates the Homebrew formula.
   GoReleaser embeds the version from the tag, so no source version bump is needed.
4. Keep the workflow run URL and its final number as `<id>`.
   If the command prints no URL, wait one minute.
   Then run `gh run list --repo pascalandy/nicely --workflow release.yml --event workflow_dispatch --branch vX.Y.Z --json databaseId,createdAt`.
   Select the one run created after the dispatch.
5. Read its identity with `gh run view <id> --repo pascalandy/nicely --json headBranch,headSha`.
   Require the chosen tag in `headBranch` and the checked SHA in `headSha`.
   If no run or several match, or either value differs, stop and report.
   Inspect existing runs after a lost dispatch response. Never dispatch again or retry a publication.
6. Wait with `gh run watch <id> --repo pascalandy/nicely --exit-status`.
   If it fails, report `gh run view <id> --repo pascalandy/nicely --log-failed` and stop.

The tag and workflow stay on the checked commit even when `main` moves on.
GitHub Actions runs only for this manual publication.

## Verify and hand off

1. Run `just release-verify vX.Y.Z`.
2. Report the release URL, tag, SHA, workflow run, and results for the assets and Homebrew tap.
3. On the mini, run `bash .agents/skills/release-nicely/scripts/brew-qa.sh vX.Y.Z` in a clone of origin/main, through the `fleet` skill.
   It installs or upgrades `ncly` from the tap, checks its version, and types `ncly <Tab>` in a clean zsh.
   `ncly` stays installed on the mini, so the next release tests the upgrade.
   If the mini is offline, ask Pascal to run it on his Mac, because it changes the Homebrew of that Mac.
   If it fails, report its log and stop.
4. Report its result with the others.
