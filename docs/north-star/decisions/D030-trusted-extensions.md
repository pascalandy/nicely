# D030 Treat extensions as trusted code and grant keys one by one

Decided 2026-10-05. A program that `ncly` runs gets the environment of `ncly` minus every key that is not granted to it, as [contract.md](../contract.md#programs-that-ncly-runs) defines. No program runs in a sandbox. A key reaches an extension only after a human grants it in a terminal. Grants stay on one machine and bind the extension's source and domain. An extension from a different source cannot inherit a grant through its name, and an update that declares a new key gets nothing until a human grants it.

**Why.** Git, `kubectl`, `gh`, and cargo pass the user's environment to their plugins, and real tools need it, such as the SSH agent and proxies. An extension runs with the user's rights: on macOS, go-keyring stores keys through `/usr/bin/security`, which any process can call, and on Linux any program of the session can read an unlocked keychain. Grants are consent and protection against accidental leaks, not isolation, and the docs say so.

**Rejected.** A minimal list of variables plus the ones a manifest declares, because it breaks tools that rely on the environment and still isolates nothing. Granting a key without a human, such as every key a tap declares when an agent adds the tap. Copying grants between machines with shared config, or treating the extension's command name as its identity.
