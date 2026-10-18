# Developer preferences

Pascal's preferences apply to every agent and milestone where principles and specs are silent.

## Decide the look with a published mockup

Every TUI or GUI appearance decision starts with a mockup, including help, errors, forms, progress, and later screens.

1. Build two or three distinct static options in one HTML file with `html-mode`.
2. Check the file in a browser.
3. Publish with `html-publish` and give Pascal the verified URL.
4. Wait for his pick before changing real components. Other work may continue.

Keep mockups and receipts outside the repository because receipts name private hosts. Republish the same file for revisions to preserve its URL.

## Keep the docs a tower

Follow the owners in [Find what you need](../../AGENTS.md#find-what-you-need). [D025](decisions/D025-docs-layout.md) records the rationale.

- State each file's one question in its opening
- Keep each meaning in one place and link from elsewhere
- Limit each session-read file to 150 lines: AGENTS.md, vision.md, principles.md, architecture.md, and this file. Limit AGENTS.md to 1250 tokens, its bytes divided by 4, which `just agents-size` checks before each commit
- Use one term per concept across code, docs, help, and commits. Add new terms, relations, and replaced words to glossary.md before use
- Link principles to decisions, with one file and index row per decision
- Keep diagrams as Mermaid source in the doc they explain

Fix documentation errors in the commit that introduces them. `just test` checks links, headings, decision indexes, session-file lengths, card references, and explicit code roles. It rejects em dashes and curly quotes in prose.

## Write clear technical English

Use ASD-STE100 Issue 9 writing practices for English prose. Procedures use imperative sentences, one instruction each, at most 20 words. Descriptive sentences have one topic and at most 25 words. Preserve exact command syntax and identifiers.

## Run macOS checks on the mini

Use the `fleet` skill to run macOS checks on the mini. Use the mbp only if the mini is offline or lacks Nicely or a required tool. Report the machine and the reason for any mbp fallback.

## Merge on the brigade's rule or Pascal's word

Open review-ready pull requests. In `brigade-mode`, the brigade decides when to merge. Otherwise, merge only when Pascal says "merge" for that pull request. Once authorized, ask for no further confirmation. Stop and report only a real blocker.

Stack work on an open pull request when it depends on that pull request. [Make a milestone ready](../milestones/README.md#make-a-milestone-ready) owns readiness. Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests.

Fix review findings on their pull request's branch, with a scenario that fails first. Reply with the fixing commit.

For one branch, rebase onto `origin/main`, push with `--force-with-lease`, and run `just signoff`. Then run `git push origin <branch>:main` to preserve the checked commit.

For a stack, follow Land in `gh-stack` for a project that pushes its head to trunk, with these checks:

1. After `gh stack unstack`, require an empty result from the skill's `gh api` stack check.
2. Push each rebased branch with `git push --force-with-lease`.
3. Require identical local, origin, and pull-request `headRefOid` SHAs for each branch.
4. Run `just signoff-layers` bottom to top. Each layer must contain the one below. Repeat for each new head.
5. After retargeting, check each pull request with `gh pr view <number> --json baseRefName,closingIssuesReferences`.
6. Require `main` as base and every Story named by closing keywords.
7. Push the run's final `summary.txt` SHA with `git push origin <top-sha>:main`.
8. Require `MERGED` from `gh pr view <number> --json state` for every pull request.
9. Require each Story closed by its pull request in `gh issue view <story> --json state,closedByPullRequestsReferences`.

Report any failed step as an incomplete merge. Never manually close an open pull request to report it merged.

## Keep UAT in separate issues

Create a separate GitHub issue for each UAT. Keep its task and procedure there, with only a link in the Epic.
