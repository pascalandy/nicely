# Finish gate

This part says what `just finish` checks in an extension repository, as [Extension repository](extension-repository.md) defines it.
`just finish` is the exit of the loop of the `nicely-extension` skill. `just check` stays green for an incomplete round.

`just finish` fails on each of these:

- A missing `vision.md`, `spec.md`, or `qa.md`
- A line of `vision.md`, `spec.md`, or `qa.md` that starts with `Fill:` after its indent, also in fenced code
- A line under `## Open lines`
- A line under **Limits.** that is no table row with a spec ID, the state `limit`, and the user's acceptance
- A `readiness.md` whose first three lines hold no `Commit: <sha>`, no `Reviewer: <who>`, or no verdict `Ready` or `Ready with stated limits`
- A reviewed commit that differs from the working tree outside `readiness.md`
- A finding of `readiness.md` without its own `###` heading under `## Findings`, or without one severity: `block`, `warn`, or `advise`
- A severity under `## Findings` outside the line `- **Severity**:` of a finding, such as in a table, in fenced code, or in prose
- `Ready` with a `block` or `warn` finding, or with a limit in `spec.md`
- `Ready with stated limits` with a `block` finding, or with a `warn` whose **Limit** names no accepted limit
