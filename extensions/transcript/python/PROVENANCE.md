# Provenance of the transcript engine

This file says where the engine comes from and how it differs from its source.

## Source

- Repository: https://github.com/pascalandy/skills
- Folder: `authoring/andy/transcript/scripts/`
- Commit: `0c9c4a223e114a8bc3c4293544202d40a5906e8f`
- License: MIT, as [MIT notice](#mit-notice) shows

| Engine file | Source |
|---|---|
| `transcript.py` | The PEP 723 header of `transcript.py` |
| `runtime.py` | `RunBudget`, `run_child`, `retry_request`, and their helpers, from `transcript.py` |
| `youtube.py` | `run_ytdlp`, `download_audio`, `clean_title`, and the URL checks, from `transcript.py` |
| `deepgram.py` | `transcribe_audio` and `parse_transcript`, from `transcript.py` |
| `outputs.py` | `save_outputs` and `write_text_atomic`, from `transcript.py` |
| `ytdlp_arc.py` | `ytdlp_arc.py` |
| `tests/` | The tests of that code, from `tests/test_transcript.py` and `tests/test_contract.py` |
| `pytest.ini` | The `pytest.ini` at the root of the source repository |

## Removed features

The tests of each removed feature left with it.

- Summary: providers, profiles, models, efforts, and the Claude and pi runners
- Prompts, and the commands that list prompts, profiles, and models
- Zoom: meeting discovery, export folders, and file names
- Queues of several URLs
- Finder and the glow preview
- The progress display, and `rich` in the header
- The CLI: commands, help, JSON answers, exit codes, signals, dry run, and doctor
- `validate_env`, which read the Deepgram key from the chezmoi keyring or `DEEPGRAM_API_KEY`
- `youtube_smoke.py`, its paid check, and the `arc-required` mode of `run_ytdlp` that only it used
- Result folders: `make_unique_dir`, `ResultFolder`, and its `meta.txt` writer
- `DeepgramError` and the failure messages of the CLI, which only the CLI read

## Intentional changes

- `transcript.py` keeps only its header and a docstring. Its code moves to the files of the table above
- Each file stays under 10,000 bytes, as `just limits` requires. Each test file follows the file it tests
- `run_child` drops `input`, `capture`, `check`, and `cwd`, which only removed code used. Its child always leads a process group
- `save_outputs` drops `base_stem`, which only Zoom used
- `SUMMARY_ATTEMPT_TIMEOUT` becomes `YTDLP_ATTEMPT_TIMEOUT`, since only the yt-dlp attempts read it now
- `install_arc_cookie_adapter` takes a `ModuleType` and sets the patched function with `setattr`
- Comments that named removed features are dropped or reworded
- `conftest.py` drops the exit-code declarations of the CLI and the fixture that restored its logger
- `pytest.ini` moves to the engine folder and replaces the import path that `conftest.py` set
- Two process tests call `run_child` with a timeout, instead of a signal to the CLI. They come from `test_a_signal_also_stops_what_a_child_started` and `test_a_signal_exits_without_a_traceback_and_stops_children_gently`
- `test_all_external_pre_summary_calls_receive_bounded_timeouts` becomes `test_every_external_call_receives_a_bounded_timeout`
- `test_one_recorded_response_saves_the_golden_files` is new, with its fixtures and goldens

## The nine pyright errors

Pyright 1.1.414 reports these errors at `0c9c4a2`, with pytest installed beside the runtime packages.

| Source location | Error | Outcome |
|---|---|---|
| `transcript.py:2111` | `parse_args` overrides `ArgumentParser.parse_args` without `namespace` | Removed with the CLI |
| `ytdlp_arc.py:80` | yt-dlp's `cookies` module does not match the `ChromiumCookiesModule` protocol | Fixed: the adapter takes a `ModuleType` |
| `tests/test_contract.py:953` | `fake_queue` gets a `dict[str, HTTPStatusError]` | Removed with queues |
| `tests/test_transcript.py:441` | A `SimpleNamespace` stands in for a rich `Task` | Removed with the progress display |
| `tests/test_transcript.py:1410` | `downloaded` passes a `str` as a `YtDlpMethod` | Removed with the pipeline tests that used it |
| `tests/test_transcript.py:1489` | `request.stream` can be an `AsyncByteStream` | Fixed: the test asserts a `SyncByteStream` first |
| `tests/test_transcript.py:2116`, `:2139`, `:2157` | A `SimpleNamespace` stands in for `ChromiumCookiesModule` | Fixed: the tests pass a `ModuleType` |

## Goldens

- `tests/fixtures/audio.webm` holds 2,272 bytes of placeholder text, since no code decodes the audio
- `tests/fixtures/deepgram_response.json` follows the documented shape of a Deepgram `/v1/listen` answer with paragraphs. It was written by hand, because no test may send a paid request
- The source program at `0c9c4a2` wrote `tests/goldens/raw_*` from those two fixtures, through `tests/goldens/make_goldens.py`. Its docstring gives the command

## MIT notice

```text
MIT License

Copyright (c) 2026 Pascal Andy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
