# Video-convert adoption

This part of [COR-104-T10](../COR-104-verification-map.md#cor-104-t10-video-convert-adopts-the-loop) keeps its adoption requirements and follow-ups.
The card links the completed delivery and its evidence.
The [adoption corrections](https://github.com/pascalandy/ncly-video-convert/blob/v0.2.2/verification/corrections.md) record each finding.
The [QA ledger](../../qa/ledger.md#corrections) holds the landed corrections of owner A, #462 and #468.

## Requirements

In ncly-video-convert, update to the core and template release and run the first build of #427.
Its publication proof runs through `sdk/verification`, and keeps its model, mapping, probes, and expected results.
It declares its publications and acceptance cases.
Each correction gets a ledger row, and a correction of A or B gets the step of [D068](../../north-star/decisions/D068-verification-map.md).
It fails first on its record, which rejects the new pin ([P1][p1]), and on #427.
Done when `just signoff` and `just release` pass.

## Follow-ups

These issues remain separate work:

- [ncly-video-convert#14](https://github.com/pascalandy/ncly-video-convert/issues/14): Intel runtime warning test
- [ncly-video-convert#18](https://github.com/pascalandy/ncly-video-convert/issues/18): opt-in metadata removal
- [nicely#484](https://github.com/pascalandy/nicely/issues/484): manifest minimum ncly core
- [nicely#485](https://github.com/pascalandy/nicely/issues/485): template Release step 5 fast-forward merge
- [nicely#486](https://github.com/pascalandy/nicely/issues/486): extensiontest records the installed core separately

[p1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/proto/P1/REPORT.md
