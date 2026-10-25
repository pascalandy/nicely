# Developer preferences

Pascal's preferences apply to every agent and backlog item where principles and specs are silent.

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
- Keep each file within 2,500 tokens, its bytes divided by 4, and AGENTS.md within 1,250. `just limits` checks each commit and names skipped files in its [rules](../../tools/cmd/plan/limits.go)
- Use one term per concept across code, docs, help, and commits. Add new terms, relations, and replaced words to glossary.md before use
- Link principles to decisions, with one file and index row per decision
- Keep diagrams as Mermaid source in the doc they explain

Fix documentation errors in the commit that introduces them. `just test` checks links, headings, decision indexes, card references, and explicit code roles. It rejects em dashes and curly quotes in prose.

## Write clear technical English

Use ASD-STE100 Issue 9 writing practices for English prose. Procedures use imperative sentences, one instruction each, at most 20 words. Descriptive sentences have one topic and at most 25 words. Preserve exact command syntax and identifiers.

`just limits` refuses a new or changed sentence over 25 words in Markdown or in a full-line Go comment. It reads the `.jinja` sources of the extension template the same way.

## Follow serious CLIs where the specs are silent

Where the specs are silent on CLI design, follow what serious CLIs agree on, such as `gh`, `kubectl`, Terraform, Docker, and cargo, and the guidelines of clig.dev. The `coding-standard` skill condenses clig.dev. When the choice becomes a rule, record the comparison in a new decision in [decisions/](decisions/README.md).

## Use uv for Python

Always use `uv` for Python execution, environments, and dependencies.
Keep the tool and package pins in `just python` and the script's dependency metadata.
Select the Python version from `extensions/transcript/python/transcript.py`.
Run checks that spawn Python inside that uv-selected environment.
For the current minimum, run `uv run --no-project --python 3.12 just check`.
Use the same prefix for `just test`, `just scenario`, and `just signoff`.
Do not install system Python or change shell profiles for these checks.

## Run macOS checks on the mini

Use the `fleet` skill to run macOS checks on the mini. Use the mbp only if the mini is offline or lacks Nicely or a required tool. Report the machine and the reason for any mbp fallback.

## Merge after an independent review

Open review-ready pull requests. In `brigade-mode`, the brigade decides when to merge. Otherwise, have an agent that did not write the change review each pull request, and quote its verdict in the pull request. Merge it yourself when `just signoff` passes, no review finding stays open, and you are very confident. Without that confidence, wait until Pascal says "merge" for that pull request. Report each merge to Pascal. Once authorized, ask for no further confirmation. Stop and report only a real blocker.

Stack work on an open pull request when it depends on that pull request. [Make an item ready](../backlog/README.md#make-an-item-ready) owns readiness. Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests.

Fix review findings on their pull request's branch. A finding about behavior gets a scenario that fails first. Reply with the fixing commit.

For one branch, rebase onto `origin/main`, push with `--no-verify --force-with-lease`, and run `just signoff`. Then push the SHA of its `Signed off on <sha>` line with `git push --no-verify origin <sha>:main`.

These pushes skip the pre-push `just test`, because `just signoff` checks that commit. `main` receives only that signed SHA, even when the branch moves later. An admin push bypasses the `signoff` ruleset, so never push the branch name to `main`. Every other push keeps the hook.

For a stack, follow Land in `gh-stack` for a project that pushes its head to trunk, with these checks. Steps 2 and 7 skip the hook too, because `just signoff-layers` checks each layer:

1. After `gh stack unstack`, require an empty result from the skill's `gh api` stack check.
2. Push each rebased branch with `git push --no-verify --force-with-lease`.
3. Require identical local, origin, and pull-request `headRefOid` SHAs for each branch.
4. Run `just signoff-layers` bottom to top. Each layer must contain the one below. Repeat for each new head.
5. After retargeting, check each pull request with `gh pr view <number> --json baseRefName,closingIssuesReferences`.
6. Require `main` as base and every Story named by closing keywords.
7. Push the top SHA that the final `summary.txt` marks `ok` with `git push --no-verify origin <top-sha>:main`.
8. Require `MERGED` from `gh pr view <number> --json state` for every pull request.
9. Require each Story closed by its pull request in `gh issue view <story> --json state,closedByPullRequestsReferences`.

Report any failed step as an incomplete merge. Never manually close an open pull request to report it merged.

## Keep UAT in separate issues

Create a separate GitHub issue for each UAT. Keep its task and procedure there, with only a link in the Epic.

## Release each accepted item

After each completed item that Pascal accepts, run the release cycle without a separate request.
Use [D065](decisions/D065-backlog-items.md) for acceptance and numbering, and [release-nicely](../../.agents/skills/release-nicely/SKILL.md) for the full procedure.
Item acceptance does not replace the merge authorization above.
