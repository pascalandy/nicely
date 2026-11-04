# Acceptance cases

This part of the [core rule index](README.md) holds S59 of COR-104-T5: the acceptance cases of the test kit.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S59 | An acceptance case uses an independent reader and compares its output with the reference. It searches each private text and refuses a bad variant. Each matrix Output case needs a case. A privacy exception declares the case in its accepted Limit row. Every matrix row of that case names the limit ID | [Acceptance cases][acceptance-cases] | [acceptance.go][kit-acceptance], [acceptance_matrix.go][kit-acceptance-matrix], [accepted-limit parser][spec-limits], [acceptance test][acceptance-test], [privacy-limit test][privacy-limit-test], [template acceptance test][template-acceptance-test]. Partial: the reviewer judges reader independence, reference independence, matrix completeness, and whether the accepted description covers the retained metadata |

[acceptance-cases]: ../contract/acceptance-cases.md
[kit-acceptance]: ../../../sdk/extensiontest/acceptance.go
[kit-acceptance-matrix]: ../../../sdk/extensiontest/acceptance_matrix.go
[acceptance-test]: ../../../sdk/extensiontest/acceptance_test.go
[spec-limits]: ../../../sdk/extensionrepo/spec.go
[privacy-limit-test]: ../../../sdk/extensiontest/privacy_limits_test.go
[template-acceptance-test]: ../../../scripts/template-acceptance-test.sh
