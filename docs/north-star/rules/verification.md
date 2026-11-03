# Verification runs

This part of the [core rule index](README.md) holds the isolation rules of real CLI verification.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| V01 | Linux Verify uses one private D-Bus session and unlocked keyring across its steps. Missing prerequisites fail by name. Cleanup stops the daemon and bus | [Verify skill][verify] | [Real CLI test][test], [session helper][helper], [wrapper][wrapper], [cleanup][cleanup], partial: macOS keeps its existing login-keychain path |

[verify]: ../../../.agents/skills/verify-nicely/SKILL.md
[test]: ../../../cmd/ncly/keyring_linux_test.go
[helper]: ../../../.agents/skills/verify-nicely/scripts/keyring.sh
[wrapper]: ../../../.agents/skills/verify-nicely/scripts/verify.sh
[cleanup]: ../../../.agents/skills/verify-nicely/scripts/cleanup.sh
