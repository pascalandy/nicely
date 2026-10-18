# Developer preferences

This file records how Pascal wants changes made, where the principles and the specs are silent. Each rule applies to every agent and every milestone.

## Decide the look with a published mockup

Every decision about how Nicely looks, in its TUI or in a GUI, goes through a mockup first. This covers a help page, an error block, a form, a progress display, and any later screen.

1. Build two or three distinct static mockups in one HTML file with the `html-mode` skill, and check it in a browser.
2. Publish it with the `html-publish` skill, and give Pascal the URL that the publication verified.
3. Wait for Pascal's pick before you change `sdk/tui` or any other real component. The rest of the task may continue meanwhile.

Keep the mockup and its receipt outside the repository, such as in `~/Documents/artifacts/<name>/`, because the receipt names a private host. To revise a mockup, publish the same file again, so the URL stays the same.

## Keep the docs a tower

The docs form layers like the code: the glossary fixes the words, the vision and the principles say what and why, the architecture says where, the specs say exactly how, and the milestones say when. [Find what you need](../../AGENTS.md#find-what-you-need) maps each question to its file, and [D025](decisions/D025-docs-layout.md) explains why. Keep the tower whenever you change a doc:

1. Each file answers one question, which its opening lines state.
2. A meaning lives in one place. Every other file links to it instead of restating it.
3. A file read every session holds at most 150 lines: AGENTS.md, vision.md, principles.md, architecture.md, and this file.
4. One word names one concept in code, docs, help, and commit messages. A new term enters glossary.md before its first use, with its relations and the words it replaces.
5. A principle links the decisions behind it. Each decision is its own file in decisions/, with a row in its index.
6. A diagram is Mermaid source inside the doc that it explains, so GitHub, Obsidian, and agents all read it.
7. `just test` checks every link and heading, the decision index, and the length of each file read every session. It also checks that each card ID names a card of a milestone. Fix a problem in the commit that causes it.

## Write clear technical English

Use these writing practices from ASD-STE100 Simplified Technical English, Issue 9 (January 15, 2025), for English prose. In procedures, give one instruction per sentence and use the imperative form. Limit procedural sentences to 20 words. In descriptive text, give each sentence one topic and limit it to 25 words. Keep command syntax and code identifiers exact.

## Run macOS checks on the mini

The `fleet` skill names Pascal's machines. The mini is the agents' Mac and stays on. The mbp is Pascal's laptop, and it is often offline.

1. Run each check that needs macOS on the mini, through the `fleet` skill.
2. If the mini lacks Nicely or a needed tool, or is offline, use the mbp.
3. In the report, name the machine. When you used the mbp, give the reason.

## Merge on the brigade's rule or Pascal's word

Open each pull request ready for review. In a `brigade-mode` run, the brigade decides when to merge. Otherwise, merge only when Pascal says "merge" for that pull request.

- When later work depends on an open pull request, stack it on that pull request.
- Fix each review finding on the branch of its pull request, with a scenario that fails first. Reply on the pull request with the commit that fixes it.

To merge a stack, ask for no confirmation, as for one branch. Stop only on a real blocker. Follow Land in the `gh-stack` skill, for a project that pushes a head to its trunk. Add these checks to its steps:

1. After `gh stack unstack`, the `gh api` stack check of the `gh-stack` skill must print nothing.
2. Push each rebased branch with `git push --force-with-lease`, as for one branch.
3. Check that each branch, its head on origin, and the `headRefOid` of its pull request name one commit.
4. Run `just signoff-layers` with the branches, bottom to top. It also fails a layer that does not contain the layer below. A new head needs a new run.
5. After the retarget, run `gh pr view <number> --json baseRefName,closingIssuesReferences` on each pull request. It must show `main` and each Story that its closing keywords name.
6. Push the last SHA in the `summary.txt` of that run: `git push origin <top-sha>:main`.
7. Check that `gh pr view <number> --json state` prints `MERGED` for each pull request.
8. Check that `gh issue view <story> --json state,closedByPullRequestsReferences` shows each Story closed by its pull request.

A failed step leaves the merge incomplete, so report it. GitHub can refuse a retarget, or leave a pull request open. Never close an open pull request by hand to report it merged.

To merge, ask for no confirmation. Stop only on a real blocker, and report it. Rebase the branch onto `origin/main`. Run `git push --force-with-lease`. Run `just signoff`. Then run `git push origin <branch>:main`, so `main` keeps the commit that was checked.
