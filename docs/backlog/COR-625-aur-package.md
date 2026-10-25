---
id: COR-625
core: "AUR package"
state: "Idea. Revisit when AUR registration reopens"
---

# COR-625 AUR package

Publishes `ncly-bin` to the AUR, so that `yay -S ncly-bin` installs `ncly` and `ncly <Tab>` completes in bash on Omarchy. It needs an AUR account with the public key `nicely-release-aur`. GoReleaser builds the package that `just release-check` verifies, and `skip_upload: true` in the `aurs` section of `.goreleaser.yaml` keeps it off the AUR. To ship it, delete that line, and say in step 3 of the `release-nicely` skill that the workflow publishes the AUR package.
