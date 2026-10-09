# D006 Serve agents through the CLI itself

Decided 2026-10-05. Agents operate `ncly` through the shell. `ncly` launches agents through their own non-interactive modes. Nicely uses no agent SDK.

**Why.** A shell command is the one interface every harness shares, so Nicely stays independent of any single agent.

**Rejected.** Limiting Nicely to Claude.
