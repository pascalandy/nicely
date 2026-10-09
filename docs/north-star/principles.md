# Nicely principles

These rules shape every design choice in Nicely.

1. **Domain first.** Commands read `ncly <domain> [<resource>] <verb>`, such as `ncly video convert`, with one verb per action across domains. `ncly <domain> --help` lists everything a domain does. A few commands stand alone, such as `ncly doctor`. A script that joins Nicely is reshaped into this grammar, whatever its interface was before.
2. **One implementation, two faces.** In interactive mode, a missing value opens a short form. In non-interactive mode, the same command fails with exit code 2 and a hint that shows the full command. Both faces produce the same result.
3. **A stable agent contract.** Machine identifiers stay in English. [Compatibility](contract.md#compatibility) protects JSON types, units, values, and meanings as well as key names, and defines how versions evolve
4. **Safe by default.** Every writing command shares preparation with `--dry-run`. Overwriting or deleting needs `--force` or a yes in the terminal. Deleted files go to the OS trash. [Retry safety](contract.md#retry-safety) covers all effects, including free writes. Outside text reaches an agent only with its tools turned off, and a check after execution is never described as write prevention
5. **Keys stay with the human.** Keys live in the OS keychain or in environment variables, never in flags, config files, the repository, or the binary. An agent that meets a missing key relays the hint, and the human runs `ncly auth login`. A key reaches an extension only after a human grants it.
6. **Check everything, install nothing silently.** `ncly doctor` checks tools and keys. Nicely installs a prerequisite only after a human says yes.
7. **Every word Nicely shows can be translated.** Core and each extension keep their text in message catalogs, never in code. English ships first, and Canadian French follows in M22. A pseudo-locale exposes untranslated text from M00. Extensions receive the language in `NCLY_LANG` and translate their own text.
8. **Core and the first-party extensions are written in Go.** An extension may embed a Python program while it moves to Go. That program talks to Go in JSON only, and Go renders and translates everything a human sees.
9. **Same paths on every machine.** Config, data, state, and cache follow XDG on macOS and Linux, under `nicely`. The shared config describes the setup the user wants and can travel between machines. The local config holds what belongs to one machine.
10. **Import Go, wrap the rest, credit both.** Nicely imports Go libraries and wraps other tools, such as `ffmpeg`. The help of a wrapping command names the tool, and every release ships the license notices of its dependencies.
11. **Build what your card needs, and decide early what is costly to change.** Write a rule before the milestone that codes it when a later milestone would otherwise force a rewrite.
12. **Compose through the CLI.** An extension uses another extension through `ncly <domain> ... --json`, the interface that agents use, never through its code. Every new extension therefore adds a command that agents and other extensions can use at once.
