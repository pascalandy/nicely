# D008 Use one flag for every confirmation

Decided 2026-10-05, revised 2026-10-13. Use `--force` for confirmations, except key grants, which require a human in a terminal. A declined prompt changes nothing and exits 2 with `CONFIRMATION_REQUIRED`. [Modes](../contract.md#modes) defines both paths.

**Why.** One confirmation concept needs one flag. clig.dev recommends `-f` or `--force`; `gh` also uses exit 2 for cancellation.

**Rejected.** A second `--yes` flag, silent key grants during updates, and exit 0 for a declined change.
