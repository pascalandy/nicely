# D009 Ask in a terminal, fail elsewhere

Decided 2026-10-05. A missing value opens a form in interactive mode and exits 2 in non-interactive mode. After a form, `ncly` prints the equivalent command after `Next time:`. `CI` and `NCLY_NO_INPUT` also select non-interactive mode.

**Why.** Agents never meet a prompt. Beginners get a guided path, and the last line teaches them the flags. Some agent tools run commands in a pseudo-terminal, so a terminal alone does not prove that a human is there.
