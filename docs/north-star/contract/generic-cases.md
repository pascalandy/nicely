# Generic cases

This part lists the generic cases of a command that reads or writes user files.
The spec of such a command starts its lines of Failures and Acceptance from this list.
Then it adds the cases of the program that the command wraps.

Each case that can occur becomes one line, with its spec ID, as [Spec IDs](spec-ids.md) defines.
Before the first review, `## Case matrix` of the spec gives a row to each such case and to each family of the intake (#423).
Each refusal in it has a pair of tests: one input passes, and one input is refused.
The line states the outcome of the case, such as a refusal with its code.
The safe outcome below is the default. Another outcome is a limit, and only the user can accept it.
A case that cannot occur for the command needs no line.
A case name is not a test. For each case of the Output group that applies to the command, `just acceptance` checks a test file of the extension's own domain.
The test makes that file, and an independent reader proves the safe outcome on it, as [Test kit](test-kit.md#acceptance-test-files) says.
An error that names the input gives its path or its item, so an agent knows which call failed.
A command that creates or replaces a user file also declares its publication, as [Publication declaration](publication.md) says.

| Group | Case | Example | Safe outcome |
|---|---|---|---|
| Input | Empty | A text file of 0 bytes | The result that the spec states for an empty input, or a refusal that names the input |
| Input | Truncated | A JSON file that ends before its last closing brace | A refusal of the input that is cut short or damaged, with an error that names the input. A partial output is only a limit that the user accepted |
| Input | Unreadable | A text file without the read permission | A refusal before any work, with an error that names the input |
| Input | Of the wrong type | A file that holds a NUL byte, given to `inspect`, which counts lines of text | A refusal with a code of the extension that names the input. The program never gets it |
| Input | Very large | A text file of several gigabytes | The command states a budget of work and memory. It finishes within that budget, or refuses with an error that names the input and the budget |
| Input | Large expansion | A small compressed text file that expands to many gigabytes | The command states the budget before the costly work, and refuses an input above it, with an error that names the input and the budget. The budget says if a timeout applies to each step or to all of the work |
| Input | With several parts | A text file that holds three notes, separated by lines `---` | The command keeps all parts, or refuses the input. It never keeps only the first part without a message. Only the user can accept a limit that keeps fewer parts |
| Input | With active content | A text file that makes the program load a remote file, such as a line `include https://example.com/part.txt`, or start another program, such as a line `!date` | A refusal, or a run in which the program loads no remote file and starts no program |
| Output | Exists, without `--force` | `notes.txt` exists before a command writes `notes.txt` from `notes.md` | A refusal before any work, and `notes.txt` stays unchanged |
| Output | Is a link to the input | `notes.txt` is a symbolic link to `notes.md` | A refusal, or a write that replaces the link. The command never writes through the link, and `notes.md` stays unchanged |
| Output | Appears during the run, without `--force` | Another process writes `notes.txt` after the check and before the write | The command never replaces that file. It publishes with a create that fails on an existing file, such as a hard link. When the file system cannot create the output without a replace, the command fails. It never uses a check and then a rename that replaces |
| Output | Temporary file | A link `notes.txt.tmp` to `notes.md`, made before the run | The command creates its temporary file with exclusive create, and with a name that no other process can predict. It never writes through a link at that path |
| Output | Keeps private data of the input | `notes.md` starts with the line `author: Ana Lima`, and `notes.txt` keeps that name | The command removes the private data, or the user accepts a limit |
| Output | Keeps data that a reader needs | `notes.md` starts with a valid header that a reader needs, with the line `encoding: latin-1`, and that header also holds the line `author: Ana Lima` | The command keeps the data that the reader needs to show the content correctly, and removes the private text. The test proves both: `Ana Lima` is not in the output, and a reader still shows the content correctly. The header uses a value that is not the default, and the test file exists before the first review. Otherwise, the user accepts a limit |
| Output | Loses data that changes the meaning | `notes.md` declares the encoding `latin-1`, and `notes.txt` keeps its bytes without that line | The output keeps the declaration, such as a text encoding or a unit, or the command applies it to the content. The intake lists each declaration that changes how a reader shows the content, not only the file types. The test file uses a value that is not the default, and the independent reference applies that declaration before it compares the content |
| Interruption | SIGINT or SIGTERM during the work | Ctrl-C while `inspect` counts the lines of a large file | Exit 130 or 143, with an answer that names the input. No partial output remains, the old output stays, and no child process remains |
| Interruption | Timeout | `inspect --timeout 1ms` on a large file | The timeout outcome, with an answer that names the input, even when the check of the program uses up the deadline. No partial output remains |
| File name | Spaces | `my notes.md` | The command passes the name as one path, unchanged |
| File name | Leading dash | `-notes.md`, which a program can read as a flag | The program reads it as a file, such as `./-notes.md`, never as a flag |
| File name | Shell characters | `notes;$(date).md` | The command passes the name as one path, and no shell reads it |
| Other | Dry run | `inspect --dry-run` leaves the file and its folder unchanged | No effect starts, and a refusal is the same as the refusal of a run. A dry run that succeeds predicts the destination and the overwrite choice of the run. It checks each ability of the destination that the run needs, or reports it as pending |
| Other | Missing program | `inspect` on a host without `python3` | `PREREQ_MISSING` with exit 78 and the install hint, before any work |

`inspect` is the example command of the extension template.
