# D023 Configure with a shared TOML file and a local one

Decided 2026-10-05. The shared config describes the setup the user wants, and the local config overrides it on one machine, as [contract.md](../contract.md#configuration) defines. `ncly` edits a config file only through a command whose job is to change the setup, and it keeps comments and formatting. An unknown key is a warning. Nicely's own environment variables start with `NCLY_`.

**Why.** TOML is easy to read and edit, and Pascal's tools already use it. One shared file lets a new machine reach the same setup, for example with `ncly tap sync`. The local file keeps machine paths out of the shared one. A warning on unknown keys keeps older versions of `ncly` working.

**Rejected.** YAML. A config that only humans write, with Nicely's own records kept per machine, because each new machine would then need every tap added again.
