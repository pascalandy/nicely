# D058 Grant a key to an extension at its path, and let the extension answer what it lacks

Decided 2026-10-14. Bind grants to service, domain, and discovery path without resolving symlinks. [Grants](../contract/grants.md) and [ncly auth](../core-spec.md#ncly-auth) own consent, revocation, storage, and `NCLY_MISSING_KEYS`.

**Why.** The discovered path survives Homebrew upgrades but distinguishes a shadowing executable. Hashes and resolved link targets change on upgrades. Only the extension knows whether this invocation needs a declared key. Missing-key codes distinguish absent consent from absent credentials.

**Rejected.** Startup prompts may request unused keys; prompts after refusal may repeat effects. Login with piped input cannot obtain grant consent. Config-based grants need an editor prematurely. `--force` never replaces a human grant.
