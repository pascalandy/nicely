# Core rule index

This folder answers one question: which check enforces each rule of core?
Core is owner A of [D068](../decisions/D068-verification-map.md).
The [template rule index](../../../template/AGENTS/rule-index.md.jinja) and its [parts](../../../template/AGENTS/rule-index) hold the rules of owner B.

Each row gives its rule one enforcer, in the cell under **Enforced by**:

- Links to the files of its checks. A note `partial:` names what the check leaves to review
- `Prose only`, for a rule that waits for a check
- `Prose on purpose:`, then the judgment that keeps the rule in prose

`just test` checks both indexes, and `just status` counts their rules by enforcer.
A card that adds a check moves its rule up, as D068 says.
[N2][n2] inventoried these 154 rules at `ad40ebbd`.

| File | Rules |
|---|---|
| [session.md](session.md) | G01 to G23 of `AGENTS.md`, and W01 to W05 of `RULES.md` |
| [preferences.md](preferences.md) | D01 to D26 of `dev-preferences.md`: mockups, docs, and Python |
| [landing.md](landing.md) | D27 to D46 of `dev-preferences.md`: reviews, merges, stacks, UATs, and releases |
| [principles.md](principles.md) | P01 to P31 of `principles.md` |
| [sdk.md](sdk.md) | S01 to S26: contract, extension, repository, kit, and i18n. S52 to S53: owned runs and results |
| [runtime.md](runtime.md) | S27 to S51: platform, record, runner, tui, wrap, the fake child, and analysis |

## Support code

`just test` also reads each Go file of the kit and of the repository gates, outside the tests.
An enforcer of an index links it, or this table says what it does.
At 7f3ae60e, `ledger_history.go` enforced rules that no row named until b3dc8b48.

| File | What it does |
|---|---|
| [extensionrepo/install.go](../../../sdk/extensionrepo/install.go) | Stages and publishes an install, and removes it |
| [extensionrepo/recipes.go](../../../sdk/extensionrepo/recipes.go) | Runs `just check` and `just contract` of an extension |
| [extensionrepo/rule_index.go](../../../sdk/extensionrepo/rule_index.go) | Counts the rules of the index of an extension for `just check` |
| [extensionrepo/release_command.go](../../../sdk/extensionrepo/release_command.go) | Runs the programs of a release with the environment of the child |
| [extensionrepo/release_economy.go](../../../sdk/extensionrepo/release_economy.go) | Measures the declared examples for the release receipt |
| [extensionrepo/release_example.go](../../../sdk/extensionrepo/release_example.go) | Reads the words of an example without a shell |
| [extensionrepo/release_install.go](../../../sdk/extensionrepo/release_install.go) | Runs the install block of `README.md` in scratch folders |
| [extensionrepo/scenario.go](../../../sdk/extensionrepo/scenario.go) | Runs one scenario for `just scenario` |
| [extensionrepo/spec.go](../../../sdk/extensionrepo/spec.go) | Reads the lines of `spec.md` for the kit and finish |
| [extensionrepo/update.go](../../../sdk/extensionrepo/update.go) | Runs `just update` |
| [extensiontest/install.go](../../../sdk/extensiontest/install.go) | Installs the built extension into the scratch folders of the kit |

[n2]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N2-prose-triage.md
