# Commands that replace a file

This file says how to prove a command that replaces a user file.

For a command that replaces a user file, prove these invariants with the commands of `scenario-commands.md` in the contract:

- The input stays unchanged: `protect` the input before each `refuses`, `signal`, and `appears` call
- A failure leaves no partial output and keeps the old output: `protect` the output path, absent or old, before each of these calls
- The command never writes through a link: give an output that is a link to the input, `protect` both, and refuse the call with `refuses`
- Without `--force`, the command never replaces a file that appears during the run: `appears` creates that file, and the call keeps it

A `protect` before a `signal` or an `appears` call covers that one call.
For `appears`, the stub of the wrapped program writes the marker, then waits until the file exists.

For a command that replays paid work, run `create-formal-verification-skill`.
