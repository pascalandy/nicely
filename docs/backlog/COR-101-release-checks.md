---
id: COR-101
core: "release checks"
state: "Planned. Five checks from the reality check of 2026-10-24 wait as open questions"
---

# COR-101 release checks

## Demo

The selected release checks prove an installed extension and the release paths that the reality check found untested.

## Open questions

1. A release never tests an installed extension: `brew-qa.sh` tests ncly alone.
   Check: install video-convert from its tag, then describe, doctor, a dry run and one real conversion
2. Real Codex was proven once, on 2026-10-17 with Codex 0.163.0.
   Check: rerun the M10 live test after a Codex upgrade.
   Each run is one small paid task that Pascal authorizes
3. No release step requires macOS `just check`.
   Check: macOS `just check` before each tag
4. No test runs a dry run while a run holds the lock.
   Check: for each recorded command, a dry run during a held lock answers like the real call
5. Template update between two releases with a dependency added to go.mod conflicted by hand in #347 U1.
   Check: the acceptance test in [#347](https://github.com/pascalandy/nicely/issues/347)
