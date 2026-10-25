---
id: XT-105
extension: "skill"
state: "Shipped by M04, bundled. Later: skill pack, index, run, MCP server, key checks for skills"
---

# XT-105 skill

## Demo

`ncly skill list` and `ncly skill view` expose skills, including the bundled `nicely` skill.

## Later

| Idea | What it would do | Revisit when |
|---|---|---|
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |

| Idea | What it would do | Revisit when |
|---|---|---|
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |

| Idea | What it would do | Revisit when |
|---|---|---|
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |

| Idea | What it would do | Revisit when |
|---|---|---|
| MCP skills server | Serves skills over the MCP Skills extension, SEP-2640, from the `skill` domain package rather than from CLI output. Per-file manifests join `view --json` as optional fields, never `list`, which agents read whole. Research in [#10](https://github.com/pascalandy/nicely/issues/10) | A host Pascal uses supports SEP-2640 and cannot read skills from disk. `ncly skill pack` answers the same need |

| Idea | What it would do | Revisit when |
|---|---|---|
| Key checks for skills | `ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation` | A selected tap slice carries skills with prerequisite manifests |

See the related MCP research in [#10](https://github.com/pascalandy/nicely/issues/10).
