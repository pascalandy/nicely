# Publication helpers

This part says what the scenario commands `keeps`, `holds`, and `without` check, and how the guard of the kit runs their calls.
They prove the publication that [Publication declaration](publication.md) declares, as COR-104-T7 of [D068](../decisions/D068-verification-map.md) decides.
The extension keeps its publication code. These commands only observe it.

## The guard

On Linux, the kit builds the guard from `sdk/extensiontest/guard`, as it builds the fake child.
It runs each call of `keeps` and `holds`, and each call of a check after `without`, under the guard:

- The guard adopts each orphan of the call. After the call exits, each descendant gets two seconds to exit. Then the check fails, and names each descendant that still runs. The guard kills them after that check
- A seccomp filter sends each link and each rename of the call and of its descendants to the guard. The kit and the guard stay outside the filter. It needs no root, no mount, and no system service
- The scope of the guard is the scenario folder, outside `TMPDIR` and the XDG folders of the scratch run. These folders hold the temporary files, the config, the records, and the cache of Nicely
- In the scope, the guard makes each link and each rename itself. So it records the operation, its target, and its result

A platform other than Linux has no such filter, so no call runs under the guard there.
`without` fails there and names the platform.
`keeps` then checks that the process group of the call is empty, and `holds` observes no commit.

## The commit

The commit is the operation that publishes the destination.
After a successful run under the guard, a link or a rename must have published the destination. A rename of a file onto itself publishes nothing.
The row of the mode must declare that operation.
So a run that writes its output in place fails, and a create that renames over the destination fails too.

## keeps

`keeps <exit> <command> <arg>...` runs any registered call in the foreground, as `signal` does.

1. The call must exit with `<exit>`, and give a valid answer.
2. The call must keep the paths of the last `protect` call, and the Protected paths of the row of its mode.
3. No descendant of the call may still run, as [The guard](#the-guard) says.
4. When a run succeeds and its mode has a row, the commit uses an operation of the row. A dry run publishes nothing.

## holds

`holds <marker> <path> <command> <arg>...` proves the timeline of a publication, for a mode with a row.
The worker writes part of its output in its private stage, and then writes the marker.
Then it waits until the marker is gone, and finishes its output.

1. The marker must not exist before the call.
2. At the marker, `<path>` must have its state before the call: absent, or the same file with the same bytes and mode. A protected path cannot be `<path>`.
3. The kit then removes the marker. The run has 10 seconds to exit, and must succeed.
4. The answer gives `<path>` at the key Destination, and a regular file is there.
5. On Linux, the commit uses an operation of the row.
6. The run keeps the paths of the last `protect` call, and the Protected paths of the row.

A run that writes its partial output at `<path>` fails step 2.

## without

`without <operation> <check> <arg>...` runs one call of `refuses`, `predicts`, `appears`, or `keeps`.
During that check, the guard refuses the operation in its scope, with the error of a file system without that ability:

| Operation | Error |
|---|---|
| `link` | `EPERM`, as vfat gives |
| `rename-noreplace` | `EINVAL`, as a file system without `RENAME_NOREPLACE` gives |
| `rename` | `EPERM` |

- A call of the check must try the operation in the scope. Otherwise the check proves nothing, and it fails
- A passing `refuses`, `predicts`, or `keeps` proves the obligation of the operation, for the row of the mode of the call. `appears` and a dry run of `keeps` prove none
- A create that falls back to a check and a rename that replaces fails `predicts` or `keeps`, as [The commit](#the-commit) says
- The refusal reaches the system call, not the file system. A dry run that reads the type of the file system cannot see it. Then `keeps` proves that the run fails closed
- Write `[linux]` before the line, because another platform fails. There, the kit checks no obligation of an operation, and prints a notice
