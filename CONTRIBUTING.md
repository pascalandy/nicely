# Contributing to Nicely

Issues and pull requests are welcome. If you work with an AI agent, point it to [AGENTS.md](AGENTS.md).

## Before you open a pull request

1. Pick work from the current milestone in `docs/milestones/`, or open an issue first.
2. Update [cli-spec.md](docs/north-star/cli-spec.md) when you change a command, flag, JSON key, error code, or exit code.
3. Add a testscript scenario in `testdata/script/` that proves the change.
4. Add user-facing text to the English catalog, with a description. Translations are welcome as separate pull requests.
5. Run `just check`.

Keep each pull request to one change.

## How a pull request gets checked

Nicely runs its checks locally instead of on every push. The maintainer checks out your branch, runs `just signoff`, and merges when it passes. The signoff status is required on `main`.

## License

You agree that your contribution is released under the [MIT license](LICENSE).
