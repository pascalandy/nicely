# Contract, extension, repository, kit, and i18n

This part of the [core rule index](README.md) holds S01 to S26, the rules of SDK package comments, and S52 to S53 of COR-104-T3.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S01 | Contract declarations, codes, verdicts, and modes agree with the contract source | [codes.go][contract-codes] | [Spec test][contract-spec], [parity test][parity], [verdict test][verdict] |
| S02 | Add a consumer compatibility case for each change of a machine value or a stored format | [codes.go][contract-codes] | Prose only |
| S03 | Standalone and bundled extensions share the path of the manifest, the handler, and the answer | [extension.go][extension] | [Bundled scenario][bundled], [ncly main test][ncly-main] |
| S04 | A run extracts its workers to a private temporary folder, and removes it when the handler returns | [extension.go][extension] | [Program test][program-test], [wrap scenario][wrap] |
| S05 | A handler never deletes the worker folder of the SDK | [extension.go][extension] | Prose only |
| S06 | Worker folders go before signal cleanup ends, at least a second before core sends SIGKILL | [extension.go][extension] | [Wrap stop scenario][wrap-stop], partial: the timing follows the budgets |
| S07 | Repository checks, contract lookup, update, install, and uninstall stay on the shared export program | [install.go][repo-install], [extensionrepo main][repo-main] | [program.go][repo-program], [template recipes test][recipes-test], partial: copied recipe calls |
| S08 | Print test notices after the results of the Go packages, and keep the exit of a failed test | [install.go][repo-install], [test main][repo-test-main] | [test.go][repo-test], [template recipes test][recipes-test] |
| S09 | An extension repository runs the kit from one Go test | [extensiontest.go][kit] | [Kit test][kit-test], partial: nothing refuses a deleted call |
| S10 | The kit checks the links of the manifest, the skill, the spec, the scenarios, and the version | [extensiontest.go][kit] | [Kit test][kit-test], [links.go][links] |
| S11 | Build the pinned host and the extension, install them in scratch, then run examples and scenarios | [extensiontest.go][kit] | [extensiontest.go][kit], [kit test][kit-test], [examples test][examples-test] |
| S12 | Credit coverage only for checked JSON answers or declared text results | [extensiontest.go][kit] | [scenario.go][kit-scenario], [calls test][calls-test] |
| S13 | Scenarios use the commands of the kit, which check answers, exits, and protected pairs | [extensiontest.go][kit] | [Answer schema][answer-schema], [kit test][kit-test], [paired test][paired-test], [protect test][protect-test] |
| S14 | Scenarios get scratch HOME and XDG folders, sentinel keys, no keychain bus, and no caller extensions | [extensiontest.go][kit] | [extensiontest.go][kit], [isolation test][isolation-test], [install test][kit-install-test], partial: absolute paths stay readable |
| S15 | The kit imports no internal package of core | [extensiontest.go][kit] | [Lint config][golangci], partial: test files are exempt |
| S16 | Build the host, the extension, and the fake child from Go source | [extensiontest.go][kit], [fake/main.go][fake] | [extensiontest.go][kit], [kit test][kit-test] |
| S17 | Copy each program to its run location, and never hard-link it | [extensiontest.go][kit] | Prose only |
| S18 | Each failure of the kit names the file, the missing link, and the fix | [extensiontest.go][kit] | [Kit test][kit-test], partial: its negative fixtures |
| S19 | Each catalog entry has a description for translators | [i18n.go][i18n], [AGENTS.md][agents] | [Rules test][i18n-rules], [rules.go][i18n-rules-go] |
| S20 | Catalog IDs start with their domain | [i18n.go][i18n] | [rules.go][i18n-rules-go], [rules test][i18n-rules], partial: core prefixes |
| S21 | Never use a message ID again for another meaning | [i18n.go][i18n] | Prose only |
| S22 | A message with a number has plural forms | [i18n.go][i18n] | [Rules test][i18n-rules], [rules.go][i18n-rules-go] |
| S23 | Render plural messages with N, Size, or Duration, never with T | [i18n.go][i18n] | [i18n test][i18n-test], [plural guard][script-test], partial: paths that run |
| S24 | Format numbers, sizes, durations, and dates with i18n | [i18n.go][i18n] | Prose only |
| S25 | Translated text sets its own width | [i18n.go][i18n] | Prose only |
| S26 | Clip command-line and variable values before a refusal quotes them | [i18n.go][i18n] | [Long values scenario][long-values] |
| S52 | The kit, signoff, release, and acceptance build each environment with `NewOwnedRun`, which alone reads `os.Environ` | [Owned runs][owned-runs] | [owned_run.go][owned-run], [callerenv][callerenv], [isolation test][isolation-test], [acceptance.go][acceptance], [template recipes test][recipes-test], partial: a program with no `Env` inherits its caller |
| S53 | Each action of the export program returns one result, whose answer, notices, and exit code one writer delivers | [Owned runs][owned-runs] | [result.go][result], [result test][result-test], [isolation test][isolation-test], [template recipes test][recipes-test] |

