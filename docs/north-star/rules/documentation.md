# Documentation

This part of the [core rule index](README.md) holds the rule for pinned documentation.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| S65 | Template documentation links use the selected revision. Bundled extension documentation uses the running release tag, or main for development | [Version pins][repository-versions], [bundled skills][skill-spec] | [Source links test][revision-sources], [render test][revision-render], [bundled links test][revision-bundled] |

[repository-versions]: ../contract/extension-repository/versions.md
[skill-spec]: ../../../extensions/skill/spec/nicely.md
[revision-sources]: ../../../tools/cmd/plan/template_links_test.go
[revision-render]: ../../../scripts/template-repository-test.sh
[revision-bundled]: ../../../extensions/skill/bundled_test.go
