# D069 Record the sessions of authors and reviewers

Chosen for [#450](https://github.com/pascalandy/nicely/issues/450), after Pascal chose recorded session IDs on 2026-11-03.
This decision specifies the recording design. It does not add production enforcement.
Today, `just finish` still accepts the reviewer name of its existing record.
The follow-up #451 implements the rule after refinement; it is outside this change.

## 1. Capture identity from the harness

| Harness | Supported floor for this design | Identity source |
|---|---|---|
| Claude Code | 2.1.317 | `SessionStart` command-hook input, field `session_id` |
| Codex | 0.166.0 | `SessionStart` command-hook input, field `session_id` |

These are chosen, tested support floors, not the first versions that introduced hooks.
Claude introduced `SessionStart` in 1.0.62, but that fact alone does not establish every field of the chosen capture contract.
The earliest complete Codex version was not established.
The recorder checks the installed version and requires an actual capture; configuration alone proves nothing.

A launcher creates one context file per invocation before starting the harness.
The hook writes the emitted ID, harness version, event, source, role and repository into that context.
The owned recording recipes accept no session ID argument, and no prompt asks an agent to provide its identity.
A launcher-generated UUID is allowed for Claude's `--session-id` only when the emitted hook confirms that exact value.
A transcript filename, display name, tool-call ID or Codex child thread ID is never a substitute.

On resume, clear, compact or fork, the hook records the new event and emitted identity.
Each commit reads the current context of its invocation; it never reads a global "current session" file.
Missing or stale capture stops recording with a setup instruction.

## 2. Install project hooks explicitly

The template supplies project hook declarations and launcher recipes.
Their recording logic lives in the shared SDK repository program, following D067; the template copies only calls and forms.
An extension adopts these files through its template update.
Nicely does not silently edit a user's global Claude or Codex settings.

The user owns project and hook trust.
Codex requires its project configuration and non-managed hooks to be trusted before they run.
The ordinary setup uses the supported trust interface, rather than editing trust storage.
The scratch proof vets its hook files and uses Codex's per-invocation hook-trust bypass.
That exception installs no persisted trust and is not the default author path.

The launcher owns the context path, role and expected repository, and exports them to the harness and its Git subprocesses.
A missing context, wrong repository or unobserved hook fails; it never falls back to an agent-written ID.

## 3. Bind each author to the actual commit

The canonical author receipt maps an actual full commit SHA to its captured author sessions.
Store receipts as Git notes under one explicit `refs/notes/nicely-sessions` ref, outside source commit history.
This avoids a receipt commit changing the commit that its own receipt names.
Each receipt includes its capture source and harness version.
Readable commit trailers may repeat the captured identity, but copied trailers alone do not establish a fresh binding.

The owned commit path reads the invocation context, creates the commit, resolves its actual SHA, then writes its receipt.
A launcher may finalize those receipts after the session returns only with exclusive ownership of its worktree and a recorded before/after range.
It must account for every commit in that range and reject unexpected writers or head movement.
The scratch proof uses Claude's Stop hook to check, commit and record the resulting SHA during the author session.
Its Git message hook adds the captured ID automatically; the launcher exports the exact-SHA receipt from those captures.

For several authors, the round's author set is the union of its per-commit receipts.
A history rewrite creates new SHAs and invalidates the old bindings and review.
An owned rewrite must capture its writer and preserve any recorded source authors in fresh receipts.
It cannot infer missing historical authors from names or copied text.
An unsupported rewrite fails until a new captured round or supported rewrite records those bindings.

## 4. Bind the reviewer to an immutable range

The reviewer launcher resolves the round's full base and head SHAs before starting a fresh root session.
It passes that exact range as the review target and captures the reviewer's hook identity in a new invocation context.
It does not resume the author, fork the author's conversation or ask a native child to satisfy the different-session rule.

After the review completes successfully, the launcher checks the same head and a clean source tree again.
It writes a review receipt attached to that exact head, preserving the head's author receipt.
The receipt names the base, head, captured reviewer session, harness version, root launch and completed result.
A failed, interrupted or missing capture yields no completed receipt.
A completed review with findings still follows the existing finding and resolution gates.

Publish the explicit notes ref with the source branches, and export a portable receipt bundle for checks.
The publication recipe must verify both writes; ordinary branch push does not transport Git notes.
A review or author receipt changes neither the source HEAD nor the review range.
Source edits, amendments and rebases need fresh bindings and another review.

## 5. Keep pre-adoption rounds as history

An extension records its adoption commit as the boundary.
Finished rounds before that boundary remain legacy; no tool invents IDs for them or silently upgrades their claims.
The first enforced round starts from that boundary and records every new commit and its review.
An open round that crosses adoption starts a new captured round, rather than using a historical reviewer name as an ID.
Legacy ancestors stay outside the enforced range.

## 6. Check a detached signoff without an active session

A detached `just signoff` fetches the explicit notes ref and freezes its object ID before checking.
It reads the build's adoption boundary and round range, and requires:

- The source HEAD equals the completed review's full head SHA and the base equals the round's base
- Every commit of the enforced range has a complete exact-SHA author receipt from a supported capture
- The reviewer receipt comes from a completed fresh root launch, with no missing capture or unresolved required finding
- The reviewer ID differs from every recorded author ID of the round
- The source HEAD and frozen receipt set remain unchanged through checks and status publication

The check needs no active harness, environment session ID or ID argument.
Missing notes, a failed fetch, an incomplete export, a changed head or a missing author fails with a recording or re-review instruction.
A portable export carries the source head and notes-ref object ID, so detached checks identify the receipt set they checked.
Git notes are editable metadata, so the signoff reports that frozen receipt set rather than treating a moving ref as immutable.

## 7. Reject a native child as a separate author review

Codex's subagent hook reports its parent's `session_id`.
Its `agent_id` identifies the child, not a new recorded session for this rule.
A child of an author therefore fails reviewer inequality even when it has a different thread ID.
The skill must start a separate root reviewer session through the launcher and compare the captured IDs.
A fresh process alone does not prove inequality; the emitted values decide it.

The scratch proof also asks a native child of the separate reviewer to report the parent ID supplied by its hook.
That third observation is expected to equal the reviewer ID.
It shows what a child exposes; it does not act as the round's reviewer.

## 8. Claim a different recorded session

The receipt proves a different recorded session under this workflow.
It does not prove independent reasoning, a different human, a different model or an absence of shared context.
Two sessions may share all those things and reach the same conclusion.
An editable hook, Git note or caller environment is not a cryptographic boundary against a hostile author.
The reviewer still judges the work; the receipt makes the identity and target claims explicit and checkable.

## Sources and scratch proof

The [OpenAI hooks contract](https://learn.chatgpt.com/docs/hooks#common-input-fields) defines the session and child fields and hook trust.
The [Claude hook contract](https://code.claude.com/docs/en/hooks#sessionstart-input) defines its emitted ID.
The [Claude changelog](https://code.claude.com/docs/en/changelog#1062) dates the earlier event introduction.
The [proof record](D069/proof.md) preserves the emitted identities and their Git binding, with private authentication folders excluded.
The prototype and negative controls exercise this design in a disposable extension; they do not make today's `just finish` enforce it.
