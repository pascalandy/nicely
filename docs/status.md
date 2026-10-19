# Current build

This page answers what the current build of `ncly` does, command by command. [README.md](../README.md#status) gives one line for each part, and [the milestones](milestones/) list the work in order. A card that changes what a user sees adds its line to README.md and its detail here. The build also establishes the JSON error contract, configuration, and English message catalog.

## Commands and extensions

The current build implements help, the version, shell completion, and `ncly describe`. `ncly describe` lists the installed commands, core's and each extension's, with their summaries. For one command, it gives the arguments, flags, effects, modes, and requirements, and the JSON Schema of its answers. It also gives each config value that the command reads, and where that value comes from.

`ncly` finds the extensions named `ncly-<domain>` in `~/.local/share/nicely/extensions/` and in `PATH`. It shows their help and completion from their manifests. It starts them with a call that it has checked against their declared arguments, flags, and flag groups. It checks their JSON answer before passing it on, and stops them and their child processes on Ctrl-C.

It bundles the `skill` extension. `ncly skill list` lists the skills in the folders that `[skill] paths` names in the config, and the skill of each installed extension. `ncly skill view <name>` prints one of them without its frontmatter, and a terminal shows it as styled Markdown. Shell completion suggests the names of the skills after `ncly skill view`. An extension's argument or flag can complete from one of its list commands.

## Doctor

`ncly doctor` checks the machine and reports on stdout. It says whether the config files parse and warns about [unknown config keys](north-star/contract/configuration.md). It lists each extension that it finds, and the one that hides another. It checks the programs that each extension requires, and names the install command of a missing or too old one. Then it runs each extension's status command, such as `ncly skill status`, which checks the skill folders, names, and files.

Last, its `auth` component gives the verdict of `ncly auth status`. It checks that the keychain answers, and that each grant still names an extension that declares its key. It also checks that each declared key is in its variable or the keychain, and granted to each extension that needs it.

## Keys

`ncly auth status` lists the keys that the installed extensions declare, without printing any key. It says whether each one comes from its environment variable, such as `DEEPGRAM_API_KEY`, or from the OS keychain. It also says which extensions hold its grant. `ncly auth login <service>` stores a key in the keychain, from a masked field that never shows it or from `--stdin`. `ncly auth logout <service>` removes it, and both ask for a typed `[y/N]` before they replace or remove a key.

`ncly auth grant <service> <domain>` lets the installed extension of a domain receive a key. A human confirms each grant with a typed `[y/N]` in a terminal. `ncly auth revoke <service> <domain>` takes the grant back. A grant holds the path where `ncly` found the extension. Another `ncly-<domain>` earlier in `PATH` therefore inherits nothing, and granting to it asks again with both paths. A Homebrew upgrade behind the same link keeps the grant.

Core passes an extension only the keys that a human granted to it at its path, from their variables or the keychain. It names each declared key that it does not pass in `NCLY_MISSING_KEYS`, with its code. The extension answers that code when it needs that key:

- `KEY_NOT_GRANTED`, with the grant command to run, for a key that no human granted
- `AUTH_MISSING` or `KEYRING_UNAVAILABLE` for a key that neither its variable nor the keychain gives

An extension that runs without a key can mark it optional in its manifest, such as a harness that falls back to its own login. `status` and doctor then pass that key when it is missing or not granted, unless another extension needs it. `ncly describe` says which keys are optional. Completion suggests the declared services after each of them, then the domains of `grant` and `revoke`.

## Forms and styles

In a terminal, a call that lacks a required value opens a form that asks only for what it lacks. The call then prints the equivalent command after `Next time:`. Elsewhere, the call fails with the full command as its hint. On a terminal, `skill list`, `auth status`, and the doctor report are styled, and a pipe gets the same text without the styles.

## agent

From a checkout, `just install` puts `ncly` in `~/.local/bin/` and the `agent` extension in `~/.local/share/nicely/extensions/`. `ncly doctor agent` checks that Codex 0.162.1 or later is installed.

`ncly agent run <task> --tools --model <id> --dry-run` checks one trusted task, the folder, and Codex. It also checks the model and effort from its flags or the `[agent]` table of the config. It then shows the invocation of Codex without starting it. It refuses with `AGENT_DEPTH_LIMIT` a call that already runs inside `[agent] max_depth` agents, 2 by default, as `NCLY_AGENT_DEPTH` counts them. It gives Codex that depth plus one.

Without `--dry-run`, it runs the task under the ID of `--run-id`, or one that it generates. It keeps a record of the run in `~/.local/state/nicely/runs/`. It saves the record before Codex starts, then the answer of Codex in `answer.txt`, before it prints that answer. When that run completed, the same call with the same `--run-id` prints that saved answer again and starts no Codex. When the call stopped before it saved its intent, the same call runs the task. A call whose run may have started a paid task without a saved answer starts no other one.

`--timeout`, 5 minutes by default, bounds the whole call. Ctrl-C, SIGTERM, and the timeout stop Codex and its child processes, and the record says what happened. A step `failed` ran no task. A step `unknown` may have billed, so the same call then starts nothing.

## transcript

`just install` also puts the `transcript` extension in `~/.local/share/nicely/extensions/`. `ncly doctor transcript` checks that `uv` and Python 3.12 or later are installed. Then `ncly transcript status` checks the cookies of the browser that `[transcript]` sets. Without a browser, it passes. With `browser = "arc"`, it warns with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` when no cookie file of Arc's default profile opens. It reads no key, starts no program, and uses no network.

`ncly transcript run youtube --url <url> --dry-run` checks the URL of one video, the output folder, and the `[transcript]` tables of the config. It then runs the transcript program through `uv` and shows its plan, with no request to YouTube or Deepgram. It reads no key and creates no record.

Without `--dry-run`, the call needs the key of `deepgram`, which a human grants to transcript. A key with a character that no request can carry, such as a line break, answers `AUTH_REJECTED` before any record. It records the run in `~/.local/state/nicely/runs/` before the download. It saves each candidate result folder in the record, then creates that folder, before the upload. The folder takes the title of the video, with `-2`, `-3`, and so on when the name is taken. Deepgram gets the audio once, and the folder receives `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json`. The call checks, hashes, and saves them in the record before it answers `output_dir` and `transcript`.

A key that Deepgram refuses gives `AUTH_REJECTED` in the item, under `RUNTIME`. Each failure keeps its cause in the item. Only a temporary failure before the result folder exits 75. A Deepgram answer lost after the upload leaves the step `unknown`, and nothing retries the upload.

The same call with the same `--run-id` returns a completed run again, without `uv`, the program, Deepgram, or a key. It checks each file first, and a file that changed or went missing answers `RESUME_UNSAFE`. A config change never moves a saved result or a pending run. Once the audio may have reached Deepgram, the same call answers `RESUME_UNSAFE` too. Another URL or `--output` under that ID answers `USAGE_INVALID`. A dry run with the ID plans `reuse`, or refuses as the call would.

Ctrl-C exits 130 and SIGTERM 143, and neither leaves a `uv`, Python, or yt-dlp process. The answer keeps the result folder. A stop after the upload leaves the step `unknown`, so the same call starts nothing, and a completion before the stop stays complete. A second call under a busy run ID exits 75 and starts nothing. Two runs of one title write two folders.

## Runs

`ncly run list` lists the saved runs, newest first, by command path, item key, or start time, without an index. It marks the ones that a process holds now. `ncly run view <run-id>` shows one record, its steps, artifacts, and errors, and whether a process holds it, without starting any work.
