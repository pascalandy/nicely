# M16 Transcript finish

Status: planned
Version: v0.16.0

## Demo

`ncly transcript` handles Zoom, prompts, and `--open` with a spinner, and a fresh agent session uses it with only the `nicely` skill.

## Scope

The rest of [extensions/transcript/spec.md](../../extensions/transcript/spec.md): Zoom, prompts, `--open`, the doctor checks, browser cookies, the spinner in the look Pascal picked in M07, and the transcript skill.

## Open questions

Settle each one in [extensions/transcript/spec.md](../../extensions/transcript/spec.md), then delete this section.

- How transcript's checks join `ncly doctor transcript`, through the mechanism that M05 settles
- What the transcript skill says beyond the `nicely` skill, such as the lines about repeating a transcription and about `summary run`

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| M16-T1 | Zoom | agent | — | todo |
| M16-T2 | Prompts and open | agent | — | todo |
| M16-T3 | Doctor and cookies | agent | — | todo |
| M16-T4 | Spinner and transcript skill | agent | M16-T1, M16-T2 | todo |
| M16-T5 | Fresh-session check | agent | M16-T3, M16-T4 | todo |

### M16-T1 Zoom

- **Read:** [ncly transcript](../../extensions/transcript/spec.md#ncly-transcript), [Output, prompts, and cookies](../../extensions/transcript/spec.md#output-prompts-and-cookies)
- **Proves:** `testdata/script/transcript_zoom.txtar`

`ncly transcript run zoom` with `--latest` or `--path`. Python selects the audio once and Go records its fingerprint. A meeting folder without audio is `NOT_FOUND`, and the folder's name becomes the title.

### M16-T2 Prompts and open

- **Read:** [Output, prompts, and cookies](../../extensions/transcript/spec.md#output-prompts-and-cookies)
- **Proves:** `testdata/script/transcript_prompt.txtar`, `testdata/script/transcript_open.txtar`

`ncly transcript prompt list` with the precedence of `prompt_paths` over the bundled prompts. `--open` opens each result folder as it appears, never during a dry run. Completion suggests profiles and prompts.

### M16-T3 Doctor and cookies

- **Read:** [Output, prompts, and cookies](../../extensions/transcript/spec.md#output-prompts-and-cookies), [ncly doctor](../north-star/cli-spec.md#ncly-doctor)
- **Proves:** `testdata/script/transcript_doctor.txtar`, `testdata/script/transcript_cookies.txtar`

`ncly doctor transcript` takes over the checks of the transcript CLI's `doctor --source`: `uv`, Python 3.12 or later with its install command, and the browser cookies. `--live` checks the key against a free Deepgram endpoint. Unreadable cookies fall back to anonymous access with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED`, once in a batch's top-level `warnings`.

### M16-T4 Spinner and transcript skill

- **Read:** [ncly transcript](../../extensions/transcript/spec.md#ncly-transcript), [Output](../north-star/cli-spec.md#output)
- **Proves:** `testdata/script/transcript_spinner.txtar`, `extensions/transcript/SKILL.md`

A run shows one spinner on a terminal, in the picked look, and none in JSON or non-interactive mode. Write the transcript skill.

### M16-T5 Fresh-session check

- **Read:** [ncly skill](../../extensions/skill/spec.md#ncly-skill), [Inspection and explicit resume](../north-star/cli-spec.md#inspection-and-explicit-resume)
- **Proves:** the session's transcript, attached to the pull request

An agent in a fresh session, given only the line that points to `ncly skill view nicely`, discovers a command, plans it, inspects a recorded failure without full transcripts or logs, then dry-runs and explicitly resumes supported work without repeating a paid step. Fix the skills until it does.

## After this milestone

These steps wait on Pascal and block no card. Pascal decides when.

- Pascal moves his key once per machine, `chezmoi secret keyring get --service=deepgram --user=api_key | ncly auth login deepgram --stdin`, grants it to transcript, then runs the paid end-to-end check from the transcript README against its listed test video, on macOS, from a local session on the Mac. Omarchy waits for [#23](https://github.com/pascalandy/nicely/issues/23)
- In `pascalandy/skills`, the `transcript` skill points to `ncly`, the `transcript` and `transcript-cli` recipes retire, `verify-transcript` checks `ncly`, and the Python copy there is deleted
