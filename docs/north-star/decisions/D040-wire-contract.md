# D040 Serve extensions through a wire contract and a public Go SDK

Decided 2026-10-09. [Extensions](../contract.md#extensions) communicate through manifests, environment, JSON answers, exit codes, and shared record files. They compose through the CLI. The public [Go SDK](../contract.md#sdk) implements the contract for core and extensions.

**Why.** Any language can use the interface that agents already use. Shared mechanisms have one implementation.

**Rejected.** Host RPC over stdio adds another protocol before a consumer needs it. Imports from core's `internal/` packages bind extensions to private implementation.
