# Manifest examples

The kit runs every manifest example in its own scratch configuration, before the scenarios.
It adds `--dry-run` when the example's command supports that mode, including examples declared on a parent command.
Without that mode, it runs only commands whose `effects` explicitly declare `paid = false`.
A missing effects declaration permits no assumption.
Each skipped example prints its reason, including commands outside the extension's manifest.
The template runs tests through `go run ./internal/export test`, which keeps package list output without `-v`.
The SDK wrapper prints collected notices to stderr after the tests, including when tests fail.
It preserves their exit code and keeps stdout valid for `go test -json`.
Core keeps Go's result cache and uses the rendered template sample to check notices.
`extensionrepo.Notice(message)` collects one notice for that run and returns any write error.
Without the wrapper, it prints to test output, which Go can hide for passing packages.
Other kit checks use this path for notices, including each declared code's `untested` reason after the scenarios.

Examples run as argument lists, without a shell, with single quotes, double quotes, and backslash escapes supported.
The kit forces JSON and non-interactive mode, checks each answer through `Coverage.Observe`, and requires exit 0.
Neither an example nor its setup calls can replace a scenario for command coverage.

An optional `testdata/examples.txtar` supplies setup commands and input files for every example.
Its setup uses the scenario commands, including `fake` and `grant`, with the same fake programs and sentinel keys.
Fixtures supply files such as `notes.md` and `task.md`, but cannot supply missing arguments or flags.
Each example gets a fresh copy, so a previous example cannot affect its result.
Other runners call `Coverage.RunExamples` with that fixture and their scenario setup.
