# D040 Serve extensions through a wire contract and a public Go SDK

Decided 2026-10-09. The interface between core and an extension is a contract on the wire: the manifest, the environment, the JSON answer, the exit code, and the record files. Core injects the keys that a human granted. An extension writes its run records in the shared format, and `ncly run` reads them. An extension uses another one through the CLI, as transcript calls `ncly agent run --json`. The public `sdk/` packages are the Go implementation of this contract, and core uses them too, so each shared layer has one implementation.

**Why.** Git, `gh`, `kubectl`, and cargo plugins work this way, and agents already use the same interface. A Python or shell extension can follow the contract from the spec alone. No protocol between processes needs a version before an extension needs it.

**Rejected.** Services that the host serves over an RPC protocol on stdio, as Terraform plugins do, because the protocol would need a design and a version before the first extension. Extensions importing core's `internal/` packages, which ties them to core's code.
