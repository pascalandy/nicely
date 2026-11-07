---
id: XT-107
extension: "image shrink"
state: "Shipped v0.1.0 in pascalandy/ncly-image-shrink. Epic #494 keeps the UAT and Nicely issues #496, #497, and #498 open"
---

# XT-107 image shrink

## Demo

`ncly image shrink photo.jpg` writes a smaller `photo.shrunk.jpg` beside the input, without its location, after a dry run.

## Scope

The extension ships in its own public repository, `pascalandy/ncly-image-shrink`, with its own versions.
[Epic 19, #494](https://github.com/pascalandy/nicely/issues/494) follows the work, and the milestone `XT-107 image shrink` holds its issues.
This is the first slice of [video and image jobs](XT-602-video-image-jobs.md).

### The command

- One job: make one image smaller to share it. Domain `image`, name `shrink`, slug `image-shrink`, as [Extension identity](../north-star/contract/extension-identity.md) defines
- One JPEG, PNG, or WebP file in. One file out, in the same format, at `<stem>.shrunk.<ext>` beside the input. The input never changes
- Success means that the output has fewer bytes than the input. Otherwise the command publishes nothing and fails with its own error code. An optimized PNG below the size limit is such a case
- The longest side becomes at most 2048 pixels by default, and one flag sets another value. The command never enlarges an image
- JPEG and WebP use a lossy quality setting. PNG stays lossless
- The output applies the orientation of the input and keeps its color profile
- The output removes the descriptive metadata: location, author, camera, and capture date. This is the purpose of a copy to share
- `--dry-run` previews the work, and `--force` replaces an existing output, as `ncly video convert` does
- No key, no network, and no cost

The intake chooses the program that does the work, such as libvips or ImageMagick.
The program must install with Homebrew on macOS and with pacman on Omarchy.
The manifest declares it, and `ncly doctor` names its install command.

### Template proof

This item also proves the template on a new extension.
An agent new to this work builds it, up to "Done when" in the template's [AGENTS.md](../../template/AGENTS.md.jinja).
The video-convert adoption proved the loop on an extension that already existed.

- An agent that did not read `ncly-video-convert` builds the command, from the template of the latest Nicely release
- The [`nicely-extension` skill](../../template/.agents/skills/nicely-extension/SKILL.md) of the new repository drives the work: intake, spec, a scenario that fails first, rounds, independent review, readiness review, and `just finish`
- When the template, the contract, or the SDK stops the agent, it opens an issue in this repository and links it in Epic 19. The issue has the label `2-type:bug` or `2-type:task`
- Nobody edits Nicely by hand to unblock the build
- The pull request names the author session and the reviewer session. [D069](../north-star/decisions/D069-recorded-sessions.md) designs their capture, and #451 will enforce it

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| XT-107-T1 | Create the repository from the template | agent | - | done |
| XT-107-T2 | Build ncly image shrink through the loop | agent | XT-107-T1 | done |
| XT-107-T3 | Release v0.1.0 | agent | XT-107-T2 | done |

### XT-107-T1 Create the repository from the template

- **Read:** [nicely-extensions](../../extensions/skill/nicely-extensions/SKILL.md), [Extension repository](../north-star/contract/extension-repository.md)
- **Proves:** just check and the hosted checks, in ncly-image-shrink

Create the public repository `pascalandy/ncly-image-shrink` with the Copier command of the `nicely-extensions` skill, at the tag that `ncly --version` prints.
Pass `domain=image`, `name=shrink`, and `module=github.com/pascalandy/ncly-image-shrink`.
Run `go mod tidy`, commit, and push.
`just check` passes on the rendered repository, and its GitHub workflow passes on Linux and macOS.

Done in the [initial push 3af2232](https://github.com/pascalandy/ncly-image-shrink/commit/3af22321e95dac6558baac3fec6db8f54b11a679), with [Linux and macOS checks](https://github.com/pascalandy/ncly-image-shrink/actions/runs/42327801797) passed.

### XT-107-T2 Build ncly image shrink through the loop

- **Read:** `AGENTS.md` and the `nicely-extension` skill of the new repository, [What a great extension does](../north-star/contract/extensions.md#what-a-great-extension-does)
- **Proves:** intake.toml, spec.md, review.toml, and just finish, in ncly-image-shrink

Build the command of [The command](#the-command) in one pull request, in at most 5 rounds, as the `nicely-extension` skill says.
Its acceptance cases use images that the tests make.
One of them holds a location that the output must not keep.
One of them cannot get smaller, and the command must refuse it without an output.
The card is done when `just finish` passes and the pull request merges.

Done in [PR #1](https://github.com/pascalandy/ncly-image-shrink/pull/1), merged at `f83de69`, with Ready in round 4.

### XT-107-T3 Release v0.1.0

- **Read:** the [README](../../template/README.md.jinja) and the [justfile](../../template/justfile) that the template renders, and [delivery.md](../../template/.agents/skills/nicely-extension/delivery.md)
- **Proves:** the release receipt, in ncly-image-shrink

Run `just signoff`, then `just release v0.1.0`.
Install the tag on om1 and on the Mini as the README says.
Install the program that the manifest declares on the Mini only with Pascal's consent.

Released [v0.1.0](https://github.com/pascalandy/ncly-image-shrink/tree/v0.1.0).
See the [release receipt](https://github.com/pascalandy/ncly-image-shrink/pull/1#issuecomment-6703974395) and [om1 and Mini runs](https://github.com/pascalandy/ncly-image-shrink/pull/1#issuecomment-6704004143).

## After

- A separate UAT issue: Pascal tells a fresh agent "Use Nicely to shrink the image <path>", then shares the result
- The cost of the build: rounds, findings, and the issues opened against Nicely. COR-104 After reviews the cost of the gates

## Later

- HEIC input, batches of images, and folders
- Other image jobs, such as crop or convert, each as its own named extension in the `image` domain
