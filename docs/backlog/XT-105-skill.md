---
id: XT-105
extension: "skill"
state: "Shipped by M04, bundled. Later: skill pack, index, run, MCP server, key checks for skills"
---

# XT-105 skill

## Demo

`ncly skill list` and `ncly skill view` expose skills, including the bundled `nicely` skill.

## Later

### `ncly skill pack`

Zips a skill for upload to the Claude apps.

Revisit condition: A skill must work in an app that cannot read GitHub.

### `ncly skill index`

Generates the remote skills index that `just remote-skills` builds today.

Revisit condition: The skills repository moves its tooling into Nicely.

### `ncly skill run`

Runs a script of a skill with its prerequisites checked.

Revisit condition: Agents struggle to run skill scripts by path.

### MCP skills server

Serves skills over the MCP Skills extension, SEP-2640, from the `skill` domain package rather than from CLI output. Per-file manifests join `view --json` as optional fields, never `list`, which agents read whole. Research in [#10](https://github.com/pascalandy/nicely/issues/10).

Revisit condition: A host Pascal uses supports SEP-2640 and cannot read skills from disk. `ncly skill pack` answers the same need.

### Key checks for skills

`ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation`.

Revisit condition: A selected tap slice carries skills with prerequisite manifests.

See the related MCP research in [#10](https://github.com/pascalandy/nicely/issues/10).
