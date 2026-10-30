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

Create the repository of a new extension from the extension template at VERSION. Replace VERSION with the release tag that `ncly --version` prints. Then follow the `AGENTS.md` of the new repository. [Extensions](https://github.com/pascalandy/nicely/blob/main/docs/north-star/contract/extensions.md) holds the rules of an extension.

```sh
uv tool run copier@9.19.1 copy --vcs-ref VERSION --defaults --data domain=DOMAIN --data name=NAME --data module=github.com/me/ncly-DOMAIN-NAME gh:pascalandy/nicely ncly-DOMAIN-NAME
```

The domain names a family of work, such as `video`, and the optional name identifies one job, such as `convert`.
Neither names the wrapped tool.
For a whole-domain extension, omit `--data name=NAME` and both `-NAME` suffixes.

## Remove an extension

`just install` puts an extension in the extensions folder: `nicely/extensions` in the data folder, such as `~/.local/share/nicely/extensions`.
`ncly doctor extensions --json` gives the `path` of each extension that it finds.

1. Move its four files out of that folder, such as to the trash: `ncly-SLUG`, `ncly-SLUG.toml`, `ncly-SLUG.locales/`, and `ncly-SLUG.skill/`. In a clone of the extension whose `just --list` shows `uninstall`, `just uninstall` does it
2. Run `ncly describe ROOT --json`. It answers `NOT_FOUND`, unless an extension that the removed one hid now takes the root
3. The removal keeps the keys and the grants. For each grant, `ncly doctor` then warns `GRANT_UNUSED`: relay its `next`, `ncly auth revoke SERVICE SLUG`, to the human
