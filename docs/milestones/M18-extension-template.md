# M18 Extension template

Status: planned
Version: v0.18.0

## Demo

An author uses the template to wrap one CLI as an extension that follows Nicely's contract.

## Scope

[#273](https://github.com/pascalandy/nicely/issues/273) owns the extension requirements.
[D004](../north-star/decisions/D004-minimal-core.md) records the default-extension rule.
Pascal settles the remaining requirements in #273 before readiness or implementation starts.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M18-T1 | Settle the extension requirements | Pascal | — | todo |
| M18-T2 | Make the extension template ready | agent | M18-T1 | todo |

### M18-T1 Settle the extension requirements

Settle the requirements in [#273](https://github.com/pascalandy/nicely/issues/273).
Define what the template provides, what an author writes, and which wrapper errors it anticipates.
Decide how an extension's contract stays aligned, how it installs from its own repository, and how that repository runs its tests.

### M18-T2 Make the extension template ready

- **Read:** [Scope](#scope), [#273](https://github.com/pascalandy/nicely/issues/273), [Make a milestone ready](README.md#make-a-milestone-ready), [Architecture](../north-star/architecture.md)
- **Proves:** this milestone's cards and linked specs cover the requirements that Pascal settled in #273

Start only after Pascal settles #273.
Follow the readiness procedure to write the specs and add implementation cards with their proofs.
Keep the template before the author guide, with a small personal script as the first extension built from it afterward.
