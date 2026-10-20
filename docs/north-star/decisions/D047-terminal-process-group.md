# D047 Keep a program that may read the terminal in the process group of ncly

Decided 2026-10-10, revised 2026-10-13 and 2026-10-20. Terminal readers keep `ncly`'s process group. Every other program starts in a new session, which gives it its own process group and no controlling terminal. It gets no descriptor of the terminal either. [Programs that ncly runs](../contract/programs-that-ncly-runs.md) owns repeated signals and cleanup.

**Why.** A background process stops with SIGTTIN when it reads the terminal. Docker [reverted separate CLI-plugin groups](https://github.com/docker/cli/commit/26560ff93c42cfe8237bb0ba40cdf349fcf7d812) after that failure. A process group of its own only moves a program out of the foreground, so that read still stops it until its timeout. A new session has no controlling terminal, so the same read fails at once. A shell probe under a PTY showed both behaviors on Linux and macOS on 2026-10-20. A process outside the session of the terminal reads a descriptor of it without a stop, so the runner passes it none. Descendant discovery still governs cleanup beyond group boundaries.

**Rejected.** Separate groups for every program. Giving the extension foreground ownership with `tcsetpgrp` would keep Ctrl-C from reaching `ncly`. Both a new session and `setpgid`, which refuses a session leader.
