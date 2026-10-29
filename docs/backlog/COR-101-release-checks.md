---
id: COR-101
core: "release checks"
state: "Active. COR-101-T2 waits for Pascal, because each run is a paid Codex task. T1 added the free checks of #375"
---

# COR-101 release checks

## Demo

The selected release checks prove an installed extension and the release paths that the reality check found untested.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-101-T1 | Free release checks | agent | — | done |
| COR-101-T2 | Live Codex test | Pascal | — | todo |

### COR-101-T1 Free release checks

- **Read:** [release-nicely](../../.agents/skills/release-nicely/SKILL.md), [Records and replay](../../extensions/agent/spec.md#records-and-replay), [Extension repository](../north-star/contract/extension-repository.md)
- **Proves:** `extensions/agent/testdata/script/agent_run_busy.txtar`, `.agents/skills/release-nicely/scripts/extension-qa.sh`, `.agents/skills/release-nicely/SKILL.md`, `scripts/template-update-test.sh`

The reality check of 2026-10-24 found four release paths that no check covered:

- A release tests ncly alone. It now installs video-convert from its latest tag on macOS, then runs describe, doctor, a dry run, and one conversion
- No release step required `just check` on macOS. The release now runs `just check --keep-going` on macOS before its tag
- No test ran a dry run while a run held the lock. Both the dry run and the real call now answer `TEMPORARY` with `ncly run view <id>` as `next`
- A dependency that an extension added to `go.mod` conflicted at each template update, as U1 of [#347](https://github.com/pascalandy/nicely/issues/347) found. `just update` now keeps it

### COR-101-T2 Live Codex test

Real Codex passed the live test of [M10](../milestones/M10-agent-run.md#after-this-milestone) once, on 2026-10-17 with Codex 0.163.0.
Rerun it after a Codex upgrade.
Each run is one small paid task that Pascal authorizes.
