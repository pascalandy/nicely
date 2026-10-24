# D042 Plan work as milestones of small cards

Decided 2026-10-09. A milestone demonstrates one capability, with at most five agent cards. A card is one pull request. [The milestones' README](../../milestones/README.md) owns readiness, ownership, dependencies, and format. `just next` selects the work; `just status` reports progress.

Milestones follow their dependencies, keep one extension's work together, and put Pascal's tools before everyday domains.
Only milestones without done cards may be renumbered, in a plan-only pull request.

Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests.

**Why.** The former mix of tasks and acceptance lists let agents finish isolated tasks while leaving progress unclear. Each card now has one owner and proof.

**Rejected.** Unchecked conventions and GitHub as the queue. The plan travels with the code.
