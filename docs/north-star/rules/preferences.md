# Mockups, docs, and Python

This part of the [core rule index](README.md) holds D01 to D26 of `dev-preferences.md`.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| D01 | Start each choice of appearance with two or three HTML mockups that differ | [dev-preferences.md][dev] | Prose only |
| D02 | Look at the mockup in a browser | [dev-preferences.md][dev] | Prose only |
| D03 | Publish the mockup, and give Pascal its checked URL | [dev-preferences.md][dev] | Prose only |
| D04 | Wait for the choice of Pascal before you change real components | [dev-preferences.md][dev], [tui.go][tui] | Prose only |
| D05 | Keep mockups and private hosting receipts outside the repository | [dev-preferences.md][dev] | Prose only |
| D06 | Publish the same mockup file again to keep its URL | [dev-preferences.md][dev] | Prose only |
| D07 | The opening of each document states its one question | [dev-preferences.md][dev] | Prose only |
| D08 | Keep each meaning in one place, and link to it elsewhere | [dev-preferences.md][dev] | Prose only |
| D09 | A file holds at most 2,500 tokens, and `AGENTS.md` at most 1,250 | [dev-preferences.md][dev] | [Limits][limits], partial: a file that grows |
| D10 | Use one term for each concept, and update the glossary before a new or replaced term | [dev-preferences.md][dev], [AGENTS.md][agents] | Prose only |
| D11 | A principle links the decisions behind it | [dev-preferences.md][dev] | Prose only; the [docs check][docs] reads only the links that exist |
| D12 | Each decision has a file and a row in the index | [dev-preferences.md][dev] | [Decisions check][decisions], [plan test][plan-test] |
| D13 | Keep each diagram as Mermaid source in the document that it explains | [dev-preferences.md][dev] | Prose only |
| D14 | Fix an error of the docs in the commit that makes it | [dev-preferences.md][dev] | Prose only |
| D15 | Links, headings, card references, and code references stay valid | [dev-preferences.md][dev] | [Docs check][docs], [plan][plan], [plan test][plan-test] |
| D16 | Prose holds no em dash and no curly quote | [dev-preferences.md][dev] | [Plan][plan], [plan test][plan-test] |
| D17 | An instruction has one action and at most 20 words, and a sentence has one topic | [dev-preferences.md][dev] | Prose only |
| D18 | A new or changed Markdown sentence or full-line Go comment has at most 25 words | [dev-preferences.md][dev] | [Limits][limits-prose], [limits test][limits-prose-test], partial: new text |
| D19 | Keep the exact syntax of commands and identifiers when you simplify prose | [dev-preferences.md][dev] | Prose only |
| D20 | Follow serious CLI conventions where the specs are silent, and record the comparison of a new rule in a decision | [dev-preferences.md][dev] | Prose only |
| D21 | Run Python, its environments, and its dependencies with uv | [dev-preferences.md][dev] | [`just python`][justfile], partial: other shell commands |
| D22 | Keep the pins of Python tools in the recipe, and the pins of dependencies in the script metadata | [dev-preferences.md][dev] | [justfile][justfile], partial: no check finds a removed pin |
| D23 | Take the Python of the transcript metadata, and run the checks with it | [dev-preferences.md][dev] | [check.sh][check], partial: no check requires uv |
| D24 | Use the documented uv prefix for check, test, scenario, and signoff | [dev-preferences.md][dev] | Prose only |
| D25 | Install no system Python, and change no shell profile, for the checks | [dev-preferences.md][dev] | Prose only |
| D26 | Run macOS checks on the mini, use the mbp only as the fallback, and say why | [dev-preferences.md][dev] | Prose only |

[dev]: ../dev-preferences.md
[agents]: ../../../AGENTS.md
[tui]: ../../../sdk/tui/tui.go
[limits]: ../../../tools/cmd/plan/limits.go
[limits-prose]: ../../../tools/cmd/plan/limits_prose.go
[limits-prose-test]: ../../../tools/cmd/plan/limits_prose_test.go
[docs]: ../../../tools/cmd/plan/docs.go
[decisions]: ../../../tools/cmd/plan/decisions.go
[plan]: ../../../tools/cmd/plan/main.go
[plan-test]: ../../../tools/cmd/plan/main_test.go
[justfile]: ../../../justfile
[check]: ../../../scripts/check.sh
