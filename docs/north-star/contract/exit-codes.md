# Exit codes

| Code | Meaning | Safe to rerun |
|---|---|---|
| `0` | Success | |
| `1` | Runtime failure. Work may be partial or already billed. | No. Never rerun automatically. |
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | Fix the call. Inspect any returned `run_id` before continuing recorded work |
| `75` | Temporary failure that meets every condition in Retry safety | Yes, with the same command |
| `78` | A human must act: a key, a prerequisite, a config fix, or a step at a terminal | Relay the hint. Inspect any returned `run_id` before continuing recorded work |
| `130` | Interrupted by Ctrl-C | |
| `143` | Terminated by SIGTERM | |

Codes 124 to 127, and every code from 128 up other than 130 and 143, stay reserved for the shell.

Only the overall process exit 75 permits an automatic repeat of the same invocation. An item error describes that item's cause, not the safety of repeating the whole command. A caller retains the complete answer and `run_id` before extracting artifact paths. After fixing a recorded failure, it inspects the run before making the same call with the same `--run-id`. The run-ID rules decide whether that call can return verified work, safely start an untouched step, or must refuse. `ncly run resume` remains in [COR-621](../../backlog/COR-621-resume-sdk.md#later); a new run ID is never an automatic recovery from an unknown paid outcome.
