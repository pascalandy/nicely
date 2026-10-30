---
name: nicely-extensions
description: Use to install, build, or remove an ncly extension.
---

# Nicely extensions

## Install an extension

[Release and install](https://github.com/pascalandy/nicely/blob/main/docs/north-star/contract/extensions.md#release-and-install) holds the install rule:

> A user installs a fixed version with `git clone --branch <tag> <repository>`, then `just install` in the clone, which need Go and `just`.
> For an annotated tag, Git can warn that the tag `is not a commit!`, then note a detached HEAD: the clone holds the tagged commit.
> `ncly describe <root> --json` then confirms the install.

For the repository `github.com/OWNER/REPO` at its highest release tag TAG:

```sh
git clone --branch TAG https://github.com/OWNER/REPO
cd REPO && just install
```

`git ls-remote --tags --refs --sort=-v:refname https://github.com/OWNER/REPO` lists the tags, highest first. Then run `ncly describe ROOT --json`, such as `ncly describe video convert --json`, and `ncly skill view SLUG`. The install grants no key.

## Build an extension

Create the repository of a new extension from the extension template at VERSION. Replace VERSION with the release tag that `ncly --version` prints, such as `v0.24.0`. A development build prints a pseudo-version instead, such as `v0.24.1-0.20261030091907-8355631969d4`. For a pseudo-version, use its commit as VERSION: the 12 characters after its last hyphen, such as `8355631969d4`.

```sh
uv tool run copier@9.19.1 copy --vcs-ref VERSION --defaults --data domain=DOMAIN --data name=NAME --data module=github.com/me/ncly-DOMAIN-NAME gh:pascalandy/nicely ncly-DOMAIN-NAME
```

The domain names a family of work, such as `video`, and the optional name identifies one job, such as `convert`. Neither names the wrapped tool. For a whole-domain extension, omit `--data name=NAME` and both `-NAME` suffixes.

Then build each command with the `nicely-extension` skill of the new repository, from the first command. That skill is in `.agents/skills/nicely-extension/`, and `AGENTS.md` gives the reading order and the recipes. [Extensions](https://github.com/pascalandy/nicely/blob/main/docs/north-star/contract/extensions.md) holds the rules of an extension.

## Update an extension repository

Run `just update VERSION` on a clean, committed Git tree.
Replace VERSION with the target Nicely tag, such as `v0.24.0`.

For an older repository whose update recipe holds a script, bootstrap from its root:

```sh
go -C "$(go mod download -json github.com/pascalandy/nicely@VERSION | sed -n 's/^[[:space:]]*"Dir": "\(.*\)",$/\1/p')" run ./sdk/extensionrepo/cmd/extensionrepo update VERSION "$PWD"
```

On a conflict, resolve each file and mark it with `git add`.
Remove each `.rej` file, then run the printed bootstrap command with `--continue`.
Success ends with `just check`.
[Update an extension](https://github.com/pascalandy/nicely/blob/main/docs/north-star/contract/extension-repository.md#update-an-extension) holds the recovery rules.

## Remove an extension

`just install` puts an extension in the extensions folder: `nicely/extensions` in the data folder, such as `~/.local/share/nicely/extensions`.
`ncly doctor extensions --json` gives the `path` of each extension that it finds.

1. Move its four files out of that folder, such as to the trash: `ncly-SLUG`, `ncly-SLUG.toml`, `ncly-SLUG.locales/`, and `ncly-SLUG.skill/`. In a clone of the extension whose `just --list` shows `uninstall`, `just uninstall` does it
2. Run `ncly describe ROOT --json`. It answers `NOT_FOUND`, unless an extension that the removed one hid now takes the root
3. The removal keeps the keys and the grants. For each grant, `ncly doctor` then warns `GRANT_UNUSED`: relay its `next`, `ncly auth revoke SERVICE SLUG`, to the human
