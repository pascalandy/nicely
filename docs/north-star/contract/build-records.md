# Build records

This part says how an extension repository records its intake and its readiness reviews, which [Finish gate](finish-gate.md) reads.
Two typed records hold this state, as COR-104-T4 of [D068](../decisions/D068-verification-map.md) asks:

- `intake.toml` holds the facts of the intake and the experiments
- `review.toml` holds the last readiness review and the ledger of the builds

Each record is TOML, like `manifest.toml`. Finish reads only these records, never the Markdown views.
A key that the record does not define fails `just finish`, such as `grade` in a finding. Each key is lowercase, so `Severity` fails too.
So a state or a grade has one place, and prose never holds one (#422).

## Views

`just records` writes the part `## Intake` of `vision.md` from `intake.toml`, and `readiness.md` from `review.toml`.
It writes no `readiness.md` while `review.toml` is missing.
It fails, and writes nothing, on a record that is not TOML or that holds an unknown key.
`just finish` fails when a view differs from what `just records` writes.
Change the record, never the view, then run `just records`, and commit both.

## The intake record

Each `[[facts]]` table is one fact. Its `kind` is one of three:

| Kind | Meaning | Fields |
|---|---|---|
| `sourced` | A source gives the fact | `statement`, `source`, `revision`, and `read`: the part of the source that the author read |
| `inference` | No source gives the fact, and the author infers it | `statement` only |
| `unknown` | Nobody can give the fact yet | `statement`, and `id`: the spec ID of its line in `spec.md` |

An unknown fact opens a line.
At the Spec commit of the build, its `id` is a row under `## Open lines` of `spec.md`.
Later, that line is open, a line of Failures or Acceptance, or an accepted limit.

Each `[[experiments]]` table is one experiment, apart from the facts, with its `command` and its `result`.
A line of `intake.toml` that starts with `# Fill:` is a line of the form that is not filled.

```toml
[[facts]]
kind = "sourced"
statement = "Exit codes: 0 with one JSON object on stdout, and 2 for a wrong argument count"
source = "worker/inspect.py of the template"
revision = "the Nicely version that go.mod pins"
read = "the function main"

[[facts]]
kind = "unknown"
statement = "The time that a file over 1 GB takes"
id = "large-file"
```

## The review record

The top keys are `commit`, the full SHA of the reviewed commit, `reviewer`, and `verdict`.
The verdict is `Ready`, `Ready with stated limits`, or `Needs work`.

Each `[[findings]]` table is one finding. It has one value for each of these keys:

- `id`: such as `R2-F3`. `R2` is the round that first reported it, and the ID stays in later rounds
- `title`, and `severity`: `block`, `warn`, or `advise`
- `state`: `open`, `fixed`, `limit`, or `declined`. A finding without a state fails, also one of the last round (#423)
- `spec`: the spec ID of its line, and `family`: the family of `## Case matrix` that its line is in
- `commit`: the fixing commit of a `fixed` finding, and `reason`: the reason of a `declined` finding

The keys `evidence`, `command`, `consequence`, `owner`, and `fix` hold prose. No check reads them, so a word there never grades.

Each `[[rounds]]` table is one round, in order from round 1.
Its keys are `head`, `change_review`, `change_commits`, `readiness_review`, and `finish`.
The table `[order]` gives the commits `intake`, `spec`, `scenario`, and `handler` of round 1 of the current build.
The table `[first_run]` gives the `commit`, the `scenario` name, the `exit` status, and the `assertion` of the first failing run.

```toml
commit = "<full SHA>"
reviewer = "agent B"
verdict = "Ready"

[[findings]]
id = "R1-F1"
title = "A missing file names no cause"
severity = "warn"
state = "fixed"
spec = "bad-input"
family = "The input is a regular file that exists"
commit = "<full SHA>"

[[rounds]]
head = "<full SHA>"
change_review = "agent A"
change_commits = "<first>..<last>"
readiness_review = "agent B"
finish = "pending"

[order]
intake = "<SHA>"
spec = "<SHA>"
scenario = "<SHA>"
handler = "<SHA>"

[first_run]
commit = "<the Scenario commit>"
scenario = "inspect"
exit = 1
assertion = "stdout '\"lines\":3'"
```

## Markdown records

A template before COR-104-T4 wrote `readiness.md` by hand, and finish parsed its prose.
An extension with such a `readiness.md` starts a new build of records when it updates.
The Intake commit of that build comes after the last commit of a Markdown `readiness.md`.
Its Git history keeps the Markdown reviews, and finish reads none of them.
