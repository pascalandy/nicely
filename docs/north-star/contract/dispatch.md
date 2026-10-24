# Dispatch

Core selects the extension by its complete command root under [Extension identity](extension-identity.md).
For example, `ncly video convert clip.mov` starts `ncly-video-convert -- clip.mov`.
A descendant such as `video convert status` passes `status`.
[Discovery](discovery.md) owns conflicts and invalid extensions.

Core validates a call to an extension from its manifest, as it validates its own commands: an unknown command or flag, a flag without its value, or an extra argument fails with `USAGE_INVALID` before the extension starts. `--help`, `--version`, `ncly help <domain>`, and a group without a verb never start it.

Core then starts the executable with canonical arguments, as [D049](../decisions/D049-canonical-arguments.md) decides.
It passes these words in order:

- The command path after the extension root
- Each caller-set flag, in manifest order, as one `--<name>=<value>` word
- `--`
- The positional arguments, including operands after the caller's `--`

A switch passes as `--<name>=true` or `--<name>=false`, and a repeated flag passes once, with its last value. A flag that the manifest declares variadic, such as the `--tag` of the fixture's `ncly hello pick`, instead passes one word per value, in the caller's order. The manifest key that declares it arrives with its first consumer, and a Go handler receives the values in order, as [SDK](sdk.md) defines. `--dry-run` and `--run-id`, which a command accepts when its `modes` sets `dry_run` or `recorded`, pass the same way after the own flags, in that order, when the caller set them. A manifest declares `--force`, `--output`, and `--timeout` as own flags of the commands that accept them, because only the command knows whether it confirms, writes a file, or runs long. A Go handler finds each of these flags in `Call.Flags` under its long name, where the SDK's record support reads `run-id`. A flag that has a variable, such as `--json`, never reaches the extension as an argument: its resolved value reaches it as that variable. For example, `ncly -v hello run -g Hi Ada --json` starts `ncly-hello run --greeting=Hi -- Ada`, with `NCLY_VERBOSE=1` and `NCLY_JSON=1`.

The extension inherits stdin and the working directory of `ncly`, and receives the environment of [Programs that ncly runs](programs-that-ncly-runs.md). Core reports no config error before it starts an extension: it resolves `lang` as help does, skipping a config file that it cannot read. The extension reports `CONFIG_INVALID` when it reads a broken file, including a wrong core key such as `lang`, as [Configuration](configuration.md) requires of every command that reads the config.
