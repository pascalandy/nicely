# Other folders

This part of [Architecture](../architecture.md#code-map) says what each folder outside the code map holds.

- `cmd/ncly/`: the entry point. It passes core an explicit list of bundled extensions and the keychain backend. When started under an extension's name, it runs that extension
- `internal/specdoc/`: reads the tables of the north-star docs for the tests that compare code and spec
- `testdata/script/`: core's testscript scenarios. An extension keeps its own in `extensions/<name>/testdata/script/`, and the test binary of `cmd/ncly` runs both
- `testdata/extensions/hello/`: the Go fixture extension that proves the SDK, laid out as a first-party extension. `./...` skips it, so `just lint` and `just test` name it
- `testdata/extensions/keyed/`: the Go fixture whose commands need a key: `keyed fetch` reads it through `Call.Key`, and `keyed relay` passes it to a program through `sdk/wrap`. Only the scenarios install it, so `hello`, which `just verify` installs in every run, declares no key, and doctor passes in a run without one
- `testdata/extensions/video-convert/` and `video-subtitle/`: named fixtures for root execution and sibling completion. `testdata/extensions/shared/export/` exports their embedded sidecars for verification
- `testdata/extensions/standalone/`: the whole-domain fixture repository of the test kit, a separate Go module that tests `sdk/extensiontest` against `go tool ncly`. Its named-extension counterpart is `testdata/extensions/standalone-named/`. `./...` skips both modules, so `just test`, `just lint`, `just fmt-check`, and `just tidy-check` cover both explicitly
- `copier.yml` and `template/`: the extension template, which creates an extension repository. Its Go files end in `.jinja`, so `./...` never builds them. `just template` renders it with `scripts/template-test.sh` and checks the rendered repository
- `tools/`: a separate module that pins the dev tools and holds the plan checker behind `just next` and `just test`
