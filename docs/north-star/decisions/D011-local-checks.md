# D011 Check locally by default, test live on request, and install nothing

Decided 2026-10-05, revised 2026-10-13. `ncly doctor` stays on the machine by default. `--live` calls free endpoints only. Invalid config is a failed report finding on stdout, and independent checks still run. Doctor never installs anything, in any mode: the hint of a failed check holds the install command for this system, its `next` holds the same command as words, and the human runs it. [ncly doctor](../core-spec.md#ncly-doctor) holds the rules.

**Why.** Doctor must be safe to run at any time. A live check proves a key works without a bill. A doctor that never installs never writes, so it needs no `--dry-run`, no confirmation, and no terminal for the password of `sudo`, and `ncly auth` stays the first command that writes. The human runs the install command either way, and the hint puts it one paste away.

**Rejected.** An install after a yes in a terminal, the first version of this entry, because it made doctor a command that writes, which then needs `--dry-run`, and an install command that asks for a password must keep the terminal, as [D047](D047-terminal-process-group.md) explains. An explicit `ncly doctor --install <check>`, because it adds a writing mode to a report command for a step that one paste already does.
