# Design principles

This part of the [core rule index](README.md) holds the rules of `principles.md`.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| P01 | Commands use a grammar that starts with the domain, and domain help | [principles.md][principles] | [Spec test][spec-test], [shared domains scenario][shared-domains], partial: declared names |
| P02 | Use one verb for each action across domains, and reshape an incoming script to that grammar | [principles.md][principles] | Prose only |
| P03 | Interactive forms and calls without input share one implementation and give the same result | [principles.md][principles], [command.go][command] | [command.go][command], [form scenario][form] |
| P04 | A missing value without input fails with exit 2, and its hint is a full command | [principles.md][principles] | [Usage scenario][usage], [form scenario][form] |
| P05 | Machine identifiers stay in English, and agents branch on stable codes, exits, and next words | [principles.md][principles] | [Parity test][parity], [language scenario][language], partial: choices of agents |
| P06 | Machine values keep their types, units, values, meanings, and version compatibility | [principles.md][principles] | [Compatibility test][compat], partial: listed consumer cases |
| P07 | A writing command and its dry run share their preparation | [principles.md][principles] | [Kit `refuses`][paired], [its test][paired-test], partial: scenarios choose the cases |
| P08 | Overwrite or delete only with force, or with consent at a terminal | [principles.md][principles] | [Auth login][auth-login], [auth form][auth-form], partial: each handler needs cases |
| P09 | The first command that deletes a user file sends it to OS trash through the SDK, per [COR-629][trash] | [principles.md][principles] | Prose only |
| P10 | Each effect is safe to retry, and a named completed run replays without new effects | [principles.md][principles] | [Agent run boundaries][agent-run], [transcript replay][replay], partial: recorded commands |
| P11 | Give outside text to an agent only with its tools off | [principles.md][principles] | Prose only |
| P12 | Never call a check after execution a prevention of writes | [principles.md][principles] | Prose only |
| P13 | Keys stay in the keychain or the environment, never in flags, config, the repository, or a binary | [principles.md][principles] | [Auth login][auth-login], [gitleaks rules][gitleaks], partial: no scan of binaries |
| P14 | On a missing key, relay the hint and leave login to the human | [principles.md][principles] | Prose only |
| P15 | An extension gets a key only after a grant by a human | [principles.md][principles] | [Grant scenario][grant], [wrap scenario][wrap], partial: a grant file proves no author |
| P16 | Doctor checks tools and keys, installs nothing, and gives a fix for each failed check | [principles.md][principles] | [Doctor core][doctor-core], [doctor probe][doctor-probe], [doctor auth][doctor-auth] |
| P17 | Each human string comes from a catalog, English first, and an extension owns its translations | [principles.md][principles], [i18n.go][i18n] | [Catalog rules test][i18n-rules], [declarations test][extension-test], partial: no scan of literals |
| P18 | Ship the pseudo-locale, and pass `NCLY_LANG` to extensions | [principles.md][principles] | [Pseudo test][pseudo], [language scenario][language], [dispatch scenario][dispatch] |
| P19 | Core and extensions are Go, and extensions use the SDK | [principles.md][principles] | [Template test][template-test], partial: fixtures can be shell |
| P20 | Embedded Python speaks only JSON to Go, and Go renders and translates human output | [principles.md][principles] | [Protocol scenario][protocol], [protocol errors][protocol-errors], partial: the existing worker |
| P21 | A wrapped user worker can keep its own language | [principles.md][principles] | [Template test][template-test], [wrap scenario][wrap] |
| P22 | Use XDG paths under `nicely` on macOS and Linux | [principles.md][principles], [config.go][config] | [Config test][config-test] |
| P23 | Portable setup goes in the shared config, and the setup of one machine in the local config | [principles.md][principles] | [Config test][config-test] |
| P24 | Import Go libraries and wrap other tools. Credit tools in help once sdk/wrap supports it, per [COR-631][credit] | [principles.md][principles] | Prose only |
| P25 | Each release ships the license notices of its dependencies | [principles.md][principles] | [Archive check][archive], [its test][artifacts-test] |
| P26 | Build only what the card needs, and decide costly constraints before code forces a rewrite | [principles.md][principles] | Prose only |
| P27 | Extensions compose through `ncly` JSON commands, never through the code of another extension | [principles.md][principles] | Prose only |
| P28 | An extension adds commands, and never replaces or changes another command | [principles.md][principles] | [Shared conflicts scenario][conflicts], [manifest commands][manifest-commands], partial: other interference |
| P29 | Shared behavior goes to its owner in the code map, and domain steps under their extension root | [principles.md][principles] | [Lint config][golangci], partial: semantic ownership |
| P30 | Add no extension hooks | [principles.md][principles] | Prose only |
| P31 | Use the keyboard, and typed y/N confirmations that keep the question and the answer visible | [principles.md][principles] | [Auth form][auth-form], [form scenario][form] |

[principles]: ../principles.md
[trash]: ../../backlog/COR-629-os-trash.md
[credit]: ../../backlog/COR-631-tool-credit.md
[spec-test]: ../../../internal/cli/spec_test.go
[shared-domains]: ../../../testdata/script/extension_shared_domains.txtar
[command]: ../../../sdk/command/command.go
[form]: ../../../testdata/script/form.txtar
[usage]: ../../../testdata/script/usage_validation.txtar
[parity]: ../../../sdk/contract/parity_test.go
[language]: ../../../testdata/script/language.txtar
[compat]: ../../../sdk/contract/compat_test.go
[paired]: ../../../sdk/extensiontest/paired.go
[paired-test]: ../../../sdk/extensiontest/paired_test.go
[auth-login]: ../../../testdata/script/auth_login.txtar
[auth-form]: ../../../testdata/script/auth_form.txtar
[agent-run]: ../../../extensions/agent/testdata/script/agent_run_boundaries.txtar
[replay]: ../../../extensions/transcript/testdata/script/transcript_replay.txtar
[gitleaks]: ../../../.gitleaks.toml
[grant]: ../../../testdata/script/grant_dispatch.txtar
[wrap]: ../../../testdata/extensions/hello/testdata/script/wrap.txtar
[doctor-core]: ../../../testdata/script/doctor_core.txtar
[doctor-probe]: ../../../testdata/script/doctor_probe.txtar
[doctor-auth]: ../../../testdata/script/doctor_auth.txtar
[i18n]: ../../../sdk/i18n/i18n.go
[i18n-rules]: ../../../sdk/i18n/rules_test.go
[extension-test]: ../../../sdk/extension/extension_test.go
[pseudo]: ../../../sdk/i18n/pseudo_test.go
[dispatch]: ../../../testdata/script/extension_dispatch.txtar
[template-test]: ../../../scripts/template-test.sh
[protocol]: ../../../extensions/transcript/testdata/script/transcript_protocol.txtar
[protocol-errors]: ../../../extensions/transcript/testdata/script/transcript_protocol_errors.txtar
[config]: ../../../sdk/config/config.go
[config-test]: ../../../sdk/config/config_test.go
[archive]: ../../../tools/cmd/release-artifacts/archive.go
[artifacts-test]: ../../../tools/cmd/release-artifacts/main_test.go
[conflicts]: ../../../testdata/script/extension_shared_conflicts.txtar
[manifest-commands]: ../../../sdk/contract/manifest_commands.go
[golangci]: ../../../.golangci.yml
