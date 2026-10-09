# D015 Target macOS and Linux with the same paths

Decided 2026-10-05. Nicely runs on macOS and on Linux, with Omarchy as the reference. Both systems use XDG paths under `nicely`. On macOS, a formula in the Homebrew tap `pascalandy/homebrew-tap` builds `ncly` from source. On Linux, v0.0.1 ships prebuilt archives on GitHub. The AUR package `ncly-bin` follows when AUR registration reopens, as [M99](../../milestones/M99-parking-lot.md) records.

**Why.** Pascal moves toward Linux machines over the years, and the same paths everywhere keep his future dotfiles simple. A binary built on the user's machine never meets Gatekeeper, so Nicely needs no Apple Developer account, and Homebrew generates the completions itself. Homebrew builds `gh` from source the same way. v0.0.1 installs with Homebrew or the Linux archive. The AUR, the native channel on Arch, follows when AUR registration reopens.

**Rejected.** Signing and notarizing, which costs US$99 per year. A cask that strips the quarantine flag with `xattr`, which GoReleaser discourages. GoReleaser's `brews` section, deprecated in v2.10.
