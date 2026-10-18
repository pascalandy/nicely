# Usage

```
ncly [global flags] <domain> [<resource>] <verb> [arguments] [flags]
ncly [global flags] <command> [arguments] [flags]
```

A domain is a noun, such as `skill`. A resource is an optional second noun inside a domain; each extension declares whether its commands need one. Nouns are singular. A standalone command, such as `doctor`, has no domain.

Each action has one verb in every domain. Reuse a verb from this table when it fits. A milestone that needs a new verb adds it here.

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

`ncly --help` lists domains and commands with one line each. `ncly <domain> --help` lists the actions of that domain, and `ncly help <domain>` prints the same text. Every help page ends with two to five examples. From [M04](../../milestones/M04-skill.md), the first example of `ncly --help` is `ncly skill view nicely`, so an agent that starts with the help finds the skill that teaches it the rest. `ncly` and `ncly <domain>` without a verb print their help on stdout and exit 0.

A usage line shows a required argument as `<name>`, an optional one as `[name]`, and one that takes every remaining word as `<name...>` or `[name...]`. It then shows each required flag outside a group, such as `--store <name>`, and each group of flags, in parentheses when the group needs one of them, such as `(--latest | --path <folder>)`, and in brackets otherwise, such as `[--email | --sms]`. The help of a command that writes, uses the network, or may bill says so in one line under its summary, and names `--dry-run` when the command supports it. A command whose declaration omits its effects names all three, as `ncly describe` reports them. Both come from the command's declaration, as `ncly describe` does, so help and discovery never disagree.
