# AGENTS.md

This folder is the fixture extension repository of the test kit: one Go module, `example.com/standalone`, laid out as [Extension repository](../../../docs/north-star/contract/extensions.md#extension-repository) says.

1. Read [vision.md](vision.md), then [spec.md](spec.md), then [Test kit](../../../docs/north-star/contract/test-kit.md).
2. Change the spec first, then the manifest, the catalog, the handler, and its scenario in `testdata/script/`.
3. Run `go test ./...` in this folder, or `just test` at the root of Nicely. `go.mod` builds the SDK and `ncly` from this checkout.

Done when `go test ./...` passes, and `just check` passes at the root.
