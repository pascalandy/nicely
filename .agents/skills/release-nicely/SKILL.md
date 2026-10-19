---
name: release-nicely
description: "Use after Pascal accepts a completed Nicely milestone, or when he requests a release: sign off, merge, generate notes, check, tag, publish, and verify."
---

# Release nicely

Run this cycle after each completed milestone that Pascal accepts, without a separate release request.
For an explicitly requested patch, use the same cycle.
[D043](../../../docs/north-star/decisions/D043-milestone-versions.md) owns version numbering.
A release is never a card and never blocks the next milestone.

Use `git`, `gh`, `just`, Go, and `uv` from the development setup.
Check them with `git --version`, `gh --version`, `just --version`, `go version`, and `uv --version`.
Install missing tools from their official installation guides before continuing.

Copy this checklist and tick each step:

```
- [ ] 1. Sign off: the milestone's pull requests pass the signoff recipe on their final heads
- [ ] 2. Merge: the authorized merge reaches origin/main with the checked commit
- [ ] 3. Changelog: `just changelog vX.Y.Z` writes and reviews this version's notes
- [ ] 4. Check: `just release-check vX.Y.Z` passes in a clean checkout of origin/main
- [ ] 5. Tag: vX.Y.Z on origin names the SHA of step 4
- [ ] 6. Start: one workflow_dispatch run of release.yml names the tag and that SHA
- [ ] 7. Wait: the workflow run succeeds
- [ ] 8. Verify: `just release-verify vX.Y.Z` passes, and the report gives each result
- [ ] 9. Handoff: Pascal receives the release URL and the Mac checks
```

## Sign off and merge

1. Confirm that the milestone is complete and Pascal has accepted it.
2. Rebase its pull requests onto `origin/main` and push their final heads.
3. Run `just signoff`, or `just signoff-layers` for a stack.
4. Merge through [dev-preferences.md](../../../docs/north-star/dev-preferences.md#merge-on-the-brigades-rule-or-pascals-word), preserving the checked commit.
   If merge authorization is missing, report the signed-off pull requests and wait for that authorization.
   An already merged milestone needs no second merge or signoff of its pull requests.
5. Fetch origin and its tags with `git fetch origin --tags`.
6. Use a clean checkout of `origin/main`, such as `git worktree add --detach <folder> origin/main`.
7. Require empty output from `git status --porcelain`.
8. Keep the SHA from `git rev-parse HEAD` and choose the version under D043.
9. Require that the chosen tag and GitHub release do not already exist.
   If either exists, inspect the existing publication and report its state instead of starting another cycle.

## Generate the changelog and check

1. Run `just changelog vX.Y.Z` in that checkout.
2. Review `.release/changelog.md` against the commits since the previous release tag.
   The recipe uses pinned git-cliff through `uv` and [cliff.toml](../../../cliff.toml) to parse Nicely's commit format.
   An explicit previous tag is supported, such as `just changelog v0.16.0 v0.2.1`.
   Notes live in the GitHub release, with this ignored file as their publication input.
   No changelog commit changes the checked SHA.
3. Run `just release-check vX.Y.Z` in the same checkout.
   If it fails, fix the cause through a pull request and repeat signoff, authorized merge, changelog, and check.
4. Require empty output from `git status --porcelain` and the same SHA from `git rev-parse HEAD`.

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
3. Ask Pascal to check `brew install pascalandy/tap/ncly` and `ncly <Tab>` in zsh on his Mac.
