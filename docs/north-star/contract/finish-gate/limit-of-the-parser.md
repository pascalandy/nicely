# The limit of the parser

This part of [Finish gate](../finish-gate.md) lists the forms of a severity that `just finish` fails on.

The severities are the words block, warn, and advise, in any case.
After the first three lines, outside the line `- **Severity**:` of a finding, finish fails on these forms, and only on these (#413, #416, #420, #422):

- A severity in code, such as `` `warn` ``, also in fenced code
- A cell of a table that holds only a severity, with only marks around it, such as `` `warn` ``, `**block**`, or `(advise)`
- In a row that names a finding ID, a cell whose last word is a severity, such as `| R1-F1 | really a warn |`
- The word severity, grade, grades, or graded, then only marks or spaces, then a severity, such as "Severity: block" or "Grade: warn"
- A severity, then a hyphen and the word level, such as "warn-level". The text "block-level" fails too, so rephrase it
- Under `## Findings`, the word severity alone

The first line `- **Assertion**:` of `## First failing run` is exempt, because it copies a line of the scenario. Finish reads only that line as the assertion.
Finish reads each other line as prose, and it passes. This is also true when the line grades in other words, such as "the grade is warn" or "a block for me".
The reviewer gives each grade only in the bullet **Severity** of its finding, and moves each other grade there before the verdict.
