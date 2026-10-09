# Nicely vision

This file says what Nicely is and how agents work with it.

## What Nicely is

Nicely is a curated, multilingual CLI toolbox that humans and AI agents operate equally well. Its command is `ncly`.

Pascal builds it for his own work first and publishes it for anyone. The repository is public, MIT licensed, and sends no telemetry.

Nicely runs on macOS and Linux. Omarchy, which is based on Arch, is the reference Linux. Windows is not a target.

## Agents in both directions

Agents operate `ncly`. Every command runs without a terminal, answers in JSON on request, and fails with an error code and a fix. [contract.md](contract.md#output) lists the few exceptions, such as `--help`. Any agent with a shell can use Nicely, so `ncly` never needs to know which agent calls it.

`ncly` also runs agents, through the `agent` extension. Profiles in the config name a harness, a model, and an effort. Other extensions use agents through `ncly agent run`, as transcript does for its summary.

Agents set `NCLY_NO_INPUT=1` and `NCLY_JSON=1`, read the bundled `nicely` skill with `ncly skill view nicely`, discover one command, check readiness, prepare its invocation with dry run, then execute it. Each extension ships a skill that teaches agents its own commands.

| Source of truth | What it establishes |
|---|---|
| [Command declaration](core-spec.md#ncly-describe) | Installed capabilities, input constraints, result types, and supported modes |
| [Doctor](core-spec.md#ncly-doctor) | Local readiness, with free network checks only on request |
| [Dry run](contract.md#operations) | The invocation's plan and checks still pending before execution |
| [Run inspection](core-spec.md#ncly-run) | Saved step evidence and `active`, which reports whether a process holds the run lock |
| Artifacts | The full output files, referenced by path and verified fingerprints |
| [Explicit resume](contract.md#inspection-and-explicit-resume) | The saved outputs that current verification permits the extension to reuse |

Commands with paid work or reusable partial results keep records. An agent inspects a recorded failure before choosing a recovery action. Overall exit 75 alone permits an automatic repeat of the same invocation. A saved status describes the operation's last recorded outcome; the execution lock establishes whether a process owns the run now.

Records and verified outputs retain useful work across sessions. Explicit resume reuses that work and runs only safe remaining steps. The SDK's record support owns records, locks, and result publication; the extension decides what to reuse or execute. Nicely needs no memory daemon, workflow interpreter, or generic engine for future domains. A record does not keep a process alive, and Nicely has no background worker or scheduler.
