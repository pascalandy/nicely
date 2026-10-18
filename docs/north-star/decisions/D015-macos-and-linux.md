# D015 Target macOS and Linux with the same paths

Decided 2026-10-05. Support macOS and Linux, with Omarchy as the reference Linux. Both use [XDG paths](../contract/configuration.md) under `nicely`. Homebrew builds macOS releases from source through `pascalandy/homebrew-tap`; Linux uses GitHub archives. [M99](../../milestones/M99-parking-lot.md) defers `ncly-bin` until AUR registration reopens.

**Why.** Shared paths simplify Pascal's move toward Linux. A local Homebrew build avoids Gatekeeper and an Apple Developer account, and Homebrew generates completions.

**Rejected.** Paid signing and notarization, a cask that strips quarantine with `xattr`, and GoReleaser's deprecated `brews` section.
