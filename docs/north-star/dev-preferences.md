# Developer preferences

This file records how Pascal wants changes made, where the principles and the specs are silent. Each rule applies to every agent and every milestone.

## Decide the look with a published mockup

Every decision about how Nicely looks, in its TUI or in a GUI, goes through a mockup first. This covers a help page, an error block, a form, a progress display, and any later screen.

1. Build two or three distinct static mockups in one HTML file with the `html-mode` skill, and check it in a browser.
2. Publish it with the `html-publish` skill, and give Pascal the URL that the publication verified.
3. Wait for Pascal's pick before you change `internal/tui` or any other real component. The rest of the task may continue meanwhile.

Keep the mockup and its receipt outside the repository, such as in `~/Documents/artifacts/<name>/`, because the receipt names a private host. To revise a mockup, publish the same file again, so the URL stays the same.

## Keep the docs a tower

The docs stack like the code: the glossary fixes the words, the vision and the principles say what and why, the architecture says where, the specs say exactly how, and the milestones say when. [Find what you need](../../AGENTS.md#find-what-you-need) maps each question to its file, and [D025](decisions/D025-docs-layout.md) explains why. Keep the tower whenever you change a doc:

1. Each file answers one question, which its opening lines state.
2. A meaning lives in one place. Every other file links to it instead of restating it.
3. A file read every session stays under 150 lines: AGENTS.md, vision.md, principles.md, architecture.md, and this file.
4. One word names one concept in code, docs, help, and commit messages. A new term enters glossary.md before its first use, with its relations and the words it replaces.
5. A principle links the decisions behind it. Each decision is its own file in decisions/, with a row in its index.
6. A diagram is Mermaid source inside the doc that it explains, so GitHub, Obsidian, and agents all read it.
7. `just test` checks every link and heading. Fix a link in the commit that breaks it.

## Wait for Pascal's review before a merge

Open each pull request ready for review, then stop. Merge only after Pascal has reviewed the pull request and says "merge", even when a plan, a milestone, or an earlier answer lets you merge. The project is new, and Pascal prefers to go slower so that he reviews the code.

- When later work depends on an open pull request, stack it on that pull request instead of merging the lower one.
- Fix each review finding on the branch of its pull request, with a scenario that fails first, and reply on the pull request with the commit that fixes it.
