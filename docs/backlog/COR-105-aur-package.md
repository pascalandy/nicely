---
id: COR-105
core: "AUR package"
state: "Ready. T1 waits on Pascal"
---

# COR-105 AUR package

## Demo

`yay -S ncly-bin` on Omarchy installs the released `ncly`.

## Cards

| Card | Title | Owner | Depends on | Status |
|---|---|---|---|---|
| COR-105-T1 | AUR account and release key | Pascal | - | todo |
| COR-105-T2 | Publish ncly-bin with each release | agent | COR-105-T1 | todo |

### COR-105-T1 AUR account and release key

[#448](https://github.com/pascalandy/nicely/issues/448) holds Pascal's procedure for the AUR account and a new release key.

### COR-105-T2 Publish ncly-bin with each release

- **Read:** [release-nicely](../../.agents/skills/release-nicely/SKILL.md), [.goreleaser.yaml](../../.goreleaser.yaml)
- **Proves:** `.goreleaser.yaml`

[#449](https://github.com/pascalandy/nicely/issues/449) holds the publication and installation checks.
GoReleaser builds the package that `just release-check` verifies.
Remove `skip_upload: true` from the `aurs` section of `.goreleaser.yaml`.
Update the release-nicely skill to say that the workflow publishes the AUR package.
The installed release completes `ncly <Tab>` in bash on Omarchy.
