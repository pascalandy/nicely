# M23 Shared domains

Status: ready
Version: v0.23.0

## Demo

`ncly video --help` lists `convert` and `subtitle` from two named video extensions installed together.

## Scope

[D064](../north-star/decisions/D064-shared-domains.md) records Pascal's decision of 2026-10-24.
[Extension identity](../north-star/contract/extension-identity.md) owns identity and command ownership.
Whole-domain extensions retain their behavior.
This milestone adds no video implementation and no install-time extension offer.

The host must run named extensions before it can keep sibling extensions apart.
The test kit then tests a named extension through its pinned host.
The template then uses that test kit.
Each card delivers one capability in one pull request, using the [code map](../north-star/architecture.md#code-map).
M21 Docs hub and M22 Canadian French become M24 and M25 under [D042](../north-star/decisions/D042-milestones-and-cards.md).
Neither milestone has a done card.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M23-T1 | Run a named extension | agent | — | todo |
| M23-T2 | Keep sibling extensions apart | agent | M23-T1 | todo |
| M23-T3 | Test a named extension repository | agent | M23-T2 | todo |
| M23-T4 | Create named extensions from the template | agent | M23-T3 | todo |

### M23-T1 Run a named extension

- **Read:** [Extension identity](../north-star/contract/extension-identity.md), [Discovery](../north-star/contract/discovery.md), [Manifest](../north-star/contract/manifest.md), [Valid manifest](../north-star/contract/valid-manifest.md), [Dispatch](../north-star/contract/dispatch.md), [Answer check](../north-star/contract/answer-check.md), [Usage](../north-star/contract/usage.md), [SDK](../north-star/contract/sdk.md), [Completion values](../north-star/contract/completion-values.md), [Error codes](../north-star/contract/error-codes.md), [ncly describe](../north-star/core-spec.md#ncly-describe), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `testdata/script/extension_shared_domains.txtar`, `testdata/script/extension_shared_conflicts.txtar`, `extensions/skill/testdata/script/skill_shared_domains.txtar`, `extensions/skill/testdata/script/skill_nicely.txtar`. Fixtures prove root dispatch, direct execution, discovery conflicts, shared help, full-path completion, describe, slug-named sidecars, prefixes, and skill lookup

Implement manifest name, slug derivation, and validation in `sdk/contract`.
Update `Check`, `Main`, and `Export` in `sdk/extension`, and root handling in `sdk/command`.
Update discovery, dispatch, help, completion, and describe in `internal/cli`.
Update `extensions/skill` through describe, never through an import of host discovery.
Update the bundled nicely skill's lookup guidance from `ncly skill view DOMAIN` to the slug.
Export the slug-named sidecars from each fixture build.
Apply the slug prefixes to catalog IDs and own codes.
Use host scenarios and fixtures for this card, before the external test kit supports a name.

Include a fixture whose `video convert` root takes one positional argument and has no descendants, keys, config, or status command.
Assert that `ncly video convert clip.mov` reaches that extension with `clip.mov`.
Prove the direct executable call `ncly-video-convert clip.mov` too.
Prove duplicate-slug selection and whole-domain precedence across different search folders.
Prove reserved domains, invalid ownership, a third identity part, and a broken sibling.
Prove that a sibling list command named by `values_from` supplies completion through its full command path.
Keep existing whole-domain scenarios green.
Keep oversized files within the growth rules of just limits through focused extraction where needed.

### M23-T2 Keep sibling extensions apart

- **Read:** [Extension identity](../north-star/contract/extension-identity.md), [Discovery](../north-star/contract/discovery.md), [Configuration](../north-star/contract/configuration.md), [Status command](../north-star/contract/status-command.md), [Grants](../north-star/contract/grants.md), [ncly doctor](../north-star/core-spec.md#ncly-doctor), [ncly auth](../north-star/core-spec.md#ncly-auth)
- **Proves:** `testdata/script/extension_shared_ownership.txtar`, `testdata/script/extension_shared_grants.txtar`. Doctor reports config ownership, components, and status paths by slug. Auth and grants keep sibling extensions separate

Update config ownership checks, doctor components, status dispatch, grants, and auth selectors in `internal/cli` to use the slug.
Prove that only the selected manifest defines known config keys, under Discovery and Configuration.
Prove doctor selects each slug and calls that extension's status command beneath its complete root.
Add one scenario proving that a grant for one named extension gives its sibling no key.
Keep existing whole-domain doctor and auth scenarios green.

### M23-T3 Test a named extension repository

- **Read:** [Test kit](../north-star/contract/test-kit.md), [Extension identity](../north-star/contract/extension-identity.md), [SDK](../north-star/contract/sdk.md), [Extension repository](../north-star/contract/extensions.md#extension-repository), [justfile](../../justfile)
- **Proves:** `sdk/extensiontest/`, `testdata/extensions/standalone/`, `testdata/extensions/standalone-named/`, `justfile`. A named fixture passes go test through the kit's real host and executable, in scratch folders. Every justfile check of the existing standalone module also covers the named module

Update `sdk/extensiontest` to build, install, run, and grant to a slug-named executable.
Check scenario coverage and skill references against the complete extension root.
Prove that a sibling reference is allowed and cannot satisfy this extension's scenario coverage.
Retain a whole-domain fixture to prove the existing kit interface still works.
Add the named fixture in `testdata/extensions/standalone-named/`.
Extend every justfile check that names the existing standalone module to cover the new module.
Prove go test, lint, formatting, and tidy checks cover both standalone modules.
The kit imports only public SDK packages and reaches the host through go tool ncly.

### M23-T4 Create named extensions from the template

- **Read:** [Extension repository](../north-star/contract/extensions.md#extension-repository), [Extension identity](../north-star/contract/extension-identity.md), [Test kit](../north-star/contract/test-kit.md), [D062](../north-star/decisions/D062-copier-template.md), [ncly skill](../../extensions/skill/spec.md#ncly-skill)
- **Proves:** `copier.yml`, `template/`, `scripts/template-test.sh`, `extensions/skill/testdata/script/skill_nicely.txtar`. A named render and a whole-domain render pass their checks and scratch install

Add the optional name input to Copier and carry the slug through the generated repository.
Keep the root Go package named after the domain.
Update the bundled nicely skill's template command and the generated authoring instructions.
Prove named sidecars, skill lookup, and install replacement by slug.
Prove that just update preserves author changes and the whole-domain default.
Nicely's just check runs both rendered repositories against the host and kit from the preceding cards.

## After this milestone

Build the public repository `pascalandy/ncly-video-convert` from the released template, after this milestone ships.
Its command wraps HandBrakeCLI as `ncly video convert <file>`.
That extension is a separate repository, not a Nicely card.

On the Mini, run `bash .agents/skills/release-nicely/scripts/brew-qa.sh v0.23.0` under [release-nicely](../../.agents/skills/release-nicely/SKILL.md#verify-and-hand-off).
Follow [dev-preferences](../north-star/dev-preferences.md#run-macos-checks-on-the-mini) for the machine choice.
Neither follow-up blocks the next milestone's first card.
