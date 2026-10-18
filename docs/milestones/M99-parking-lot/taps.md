# Taps

Parked from M12, whole: `ncly tap add` and `ncly tap list`, release packaging of the first-party
extensions, staged installation and source receipts, arbitrary Git remotes, automatic updates,
`tap sync`, `tap remove`, ownership and conflict recovery for updates, tap-provided skill
discovery, `skill link`, harness-specific skill destinations, standalone skill prerequisite
manifests and their doctor checks, and a private fleet extension. Today `just install` installs
`ncly` and the first-party extensions from a checkout, and Homebrew ships `ncly` itself.

Revisit when a machine or a user without a source checkout needs an extension. Start with one
first-party extension from the official tap, installed as a complete artifact that records its
source, grants no key, and overwrites no user-owned destination.
[M12 at e0a0122](https://github.com/pascalandy/nicely/blob/e0a0122c18d284b2e5b148ede34e972b388bb858/docs/milestones/M12-taps.md)
holds that slice, its open questions, and the `tap list` look that Pascal picked in M07-T1.
Revisit updates after a second real version needs installation, and skill distribution or fleet
work for a concrete consumer. That work must retain
[D035](../../north-star/decisions/D035-sync-ownership.md) where it manages persistent copies.
