# Round limit and delivery

This part of the `nicely-extension` skill says how a build reaches "Done when" in `AGENTS.md`, when it stops, and when it delivers.

## Reach "Done when" in 5 rounds

- Round 1 includes a test file for each generic case that applies
- Ask the first reviewer to check that `## Case matrix` is complete and to run it, before it reviews any repair
- Ask the first reviewer to test families of cases, such as all the declarations that change the meaning, not one counterexample
- Fix a root cause, such as a parser or a conversion rule, before the next review
- Keep the last rounds for verification, not for one patch per warning

## Round limit

At the limit of 5 rounds, stop. Do not start round 6.
Report to the user each open line of `spec.md`, and each finding that the ledger records as `open`, with its ID.
Then wait for the user.
An answer that lets you choose, such as "decide yourself", accepts no limit and starts no round.
The user can accept each open finding as a limit. That needs no new round, as `readiness-review.md` says.

## Review, stop, and delivery

These are three different acts:

- A pull request opens for review. It can open in round 1, before the loop is done
- A stop report goes to the user at the round limit. It delivers nothing
- The delivery, such as a merge or a release, starts only after `just finish` passes on the head

A delivery skill or a merge gate, such as a "full stop" or "full send" request, is a delivery.
It never replaces "Done when", and it never ends the rounds early.
The release receipt and the human items of `qa.md` come after the merge, as "Release" in `AGENTS.md` says.
