# D008 Use one flag for every confirmation

Decided 2026-10-05, revised 2026-10-13. In non-interactive mode, `--force` answers yes to every confirmation, such as an overwrite or a deletion. There is no `--yes`. Granting a key is the exception: it waits for a human in a terminal, as principle 5 requires. In interactive mode, any answer but yes changes nothing and exits 2 with `CONFIRMATION_REQUIRED`, as [Modes](../contract.md#modes) defines.

**Why.** clig.dev asks for `-f` or `--force` when a confirmation cannot be asked. One concept gets one flag, so an agent never has to guess which of two flags a step needs. Exit 2 already means a missing confirmation, so a declined one needs no new code, and `gh` also exits 2 when a human cancels a prompt.

**Rejected.** `--yes` beside `--force`, as in Pascal's script conventions, which serve scripts that have both a prompt and a separate safety check. `--force` granting the keys a tap requests, because an update could then gain a key in silence. Exit 0 after a no, as `rm -i` gives, because a script could not tell a declined change from a done one.
