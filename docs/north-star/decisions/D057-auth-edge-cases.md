# D057 Settle the edge cases of ncly auth with gh, Docker, and clig.dev

Decided 2026-10-13. Follow [ncly auth](../core-spec.md#ncly-auth) for trimmed nonempty keys, nonterminal `--stdin`, absent-key logout, and declared-service status.

**Why.** Trimming supports piped secret-store output, as `gh` does. Refusing a terminal prevents visible secret entry or an unattended wait. Missing-key logout errors expose typos. Status agrees with doctor and avoids dumping macOS's entire keychain to enumerate stored services.

**Rejected.** Docker's narrower whitespace rule for opaque passwords and successful absent-key logout. Terminal reads under `--stdin` expose secrets. JSON status must not return exit 0 for a failed report.
