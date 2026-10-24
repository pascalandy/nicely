# M24 Docs hub

Status: planned
Version: v0.24.0

## Demo

`ncly docs serve <project...>` shows the docs of a small, explicit set of projects in one local site.

## Scope

Each selected project keeps its own `docs/` folder.
The extension reads those folders and combines their rendered pages, as [D022](../north-star/decisions/D022-private-docs.md) decides.

```
ncly docs serve <project...>
```

Pass the project roots explicitly on each invocation.
The command uses one existing renderer, builds only in Nicely's disposable cache, and prints its loopback URL.
It serves until cancelled, then stops its process tree and removes only its own temporary build.
The source folders remain unchanged.

Keep project names and navigation sufficient to distinguish the selected sources.
Serve only the selected documentation and the generated cache.
Reject paths or symlinks that would expose files outside those roots.
Bind to loopback, and disable renderer behavior that executes project code or fetches remote content.

There is no persistent hub mirror or project registry.
Scanning, synchronization, visibility management, publishing, and search remain in
[M99](M99-parking-lot/docs-hub.md).
This milestone changes no core mechanism.

## Open questions

Settle each one in `extensions/docs/spec.md`, then delete this section.

- Which recurring cross-project reading task Pascal has, with its docs roots. Without one, move M24 to M99, as D061 says
- How the URL reaches the caller. Core buffers an extension's JSON stdout until exit, so a serving command cannot answer its URL in JSON
- Which existing renderer supports multiple explicit docs roots with simple setup, useful navigation, and the required local-only behavior. Try one renderer first, and compare a second only if the first fails a requirement
- Which installed programs and versions it requires, and which doctor checks and install hints describe that setup
- How its build, serving, path boundaries, port selection, and cancellation are proven with two real project docs folders. A hostile fixture holds a config hook, a raw `<script>`, and a remote image
- Readiness adds a Pascal card that picks the docs view before serving starts

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M24-T1 | Serve selected docs | agent | — | todo |
| M24-T2 | Paths and lifecycle | agent | M24-T1 | todo |

### M24-T1 Serve selected docs

- **Read:** [Scope](#scope), [D022](../north-star/decisions/D022-private-docs.md)
- **Proves:** `extensions/docs/testdata/script/docs_serve.txtar`

Use the chosen renderer to serve two explicitly supplied projects with useful navigation.
Print the local URL, and report missing docs or dependencies before starting the server.
Keep sources unchanged and generated files inside disposable cache.

### M24-T2 Paths and lifecycle

- **Read:** [Scope](#scope), [Programs that ncly runs](../north-star/contract/programs-that-ncly-runs.md)
- **Proves:** `extensions/docs/testdata/script/docs_serve_boundaries.txtar`

Prove that the listener is loopback-only and that served paths cannot escape the selected documentation.
Stop the renderer and server on cancellation.
It does not reuse the stop policy of doctor, which [#215](https://github.com/pascalandy/nicely/issues/215) questions.
A failed build leaves sources unchanged and needs no publication recovery.

## After this milestone

Pascal reads a small selection of his project docs in the local site and judges whether the aggregation saves time.
