# D064 Let several extensions share a domain

Decided 2026-10-24 by Pascal.
Several authors can add independent jobs to one domain, such as video conversion and subtitles.
[Extension identity](../contract/extension-identity.md) defines the optional manifest name, the slug, the command root, and each identity use.
Whole-domain extensions keep their existing identity and behavior.

**Why.** A user can install ten video extensions without making one author own every video task.
The community can add niche jobs without changing another author's extension.
Each named extension owns one subtree, so core needs no merge of declarations at individual command paths.

**Revises.** [D003](D003-domain-first.md) retains domain-first commands and adds named extension roots.
[D005](D005-extensions-and-taps.md) and [D048](D048-static-manifest-and-catalogs.md) use the slug for executable discovery and sidecars.
[D044](D044-additive-extensions.md) keeps extensions additive within a shared domain.
[D049](D049-canonical-arguments.md) removes the extension root from canonical arguments.
[D054](D054-extension-skills-through-describe.md) identifies skills by slug and exposes them at extension roots.
[D055](D055-completion-values-from-a-list-command.md) selects a completion provider by its full command path.
[D056](D056-extension-status-command.md) and [D058](D058-grants-by-path.md) identify status checks and grants by slug.
[D062](D062-copier-template.md) and [D063](D063-sdk-test-kit.md) carry that identity through the template and the test kit.

**Rejected.** One extension per domain would force unrelated authors to coordinate every addition.
Tool-named domains and extension names expose implementation choices instead of jobs.
A main extension would give one author special ownership of a shared family.
Deeper identity nesting would complicate discovery without a selected need.

The first consumer is the public repository `pascalandy/ncly-video-convert`, built from the released template after [M23](../../milestones/M23-shared-domains.md).
It wraps HandBrakeCLI as `ncly video convert <file>` and is not a Nicely card.
