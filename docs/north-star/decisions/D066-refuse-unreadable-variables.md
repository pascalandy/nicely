# D066 Refuse a variable or a language that ncly cannot read

Decided 2026-10-29, for item 6 of #375. [Global flags](../contract/global-flags.md) and [Configuration](../contract/configuration.md) hold the rule.
`1` turns on the variable of a switch, and nothing or `0` leaves it off.
Any other value, and a `--lang` or `NCLY_LANG` that names no language, is `USAGE_INVALID`.

**Why.** A silent off hides a typo, such as `NCLY_JSON=true`, and the agent then reads human text instead of JSON.
Serious CLIs refuse such a value instead of reading it as off.
git exits 128 for `GIT_OPTIONAL_LOCKS=maybe` and for `git -c core.bare=maybe status`.
cargo exits 101 for `CARGO_TERM_VERBOSE=maybe`, and Docker exits 1 for `DOCKER_BUILDKIT=maybe`.
Help, the version, and shell completion still work, as they do with an invalid flag value.

**Rejected.** `true` and `yes` as on, because an older `ncly` reads them as off.
A warning alone, which a caller of machine output may never read.
A check of the config and the locale variables, which other machines and programs set.
cargo's refusal of `0`, which git accepts as off.
