# D008 Use one flag for every confirmation

Decided 2026-10-05. In non-interactive mode, `--force` answers yes to every confirmation, such as an overwrite or a deletion. There is no `--yes`. Installing a prerequisite and granting a key are the exceptions: both wait for a human in a terminal, as principles 5 and 6 require.

**Why.** clig.dev asks for `-f` or `--force` when a confirmation cannot be asked. One concept gets one flag, so an agent never has to guess which of two flags a step needs.

**Rejected.** `--yes` beside `--force`, as in Pascal's script conventions, which serve scripts that have both a prompt and a separate safety check. `--force` granting the keys a tap requests, because an update could then gain a key in silence.
