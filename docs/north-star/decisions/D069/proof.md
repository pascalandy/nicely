# D069 scratch recording proof

Recorded on 2026-11-03 with Claude Code 2.1.317 and Codex 0.166.0.
The disposable extension and harness state lived under the Cook work folder.
Private authentication folders were excluded from every evidence copy and quote.
Their auth symlinks were removed by link path after the harnesses stopped; both original auth files still exist.

Claude edited the greeting implementation. Its Stop hook ran the real-host check and committed the edit.
The Git hook obtained the author trailer from SessionStart capture, without an agent-supplied ID.
A fresh Codex root reviewed the fixed base/head range, passed the same check and left the source tree unchanged.
Its native child reported the parent ID from the SubagentStart capture.
The child observation equals the reviewer ID; it is not a third distinct session.

## Claude author

```json
{
  "event": "SessionStart",
  "session_id": "b7897dc7-005f-4d0b-8a9e-ba821f3952d4",
  "cwd": "/Users/me/session-proof/extension",
  "source": "startup",
  "agent_id": null,
  "harness": "claude-code",
  "version": "2.1.317",
  "role": "author"
}
```

## Codex reviewer

```json
{
  "event": "SessionStart",
  "session_id": "01a1a354-e946-7f63-b6a4-a832a1578c3f",
  "cwd": "/Users/me/session-proof/extension",
  "source": "startup",
  "agent_id": null,
  "harness": "codex",
  "version": "0.166.0",
  "role": "reviewer"
}
```

## Parent ID reported by the native child

```json
{
  "event": "SubagentStart",
  "session_id": "01a1a354-e946-7f63-b6a4-a832a1578c3f",
  "cwd": "/Users/me/session-proof/extension",
  "source": null,
  "agent_id": "01a1a355-1252-7cc1-b136-503e9eba26c1",
  "harness": "codex",
  "version": "0.166.0",
  "role": "reviewer"
}
```

## Git binding

The recorded output of `git log -1 --format=fuller --notes=nicely-session-proof` includes both receipts:

```text
commit 0f3f093575f806b5d1da5d07e23551b82ef524dc
Author:     Nicely session proof <pascalandy@users.noreply.github.com>
AuthorDate: Tue Nov 3 16:29:28 2026 -0500
Commit:     Nicely session proof <pascalandy@users.noreply.github.com>
CommitDate: Tue Nov 3 16:29:28 2026 -0500

    🚑 fix: sessionproof: Greet the requested person

    Nicely-Session-ID: b7897dc7-005f-4d0b-8a9e-ba821f3952d4
    Nicely-Harness: claude-code
    Nicely-Harness-Version: 2.1.317

Notes (nicely-session-proof):
    {"author": {"commit": "0f3f093575f806b5d1da5d07e23551b82ef524dc", "session_id": "b7897dc7-005f-4d0b-8a9e-ba821f3952d4", "harness": "claude-code", "version": "2.1.317", "capture_source": "SessionStart+Stop"}, "review": {"base": "bf13be819abe8b8c06bf29ddaf0dadd47139ad81", "head": "0f3f093575f806b5d1da5d07e23551b82ef524dc", "session_id": "01a1a354-e946-7f63-b6a4-a832a1578c3f", "harness": "codex", "version": "0.166.0", "capture_source": "SessionStart", "completed": true, "root": true}}
```

The review receipt binds base `bf13be819abe8b8c06bf29ddaf0dadd47139ad81` and head `0f3f093575f806b5d1da5d07e23551b82ef524dc`.
The prototype uses `refs/notes/nicely-session-proof`; D069 selects `refs/notes/nicely-sessions` for production.
The final prototype notes object is `e5b66060e50addbc678e9ef56e2b111ae2a1b42b`.
No receipt mutation changed source HEAD.

## Prototype controls

The detached check ran without an active harness or a session context.

| Control | Result |
|---|---|
| same-session reviewer | reject, exit 1 |
| missing author receipt | reject, exit 1 |
| changed review target | reject, exit 1 |
| incomplete reviewer | reject, exit 1 |
| missing author capture prevents commit | reject, exit 1 |
| detached check without active harness | accept, exit 0 |

The first Codex attempt did not load the project hooks, so its review is excluded from the identity proof.
The successful invocation persisted its project-trust entry and sessions in the private scratch Codex home.
Its per-invocation hook-trust bypass applied to the vetted hooks. No user-global trust setting changed.
These prototype controls do not implement the production gate of #451.
