# D061 Ship the smallest useful capability before its variants

The scope audit of [PR #156](https://github.com/pascalandy/nicely/pull/156), on 2026-10-17,
reviews every remaining milestone. Several harnesses were the example that exposed the
problem, not a new Codex-only purpose for Nicely. The CLI toolbox remains the North Star.

A first capability serves one concrete job with the fewest necessary inputs, outputs, and
steps. A brainstorm is not a commitment to every variant. Additional providers, input sources,
authentication paths, batches, renderers, distribution methods, or commands need their own
observed use before joining the delivery queue. Shared conventions do not automatically make
every simple command implement them all.

Each milestone delivers its own useful increment. Later planned milestones are candidates,
not acceptance criteria that must all ship before the earlier increment can be used. M99 is
the place for deferred capabilities, not an obligation to build them later. Moving an entire
milestone there removes its file from the delivery queue; leaving it `planned` would still
schedule it. No new planner state or framework is needed.

## Audit coverage

- M00–M08: keep shipped behavior and safeguards. The audit does not reopen completed work.
  Epic 2 ([#150](https://github.com/pascalandy/nicely/issues/150)), M07–M08, shipped in full
  while the audit was open, including its forms, styled reports, and optional keys.
- Epic 3 ([#151](https://github.com/pascalandy/nicely/issues/151)), M09–M10: run and inspect one
  trusted Codex task; defer profiles and multi-adapter machinery. Split durable records from
  execution proof without building partial continuation for a consumer that does not exist.
- Epic 4 ([#152](https://github.com/pascalandy/nicely/issues/152)), M11–M12: move Codex into the
  first agent slice, remove the empty adapter milestone, and keep one official tap installation;
  defer updates, skill linking, and fleet management.
- M13–M16: transcribe one YouTube URL and prove that workflow; defer batches, Zoom, summary,
  and explicit resume. Keep the old CLI until its required uses are actually covered.
- M17–M20: select one daily job before adding a domain. Keep Markdown view as a bounded
  candidate; park the unselected video and image suites and universal batch requirements.
- M21: aggregate explicitly selected local docs with one existing renderer; defer discovery,
  synchronization, visibility management, publishing, and custom search.
- M22: translate core, the SDK, and extensions that actually ship, using existing full catalog
  parity. Do not invent a partial translation mode or translate deferred commands.

[The parking lot](../../milestones/M99-parking-lot.md) records each deferred capability and
its revisit condition. The milestone files remain the canonical execution plan under D042;
old issue descriptions do not reintroduce removed requirements.

## What a scope cut must preserve

Required-key grants and source identity stay in M08. A paid command still needs durable
intent, verified results, inspection, and refusal to repeat an unknown effect. Commands still
honor dry run, truthful exits, cancellation, output validation, and overwrite consent where
they apply. Reusing the existing contract is smaller than inventing an unsafe shortcut.

Rejected: interpreting the audit as "build only Codex", building all brainstormed variants
under one small-looking card, or deleting safety and proven workflows to reduce a line count.
