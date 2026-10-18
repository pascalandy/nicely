# D005 Add extensions the Git way and carry them in taps

Decided 2026-10-05, revised 2026-10-09. Discover `ncly-<domain>` executables through a static manifest. Bundled Go extensions share their contract, as [Extensions](../contract/extensions.md) defines.

**Scope update, 2026-10-17.** [M99](../../milestones/M99-parking-lot/taps.md) defers tap distribution. `just install` installs first-party extensions from a checkout, with their existing skill folders.

**Why.** Executable discovery follows Git, `kubectl`, and cargo. The manifest adds inspectable capabilities and a verifiable output contract without restricting the implementation language.

**Rejected.** Taps for skills only, capability guesses from names or versions, and successful results from malformed protocol output. Each extension translates its own text because core cannot know its strings.
