# Spec IDs

The kit reads each part of `spec.md` named `Failures` or `Acceptance`. A part starts at a heading, or at a line that starts with a bold label, such as `**Failures.**`. It ends at the next one. Each row of a table and each bullet of these parts is one promise, and starts with its spec ID in backticks:

| ID | Condition | Code |
|---|---|---|
| `missing-file` | The file is missing or is no regular file | `USAGE_INVALID` |

- `duration`: The duration of the output, within 0.1 s

A spec ID is lowercase words that hyphens join. No two lines share one. A bullet takes a colon after its ID, so a bullet that only starts with code holds no ID. A prose paragraph holds no promise.

A test claims IDs with a line of its own, `# spec: <id>...`, above the commands that prove them. The line goes in the script of a scenario, or in `acceptance.sh`, which proves the real input of `## Acceptance`.
