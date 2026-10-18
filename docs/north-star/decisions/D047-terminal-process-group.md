# D047 Keep a program that may read the terminal in the process group of ncly

Decided 2026-10-10, revised 2026-10-13. Terminal readers keep `ncly`'s process group. Other programs get their own group. [Programs that ncly runs](../contract.md#programs-that-ncly-runs) owns repeated signals and cleanup.

**Why.** A background process stops with SIGTTIN when it reads the terminal. Docker [reverted separate CLI-plugin groups](https://github.com/docker/cli/commit/26560ff93c42cfe8237bb0ba40cdf349fcf7d812) after that failure. Descendant discovery still governs cleanup beyond group boundaries.

**Rejected.** Separate groups for every program. Giving the extension foreground ownership with `tcsetpgrp` would keep Ctrl-C from reaching `ncly`.
