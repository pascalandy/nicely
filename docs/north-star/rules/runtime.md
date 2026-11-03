# Platform, record, runner, tui, wrap, and analysis

This part of the [core rule index](README.md) holds S27 to S51, the rules of SDK package comments.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S27 | Platform owns the support of process trees, durable state writes, and file locks | [proctree.go][proctree] | [File test][platform-file], [runner test][runner-test], partial: the implemented behavior |
| S28 | Records keep run IDs, the data of each run, execution locks, transitions, and checked success artifacts | [record.go][record] | [Record test][record-test], [agent run record][agent-record], partial: the implemented records |
| S29 | Domains choose step meanings, effects, and result checks, and record support chooses no domain action | [record.go][record] | Prose only |
| S30 | Write cleanup records with a context that is not cancelled, inside the cleanup budget | [record.go][record] | [Transcript cancel scenario][transcript-cancel], partial: one command |
| S31 | Runner owns the child environment, the session, cancellation, captured output, and the exit report | [runner.go][runner] | [Runner test][runner-test], [session test][session-test], [streams test][streams-test], partial: its own path |
| S32 | Start a program without a terminal with Setsid alone, without Setpgid | [runner.go][runner] | [Session test][session-test] |
| S33 | A program without a terminal gets no terminal descriptor | [runner.go][runner] | [Session test][session-test], [streams.go][streams] |
| S34 | Bound the kept output of a noisy program with Capture | [runner.go][runner] | [Session test][session-test], [capture.go][capture], partial: callers choose the limits |
| S35 | Read stdout and stderr through pipes of the runner, so descendants cannot block Wait | [runner.go][runner] | [Runner test][runner-test], [streams test][streams-test] |
| S36 | Signal and clean the whole process tree, not only the direct child | [runner.go][runner] | [Runner test][runner-test], [wrap stop scenario][wrap-stop] |
| S37 | Keep the signal handler until cleanup ends, also for a second Ctrl-C | [runner.go][runner] | [Stubborn cancel scenario][cancel-stubborn], [signal.go][runner-signal] |
| S38 | Shorten cleanup only in the test binary, so scenarios reach SIGKILL without a long wait | [runner.go][runner] | [Script test][script-test], [wrap stop scenario][wrap-stop] |
| S39 | Build each interactive screen from `sdk/tui` | [tui.go][tui] | Prose only |
| S40 | With no specified look, use the shared help and error styles, and print the rest plainly | [tui.go][tui] | Prose only; the [styles scenario][styles] checks the existing styles |
| S41 | Use `sdk/wrap` for one program, and `sdk/runner` for several | [wrap.go][wrap-go] | Prose only |
| S42 | A wrapped program uses the prerequisite probe of doctor | [wrap.go][wrap-go] | [Wrap scenario][wrap], [wrap stop scenario][wrap-stop] |
| S43 | A wrapped program gets empty stdin, no terminal, only named keys, and bounded or discarded output | [wrap.go][wrap-go] | [Wrap scenario][wrap], [session test][session-test] |
| S44 | Wrap maps the common failures to contract outcomes, and the decoder of the handler maps the rest | [wrap.go][wrap-go] | [Wrap scenario][wrap], partial: domain meanings |
| S45 | Pass the signal context of the handler, and use Program.Timeout for the deadline of the program | [wrap.go][wrap-go] | [Wrap stop scenario][wrap-stop], partial: any context passes |
| S46 | A probe that uses up the command deadline answers a timeout, not a missing prerequisite | [wrap.go][wrap-go] | [Wrap stop scenario][wrap-stop] |
| S47 | Set Program.Input, so runtime and signal messages name the item | [wrap.go][wrap-go] | Prose only |
| S48 | The fake child runs its action file in order, with defined output, input, sleep, exit, and missing actions | [fake/main.go][fake] | [Fake test][fake-test] |
| S49 | The env action of the fake child prints names only, never values | [fake/main.go][fake] | [Fake test][fake-test] |
| S50 | Discard a stream of a runner.Program with nil, never io.Discard, so no descendant holds its pipe | [runnerdiscard.go][runnerdiscard] | [runnerdiscard test][runnerdiscard-test], [template recipes test][template-recipes-test], partial: no variable flow |
| S51 | Add an analyzer as a package listed in Analyzers, with cases that fail first on its past mistake | [analysis.go][analysis] | Prose only |

[proctree]: ../../../sdk/platform/proctree.go
[platform-file]: ../../../sdk/platform/file_test.go
[runner-test]: ../../../sdk/runner/runner_test.go
[record]: ../../../sdk/record/record.go
[record-test]: ../../../sdk/record/record_test.go
[agent-record]: ../../../extensions/agent/testdata/script/agent_run_record.txtar
[transcript-cancel]: ../../../extensions/transcript/testdata/script/transcript_cancel.txtar
[runner]: ../../../sdk/runner/runner.go
[session-test]: ../../../sdk/runner/session_test.go
[streams-test]: ../../../sdk/runner/streams_test.go
[streams]: ../../../sdk/runner/streams.go
[capture]: ../../../sdk/runner/capture.go
[cancel-stubborn]: ../../../testdata/script/extension_cancel_stubborn.txtar
[runner-signal]: ../../../sdk/runner/signal.go
[script-test]: ../../../cmd/ncly/script_test.go
[tui]: ../../../sdk/tui/tui.go
[styles]: ../../../testdata/script/styles.txtar
[wrap-go]: ../../../sdk/wrap/wrap.go
[wrap]: ../../../testdata/extensions/hello/testdata/script/wrap.txtar
[wrap-stop]: ../../../testdata/extensions/hello/testdata/script/wrap_stop.txtar
[fake]: ../../../sdk/extensiontest/fake/main.go
[fake-test]: ../../../sdk/extensiontest/fake/main_test.go
[runnerdiscard]: ../../../sdk/analysis/runnerdiscard/runnerdiscard.go
[runnerdiscard-test]: ../../../sdk/analysis/runnerdiscard/runnerdiscard_test.go
[template-recipes-test]: ../../../scripts/template-recipes-test.sh
[analysis]: ../../../sdk/analysis/analysis.go
