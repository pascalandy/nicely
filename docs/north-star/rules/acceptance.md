# Acceptance cases

This part of the [core rule index](README.md) holds S59 of COR-104-T5: the acceptance cases of the test kit.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S59 | An acceptance case reads the output with a reader that is not the extension or a wrapped program, and compares the reference. It searches each private text, and its bad variant fails. Each Output case of `## Case matrix` needs a case | [Acceptance cases][acceptance-cases] | [acceptance.go][kit-acceptance], [acceptance_matrix.go][kit-acceptance-matrix], [acceptance test][acceptance-test], [template acceptance test][template-acceptance-test], partial: the reviewer judges the independence of the reader and the reference. The reviewer also checks that the matrix names each case that applies |

[acceptance-cases]: ../contract/acceptance-cases.md
[kit-acceptance]: ../../../sdk/extensiontest/acceptance.go
[kit-acceptance-matrix]: ../../../sdk/extensiontest/acceptance_matrix.go
[acceptance-test]: ../../../sdk/extensiontest/acceptance_test.go
[template-acceptance-test]: ../../../scripts/template-acceptance-test.sh