[agents]: ../../../AGENTS.md
[contract-codes]: ../../../sdk/contract/codes.go
[contract-spec]: ../../../sdk/contract/spec_test.go
[parity]: ../../../sdk/contract/parity_test.go
[verdict]: ../../../sdk/contract/verdict_test.go
[extension]: ../../../sdk/extension/extension.go
[bundled]: ../../../testdata/script/bundled_extension.txtar
[ncly-main]: ../../../cmd/ncly/main_test.go
[program-test]: ../../../sdk/extension/program_test.go
[wrap]: ../../../testdata/extensions/hello/testdata/script/wrap.txtar
[wrap-stop]: ../../../testdata/extensions/hello/testdata/script/wrap_stop.txtar
[repo-install]: ../../../sdk/extensionrepo/install.go
[repo-main]: ../../../sdk/extensionrepo/cmd/extensionrepo/main.go
[repo-program]: ../../../sdk/extensionrepo/program.go
[recipes-test]: ../../../scripts/template-recipes-test.sh
[repo-test-main]: ../../../sdk/extensionrepo/test/main.go
[repo-test]: ../../../sdk/extensionrepo/test.go
[kit]: ../../../sdk/extensiontest/extensiontest.go
[fake]: ../../../sdk/extensiontest/fake/main.go
[kit-test]: ../../../sdk/extensiontest/extensiontest_test.go
[links]: ../../../sdk/extensiontest/links.go
[examples-test]: ../../../sdk/extensiontest/examples_test.go
[kit-scenario]: ../../../sdk/extensiontest/scenario.go
[calls-test]: ../../../sdk/extensiontest/calls_test.go
[answer-schema]: ../../../sdk/extensiontest/answer_schema.go
[paired-test]: ../../../sdk/extensiontest/paired_test.go
[protect-test]: ../../../sdk/extensiontest/protect_test.go
[isolation-test]: ../../../sdk/extensiontest/isolation_test.go
[kit-install-test]: ../../../sdk/extensiontest/install_test.go
[golangci]: ../../../.golangci.yml
[i18n]: ../../../sdk/i18n/i18n.go
[i18n-rules]: ../../../sdk/i18n/rules_test.go
[i18n-rules-go]: ../../../sdk/i18n/rules.go
[i18n-test]: ../../../sdk/i18n/i18n_test.go
[script-test]: ../../../cmd/ncly/script_test.go
[long-values]: ../../../testdata/script/long_values.txtar
[owned-run]: ../../../sdk/extensionrepo/owned_run.go
[owned-runs]: ../contract/owned-runs.md
[callerenv]: ../../../sdk/analysis/callerenv/callerenv.go
[acceptance]: ../../../sdk/extensionrepo/acceptance.go
[result]: ../../../sdk/extensionrepo/result.go
[result-test]: ../../../sdk/extensionrepo/result_test.go
