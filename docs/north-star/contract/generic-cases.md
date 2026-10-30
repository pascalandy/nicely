# Generic cases

This part lists the generic cases of a command that reads or writes user files.
The spec of such a command starts its lines of Failures and Acceptance from this list.
Then it adds the cases of the program that the command wraps.

Each case that can occur becomes one line, with its spec ID, as [Spec IDs](spec-ids.md) defines.
The line states the outcome of the case, such as a refusal with its code.
A case that cannot occur for the command needs no line.

| Group | Case | Example |
|---|---|---|
| Input | Empty | A text file of 0 bytes |
| Input | Truncated | A JSON file that ends before its last closing brace |
| Input | Unreadable | A text file without the read permission |
| Input | Of the wrong type | A file that holds a NUL byte, given to `inspect`, which counts lines of text |
| Input | Very large | A text file of several gigabytes |
| Output | Exists, without `--force` | `notes.txt` exists before a command writes `notes.txt` from `notes.md` |
| Output | Is a link to the input | `notes.txt` is a symbolic link to `notes.md` |
| Output | Appears during the run | Another process writes `notes.txt` after the check and before the write |
| Interruption | SIGINT or SIGTERM during the work | Ctrl-C while `inspect` counts the lines of a large file |
| Interruption | Timeout | `inspect --timeout 1ms` on a large file |
| File name | Spaces | `my notes.md` |
| File name | Leading dash | `-notes.md`, which a program can read as a flag |
| File name | Shell characters | `notes;$(date).md` |
| Other | Dry run | `inspect --dry-run` leaves the file and its folder unchanged |
| Other | Missing program | `inspect` on a host without `python3` |

`inspect` is the example command of the extension template.
