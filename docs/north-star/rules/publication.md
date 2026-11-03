# Publication helpers

This part of the [core rule index](README.md) holds S59 to S63, the rules of the publication helpers of COR-104-T7.
[Publication helpers][helpers] in the contract defines them.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S59 | Only Linux runs a call under the guard, which adopts its orphans. A descendant that runs two seconds after the call exits fails the check | [Publication helpers][helpers] | [guard/main.go][guard], [guard_linux.go][guard-linux], [guarded.go][guarded], [keeps test][keeps-test], partial: macOS checks only the process group |
| S60 | `without` refuses one operation in the scenario folder for one call alone. Its check must meet the refusal, and its pass proves the obligation | [Publication helpers][helpers] | [watch_linux.go][guard-watch], [guarded.go][guarded], [publication_report.go][publication-report], [guarded test][guarded-test], [paired test][paired-test], partial: the type of the file system stays the same |
| S61 | After a successful run under the guard, an operation of its row published the destination | [Publication helpers][helpers] | [guarded.go][guarded], [predicts.go][predicts], [paired test][paired-test], [guarded test][guarded-test] |
| S62 | `keeps` checks the exact exit, the protected paths, and the children of any registered call | [Publication helpers][helpers] | [keeps.go][keeps], [keeps test][keeps-test] |
| S63 | While the worker waits after a partial staged write, the public path keeps its state before the call | [Publication helpers][helpers] | [holds.go][holds], [holds test][holds-test] |

[helpers]: ../contract/publication-helpers.md
[guard]: ../../../sdk/extensiontest/guard/main.go
[guard-linux]: ../../../sdk/extensiontest/guard/guard_linux.go
[guard-watch]: ../../../sdk/extensiontest/guard/watch_linux.go
[guarded]: ../../../sdk/extensiontest/guarded.go
[keeps]: ../../../sdk/extensiontest/keeps.go
[holds]: ../../../sdk/extensiontest/holds.go
[predicts]: ../../../sdk/extensiontest/predicts.go
[publication-report]: ../../../sdk/extensiontest/publication_report.go
[guarded-test]: ../../../sdk/extensiontest/guarded_test.go
[keeps-test]: ../../../sdk/extensiontest/keeps_test.go
[holds-test]: ../../../sdk/extensiontest/holds_test.go
[paired-test]: ../../../sdk/extensiontest/paired_test.go
