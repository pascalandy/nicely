# D001 Build from scratch in Go with Cobra and Charm

Decided 2026-10-05, revised 2026-10-14. Nicely uses Go, Cobra, and Charm's Lip Gloss, Huh, Bubble Tea, Bubbles, and Glamour.

Lip Gloss v2.0.6 could hang at package load through `tmux info`, even for `ncly --version`. Pascal approved a temporary fork on 2026-10-14. [M07](../../milestones/M07-forms.md#m07-t2-lip-gloss-route) owns the fix, its proof, and removal after an upstream release. Until that fix, `sdk/tui` uses `x/ansi` and `colorprofile`, and affected commands ship non-interactive output first.

**Why.** Cobra follows established CLI practices. Charm covers the interface in one library family.

**Rejected.** The Stripe CLI has too much Stripe-specific code. Fang hard-codes English headings and uses Nicely's `-v` for `--version`. Checked on 2026-10-05.
