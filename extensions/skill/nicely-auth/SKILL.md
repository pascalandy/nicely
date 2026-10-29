---
name: nicely-auth
description: Use when an ncly command needs a key, or answers AUTH_MISSING, KEYRING_UNAVAILABLE, KEY_NOT_GRANTED, or TERMINAL_REQUIRED.
---

# Nicely keys and grants

A key stays with the human. You find where each key comes from, and relay what the human must do.

- List the keys that installed extensions need, and where each one comes from, with `ncly auth status --json`. Its findings are under `services`, and it never prints a key
- On `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, `KEY_NOT_GRANTED`, or `TERMINAL_REQUIRED`, relay `hint` and `next` to the human, who stores the key with `ncly auth login` and grants it to an extension with `ncly auth grant`, or takes a grant back with `ncly auth revoke`. Never type or pipe a key yourself, and never grant one
- Auth grant and revoke select one extension with its slug, as doctor does: `ncly auth grant SERVICE SLUG`
- `ncly doctor` warns `GRANT_UNUSED` for a grant whose extension is gone from its path, or no longer declares the service. Relay its `next` to the human
