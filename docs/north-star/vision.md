# Nicely vision

This file says what Nicely is and how agents work with it.

## What Nicely is

Nicely is a curated, multilingual CLI toolbox that humans and AI agents operate equally well. Its command is `ncly`.

Pascal builds it for his own work first and publishes it for anyone. The repository is public, MIT licensed, and sends no telemetry.

Nicely runs on macOS and Linux. Omarchy, which is based on Arch, is the reference Linux. Windows is not a target.

## Agents in both directions

Agents operate `ncly`. Every command runs without a terminal, answers in JSON on request, and fails with an error code and a fix. [Output](contract/output.md) in the contract lists the few exceptions, such as `--help`. Any agent with a shell can use Nicely, so `ncly` never needs to know which agent calls it.

`ncly` also runs agents, through the `agent` extension. Its first capability runs one trusted Codex task, with model and effort settings. Additional harnesses and named profiles remain in [XT-106](../backlog/XT-106-agent.md#later). Other extensions compose through the same CLI when their own workflow needs it. [Transcript summary](../backlog/XT-102-transcript.md#later) remains deferred until a verified tools-off path exists.

Agents set `NCLY_NO_INPUT=1` and `NCLY_JSON=1`, then read the bundled `nicely` skill with `ncly skill view nicely`. They discover a command with `ncly describe`. They dry-run a command that writes or bills: the dry run checks what that invocation needs. For a recorded command, they choose a [run ID](contract/run-ids.md) before the call, so that a lost answer never causes an automatic second charge. `ncly doctor` checks the whole machine, during setup or when a dry run reports a problem that the agent cannot explain. Each extension ships a skill that teaches agents its own commands.

Every extension extends the same loop instead of adding one: its commands appear in `ncly describe`, its skill in `ncly skill list`, and its codes in its manifest. An agent that knows the loop uses a new extension without learning anything else.

| Source of truth | What it establishes |
|---|---|
| [Command declaration](core-spec.md#ncly-describe) | Installed capabilities, input constraints, result types, and supported modes |
| [Doctor](core-spec.md#ncly-doctor) | The machine's readiness, with free network checks only on request |
| [Dry run](contract/operations.md) | The invocation's plan, its readiness, and the checks that only execution can make |
| [Run inspection](core-spec.md#ncly-run) | Saved step evidence and `active`, which reports whether a process holds the run lock |
| Artifacts | The full output files, referenced by path and verified fingerprints |
| [Run ID](contract/run-ids.md) | The one run that a call names, so a repeated call verifies saved work before any further effect |

Commands with paid work or reusable partial results keep records. An agent inspects a recorded failure before choosing a recovery action. Overall exit 75 alone permits an automatic repeat of the same invocation. A saved status describes the operation's last recorded outcome; the execution lock establishes whether a process owns the run now.

Records and verified outputs retain useful work across sessions. The first recorded commands save one item and one paid step, with inspection and verified completion replay. [Explicit resume](../backlog/COR-621-resume-sdk.md#later) waits for a workflow with reusable partial work. The SDK's record support owns records, locks, and result publication; the extension decides what to reuse or execute. Nicely needs no memory daemon, workflow interpreter, or generic engine for future domains. A record does not keep a process alive, and Nicely has no daemon or scheduler.
