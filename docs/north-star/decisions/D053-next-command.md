# D053 Give a problem its next command as words

Decided 2026-10-12. A problem in `errors` or `warnings` may add `next`: the command that its hint names, as an array of words, such as `["ncly","skill","list"]`, when that command runs as is. The hint stays the text for a human. The exit code still decides who acts, so `next` never permits a repeat that the exit code forbids. [Output](../contract.md#output) holds the rules.

**Why.** A hint can wrap a command in a sentence, such as ``Run `ncly auth login deepgram` in a terminal``, and its words around the command are translated. An agent that copies the command out of it can misquote a path or a URL, and a script cannot find the command at all. An array of words runs without a shell, so nothing needs quoting. The code tells an agent what went wrong, the exit code who acts, and `next` what to run, so an agent never parses the message or the hint.

**Rejected.** A list of actions, each with a kind and its possible effects, because the exit code already says who acts, and `ncly describe` and a dry run already give a command's effects. A shell string, which needs quoting rules. Parsing the hint.
