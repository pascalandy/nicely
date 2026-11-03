# Formal models

This part of the [core rule index](README.md) holds S56 to S58, the TLC runner rules of COR-104-T8.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S56 | The TLC jar holds the SHA-256 that the suite pins before Java starts, from the URL or a local jar | [verification.go][verification] | [tlc.go][tlc], [verification test][verification-test], [template models test][models-test] |
| S57 | A model run writes its record only when each check gives its expected outcome, and no input changed | [verification.go][verification] | [verification.go][verification], [tlc.go][tlc], [verification test][verification-test] |
| S58 | `just check` reads only the record of the models, with their pins, inputs, and outcomes | [verification.go][verification] | [record.go][verification-record], [models.go][models], [verification test][verification-test], [template models test][models-test], partial: the record binds the model, not the code that it models |

[models]: ../../../sdk/extensionrepo/models.go
[models-test]: ../../../scripts/template-models-test.sh
[tlc]: ../../../sdk/verification/tlc.go
[verification]: ../../../sdk/verification/verification.go
[verification-record]: ../../../sdk/verification/record.go
[verification-test]: ../../../sdk/verification/verification_test.go
