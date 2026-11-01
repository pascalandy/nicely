# Intake

This file says which facts the intake of a command gets, before any code.

## Intake of a script

Get these facts before you write code. Read the source of the script, and ask its author what the source does not show. Do not run the script to learn what it does, because even its `--help` can write or bill.

- The program and its entry point, such as `python3 tool.py`, with its source and revision, for `vision.md`
- Its inputs: the arguments, stdin, files, environment variables, and the folder that it reads
- Its result: what proves that the work happened, on which stream, and in which format
- Its effects: each write, network call, and paid request, also in a dry run or in `--help`
- Its keys and logins, and where it reads them
- How long it runs, what it starts in the background, and whether it can ask a question
- Its exit codes, and what a repeat of the same call does

Write each fact that nobody can give as an open line of `spec.md`.

## Intake of a program

A third-party program, such as an encoder, has no author to ask.
Get the facts of the script intake from its documentation and its manual page.
Also read its exit codes, its limits, and its known failure modes, such as the open issues of its tracker.
Do not run a command that writes or bills to learn its behavior.
Write each fact that the documentation does not give as an open line of `spec.md`.

## Evidence of the intake

The intake records what you read, not only where the facts are (#420):

- For each settled fact: a link to the part of the source that you read, at a fixed revision. Such a part is a function of the script at a commit, or a section of the manual page of a version (#423)
- Each fact that you did not read in a source, marked `inference`
- The result of each experiment, apart from the documented facts, with its command and what it gave
- Each declaration of the input that changes how a reader shows the content, not only the file types. `generic-cases.md` in the contract says why
- Each fact that no source gives, as an open line of `spec.md` with its ID, before the spec commit

A link alone is not evidence of a read. A reviewer of round 1 refuses a fact that only cites a source. Read the source first. Then run experiments only to confirm what you read, with a command that neither writes nor bills.
