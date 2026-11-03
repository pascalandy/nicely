# Update and fixes of Nicely

This part of the `nicely-extension` skill says how an extension starts its first build after `just update`.
It also says what to do with a failure that only Nicely can fix.

## First build after an update

An extension from a template before `just finish` adopts the loop through `just update`, as `update-an-extension.md` in the contract says (#427).
Its first build starts after the update commit:

1. Resolve each conflict of the update. Keep each new `Fill:` line and each new part, such as `## Case matrix`. The first build fills them.
2. Commit the update on its own. It adds `intake.toml`, with the example facts of the template.
3. Do round 1 of `SKILL.md` for the commands that exist. In step 1, read the current `spec.md`, and the wrapped programs at the revisions that `vision.md` names. Replace the example facts.
4. In step 2, keep each line that a test claims. Write each new line, such as a generic case that the older template did not list. A line that no test proves goes under `## Open lines`.
5. In step 3, the first scenario claims one of these open lines, and it fails first.

`just finish` refuses an Intake commit that is not after the commit that added `intake.toml`.
A Markdown `readiness.md` of an older template stays in the Git history. Finish reads none of it.

## A fix that belongs to Nicely

Some failures come from the SDK or the template, not from the extension.
For example, `sdk/wrap` writes the timeout answer, so an extension could not add the input file to it (#407).
When only Nicely can fix a failure, do the step of [D068](https://github.com/pascalandy/nicely/blob/main/docs/north-star/decisions/D068-verification-map.md#move-a-correction):

1. Open an issue in pascalandy/nicely with the input, the host, and the failure.
2. Record a limit under **Limits.** of `spec.md` that links the issue. The user accepts it like any other limit, as `readiness-review.md` says.
3. When a Nicely release carries the fix, run `just update` to it. Then replace the limit with a line of Failures or Acceptance, and its scenario, which proves it.

Only you can judge that a fix belongs to Nicely. No check decides it.
