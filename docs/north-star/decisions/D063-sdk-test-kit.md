# D063 Let the SDK own the test kit of an extension repository

Decided 2026-10-20 in [#273](https://github.com/pascalandy/nicely/issues/273). The package `sdk/extensiontest` lets an extension repository test itself with `go test ./...`. It runs the scenarios against the real `ncly-<slug>`, installed with its sidecars in scratch folders. The host is the `ncly` that `go tool ncly` builds from the same Nicely version. [Test kit](../contract/test-kit.md) owns what it runs and checks. First-party scenarios that need test-only instrumentation, such as the step boundaries of `agent`, stay in core's tests.

**Why.** Every scenario runs today from core's tests in `cmd/ncly`, through private helpers, so a separate repository cannot run its own. A kit in the SDK ships with the version that it tests, through `go get`. The kit builds its fake child program from Go source, because a Go module zip keeps no executable bit.

**Rejected.** A copy of core's test helpers in each repository, which drifts. A test module with a version of its own. Moving the instrumented first-party scenarios onto the kit now.
