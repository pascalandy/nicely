# AGENTS.md

This file tells an agent how to change Nicely. [README.md](README.md) describes Nicely for its users.

## Read in this order

1. [docs/north-star/guide.md](docs/north-star/guide.md), at the start of every session.
2. The current milestone in `docs/milestones/`: the lowest-numbered file whose status is not `done`.
3. The sections of [docs/north-star/cli-spec.md](docs/north-star/cli-spec.md) that your task touches.
4. [docs/north-star/decisions.md](docs/north-star/decisions.md), only when a rule blocks your task.

## Work on a milestone task

Copy this checklist and tick each step when its condition holds.

```
- [ ] 1. Spec: cli-spec.md covers every command, flag, JSON key, and error code you add or change
- [ ] 2. Scenario: a testscript scenario fails for the missing behavior
- [ ] 3. Code: the scenario passes
- [ ] 4. Text: each new user-facing string is a catalog entry with a description
- [ ] 5. Checks: `just check` passes
- [ ] 6. Docs: the milestone boxes you finished are ticked, and README.md shows any change a user sees
- [ ] 7. Sign off: `just signoff` passes
```

1. **Spec first.** Write the section in cli-spec.md before code, tagged with the milestone. Done when a reader can call the command from the spec alone.
2. **Scenario before code.** Scenarios live in `testdata/script/` as `.txtar` files. Cover the JSON success path, one failure with its error code and exit code, and `--dry-run` for a command that writes. Put stub executables in `PATH` for harnesses and paid services, and use go-keyring's mock for the keychain. Done when the new scenario fails for the expected reason.
3. **Code.** Follow the layout in [M0](docs/milestones/M0-foundation.md). Build interactive screens from `internal/tui` only, so every command looks the same.
4. **Text.** Give each catalog entry an ID and a description of where the text appears, so a translator picks the right sense.
5. **Checks.** When `just check` fails, fix the cause and rerun. After three failed attempts on the same check, stop and report what you tried.
6. **Docs.** Tick the boxes of the milestone's "Done when" list that your work completes, and update README.md when a user would see the change.
7. **Sign off.** `just signoff` reruns the checks, then calls `gh signoff`. Always sign off through the recipe, because a bare `gh signoff` posts a green status without running anything.

## Rules

- Where cli-spec.md is silent on CLI design, load the `coding-standard` skill if your harness has it. Otherwise follow clig.dev.
- Use the terms defined in the guide. Add a new term to the guide before using it.
- Prefer end-to-end testscript scenarios. Write a unit test only for logic that a scenario cannot reach.
- Write code, comments, docs, and commit messages in English.
- Keep Python under `python/`. A Python component returns JSON to Go, and Go owns display, translation, and exit codes.
- Stay inside the current milestone. Write any other idea as one line in [M99](docs/milestones/M99-parking-lot.md), then continue.
- To change a principle or the agent contract, add an entry to decisions.md and ask Pascal before merging.
- Use placeholders such as `/Users/me`, `host-a`, and `example.com` in code, tests, and docs. The pre-commit hook runs gitleaks with rules for personal paths, hostnames, and local IPs.
- When `ncly` deletes a user's file, it moves the file to the trash through `internal/platform`.

## Release

Release only when Pascal asks.

1. Run `just release-check vX.Y.Z` and fix anything it reports.
2. Tag and push: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. Start the release: `gh workflow run release.yml -f tag=vX.Y.Z`.

GitHub Actions runs only when started this way.

## Review an external pull request

1. Run `gh pr checkout <number>`.
2. Compare the diff with the current milestone and cli-spec.md.
3. Run `just signoff`. The signoff status is a required check on `main`, so the pull request merges only after this step.
