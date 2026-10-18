# D030 Treat extensions as trusted code and grant keys one by one

Decided 2026-10-05. Extensions run with the user's rights and environment, without a sandbox. [Grants](../contract.md#grants) controls which keys core passes, with human consent per key, domain, and source on each machine.

**Why.** Tools need variables such as SSH agents and proxies. A same-user program can still read files or an unlocked keychain. Grants prevent accidental leaks and record consent; they provide no isolation.

**Rejected.** An environment allowlist that breaks tools without isolating them. Automatic grants, grants copied with shared config, and command names alone as extension identity.
