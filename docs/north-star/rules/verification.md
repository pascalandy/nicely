# Verification runs

This part of the [core rule index](README.md) holds the isolation rules of real CLI verification.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| V01 | Linux Verify uses one private D-Bus session and unlocked keyring across its steps. Missing prerequisites fail by name. Cleanup, missing owned state, or a six-hour lifetime ends the session. A normal stop records exit 0 without a shell diagnostic | [Verify skill][verify] | [Real CLI test][test], [lifecycle test][lifetime-test], [session helper][helper], [wrapper][wrapper], [cleanup][cleanup]. Partial: macOS keeps its existing login-keychain path |
| V02 | A future enforced build binds each author and a completed different root reviewer session to exact commits, using harness-captured IDs | [D069][sessions] | Prose on purpose: this decision specifies the design; #451 implements production enforcement after refinement. The [scratch proof][session-proof] is evidence of capture only |

[sessions]: ../decisions/D069-recorded-sessions.md
[session-proof]: ../decisions/D069/proof.md
[verify]: ../../../.agents/skills/verify-nicely/SKILL.md
[test]: ../../../cmd/ncly/keyring_linux_test.go
[lifetime-test]: ../../../cmd/ncly/keyring_lifecycle_linux_test.go
[helper]: ../../../.agents/skills/verify-nicely/scripts/keyring.sh
[wrapper]: ../../../.agents/skills/verify-nicely/scripts/verify.sh
[cleanup]: ../../../.agents/skills/verify-nicely/scripts/cleanup.sh
