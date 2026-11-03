# Commands that write a file

This file says how to declare and prove each command that creates or replaces a user file, a command that only creates included.

## Declare the publication

- Give each mode of the command a row in `## Publication` of `spec.md`, as `publication.md` in the contract says
- Prove each mode with `predicts`: a successful dry run changes no file, and predicts the destination and the overwrite choice of its run
- Make the dry run check each operation of the destination as the run does, such as a hard link
- When the dry run cannot check an operation, declare `pending`, and list `publication.<operation>` in `pending_checks`
- Prove each operation of a row with `[linux] without <operation>` before `refuses`, `predicts`, or `keeps`, as `publication-helpers.md` in the contract says
- In the second table, give a reason for each obligation that no check proves

## Stage the output

- Create the stage in the folder of the destination, so that the commit stays on one file system
- Create the stage with an exclusive create and a name that no other process can predict, such as `os.CreateTemp`
- Give the stage the mode 0600, or 0700 for a folder, so that no other user reads the work before the commit
- Never write through a link at the path of the stage
- State the permission policy of the published file in `spec.md`. By default, a new file gets 0666 less the umask, and a replaced file keeps the mode of the old file
- Set that mode on the finished stage just before the commit, because a link or a rename keeps the mode of the stage
- Remove the stage on each way out of the run, a failure and a signal included

## Publish the output

- In the mode `create`, publish with `link` or `rename-noreplace`. Never check the destination and then rename over it
- When the destination refuses each operation of the row, fail closed. Never fall back to an operation that the row does not declare
- In the mode `replace`, publish with `rename`, and only after `--force` or a yes in the terminal

## Fail before or after the commit

The commit is the link or the rename that makes the destination hold the new file.

- A failure before the commit keeps the old output, and leaves an absent output absent
- A failure after the commit, such as a failed cleanup, keeps the new output. Its answer still gives the destination, and reports the failure
- A signal after the commit keeps the new output too. Prove it with `survives` and `signal`
- Prove the timeline with `holds`. The stub writes part of the output in the stage, then the marker, and waits until the kit removes the marker
- Prove any other outcome with `keeps <exit>`, which also requires that no child of the run remains

## Prove the invariants

Prove these invariants of each command that writes a user file with the commands of `scenario-commands.md` in the contract:

- The input stays unchanged: `protect` the input before each `refuses`, `signal`, `appears`, `keeps`, and `holds` call
- A failure leaves no partial output and keeps the old output: `protect` the output path, absent or old, before each of these calls
- The command never writes through a link: give an output that is a link to the input, `protect` both, and refuse the call with `refuses`
- Without `--force`, the command never replaces a file that appears during the run: `appears` creates that file, and the call keeps it

A `protect` before a `signal`, `appears`, `keeps`, or `holds` call covers that one call.
For `appears`, the stub of the wrapped program writes the marker, then waits until the file exists.

For a command that replays paid work, run `create-formal-verification-skill`.
