---
name: transcript
description: Use when transcribing one YouTube video with Deepgram through ncly transcript.
---

# transcript

`ncly transcript run youtube` downloads the audio of one YouTube video and sends it to Deepgram, which bills the call. It runs a Python program through `uv`, so `uv` and Python 3.12 or later must be installed.

- Give one video with `--url`, such as `https://www.youtube.com/watch?v=<id>`, `https://youtu.be/<id>`, or `https://www.youtube.com/shorts/<id>`. A playlist page answers `USAGE_INVALID`, and a watch URL with playlist parameters still selects only its video
- Choose the folder that receives the result folder with `--output`, or with `output` in the `[transcript.youtube]` table of the config. That folder must exist. Without either, the call uses the current folder
- Check the call with `ncly transcript run youtube --url "<url>" --dry-run --json` first. It reads no key, downloads nothing, sends nothing to Deepgram, and writes no user file. It may prepare the cache of `uv`, which can use the network. Its `plan` names the output folder, and its `pending_checks` holds `auth.deepgram`, which only execution checks
- When the dry run answers `PREREQ_MISSING`, relay the hint: a human installs `uv` or Python. When it answers `RUNTIME` because the program stopped before its plan, check the network and run the dry run again with `--verbose`
- Run it with a run ID that you choose, and keep the whole answer: `ncly transcript run youtube --url "<url>" --run-id <id> --json`. A human must grant the key of `deepgram` to transcript first. On `KEY_NOT_GRANTED`, `AUTH_MISSING`, or `KEYRING_UNAVAILABLE`, relay `hint` and `next` to the human
- On success, `results[0].output_dir` is the new result folder, and `results[0].transcript` is its `raw_transcript.txt`, beside `raw_sentences.txt` and `raw_transcript.json`
- A failed item keeps its cause in `results[0].errors`, and the answer's own code is the verdict of the call. `AUTH_REJECTED` means that Deepgram refused the key, or that the key holds a character that no request can carry. Relay its hint to the human, who replaces the key
- Exit 1 never permits a repeat: read the run with `ncly run view <id> --json` first. The same call with the same `--run-id` returns a completed run again without a charge, and answers `RESUME_UNSAFE` once the audio may have reached Deepgram. Only a human chooses a new run ID, which may bill again
- Exit 75 with `TEMPORARY` permits the same call again
- A dry run with the `--run-id` of a saved run decides as the call would. A completed run plans `reuse`, with its `output_dir` and no effect, and a run that cannot continue answers the same refusal
- `USAGE_INVALID` under a saved `--run-id` means that the ID names another call, with another URL or `--output`. `RESUME_UNSAFE` also answers a saved file that changed or went missing: relay it to the human, who inspects the folder
- The call stops after `--timeout`, `570s` by default
