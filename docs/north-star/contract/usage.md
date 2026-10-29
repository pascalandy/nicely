# Usage

```
ncly [global flags] <domain> [<resource>] <verb> [arguments] [flags]
ncly [global flags] <command> [arguments] [flags]
```

A domain is a noun, such as `skill`. A resource is an optional second noun inside a domain; each extension declares whether its commands need one. Nouns are singular. A standalone command, such as `doctor`, has no domain.
Named extensions add jobs through [Extension identity](extension-identity.md), such as `ncly video convert <file>`.
A job name can be the action itself or group further commands beneath its root.

Each action has one verb in every domain. Reuse a verb from this table when it fits.

| Verb | Meaning |
|---|---|
| `list` | Lists items |
| `view` | Shows one item |
| `add` | Adds an item |
| `remove` | Removes an item |
| `run` | Runs a task |
| `sync` | Brings local copies up to date with their source |
| `login`, `logout` | Stores or removes a key |
| `grant`, `revoke` | Lets an extension receive a key, or stops it |
| `status` | Reports the current state |
| `resume` | Continues a recorded run after checking which steps are safe |

`ncly --help` lists domains and commands with one line each.
Its extensions section lists whole-domain extensions and shared domain groups, including groups whose describe source is core.
`ncly <domain> --help` lists the actions of that domain, and `ncly help <domain>` prints the same text. Every help page ends with two to five examples. The first example of `ncly --help` is `ncly skill view nicely`, so an agent that starts with the help finds the skill that teaches it the rest. `ncly` and `ncly <domain>` without a verb print their help on stdout and exit 0.

A word that names no command under its parent fails with `USAGE_INVALID`. When a command of that parent is close to the word, the message suggests it, such as `Did you mean "describe"?`. The hint and `next` then open its help, such as `ncly describe --help`. Cobra's suggestions find the close commands: a name that starts with the word, or that two edits or fewer turn into the word, ignoring case. Then:

- The first close command in alphabetical order wins
- A hidden command, such as `__complete`, and `help` are never suggested, as Cobra skips them
- A word that equals a command name, as an operand after `--` can, gets no suggestion
- Without a close command, the hint and `next` open the help of the parent

A usage line shows a required argument as `<name>`, an optional one as `[name]`, and one that takes every remaining word as `<name...>` or `[name...]`. It then shows each required flag outside a group, such as `--store <name>`, and each group of flags, in parentheses when the group needs one of them, such as `(--latest | --path <folder>)`, and in brackets otherwise, such as `[--email | --sms]`. The help of a command that writes, uses the network, or may bill says so in one line under its summary, and names `--dry-run` when the command supports it. A command whose declaration omits its effects names all three, as `ncly describe` reports them. Both come from the command's declaration, as `ncly describe` does, so help and discovery never disagree.
