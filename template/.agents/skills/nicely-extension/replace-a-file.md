# Commands that write a file

This file says how to declare and prove a command that creates or replaces a user file.

## Declare the publication

- Give each mode of the command a row in `## Publication` of `spec.md`, as `publication.md` in the contract says
- Prove each mode with `predicts`: a successful dry run changes no file, and predicts the destination and the overwrite choice of its run
- Make the dry run check each operation of the destination as the run does, such as a hard link
- When the dry run cannot check an operation, declare `pending`, and list `publication.<operation>` in `pending_checks`
- In the second table, give a reason for each obligation that no check proves, such as the operation `link`

## Prove the invariants

For a command that replaces a user file, prove these invariants with the commands of `scenario-commands.md` in the contract:

- The input stays unchanged: `protect` the input before each `refuses`, `signal`, and `appears` call
- A failure leaves no partial output and keeps the old output: `protect` the output path, absent or old, before each of these calls
- The command never writes through a link: give an output that is a link to the input, `protect` both, and refuse the call with `refuses`
- Without `--force`, the command never replaces a file that appears during the run: `appears` creates that file, and the call keeps it

A `protect` before a `signal` or an `appears` call covers that one call.
For `appears`, the stub of the wrapped program writes the marker, then waits until the file exists.

For a command that replays paid work, run `create-formal-verification-skill`.
