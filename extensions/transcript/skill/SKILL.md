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
- The call stops after `--timeout`, `570s` by default
