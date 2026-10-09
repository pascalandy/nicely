# Developer preferences

This file records how Pascal wants changes made, where the principles and the specs are silent. Each rule applies to every agent and every milestone.

## Decide the look with a published mockup

Every decision about how Nicely looks, in its TUI or in a GUI, goes through a mockup first. This covers a help page, an error block, a form, a progress display, and any later screen.

1. Build two or three distinct static mockups in one HTML file with the `html-mode` skill, and check it in a browser.
2. Publish it with the `html-publish` skill, and give Pascal the URL that the publication verified.
3. Wait for Pascal's pick before you change `internal/tui` or any other real component. The rest of the task may continue meanwhile.

Keep the mockup and its receipt outside the repository, such as in `~/Documents/artifacts/<name>/`, because the receipt names a private host. To revise a mockup, publish the same file again, so the URL stays the same.

## Wait for Pascal's review before a merge

Open each pull request ready for review, then stop. Merge only after Pascal has reviewed the pull request and says "merge", even when a plan, a milestone, or an earlier answer lets you merge. The project is new, and Pascal prefers to go slower so that he reviews the code.

- When later work depends on an open pull request, stack it on that pull request instead of merging the lower one.
- Fix each review finding on the branch of its pull request, with a scenario that fails first, and reply on the pull request with the commit that fixes it.
