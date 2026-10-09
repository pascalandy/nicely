# D003 Put the domain first and keep one verb per action

Decided 2026-10-05. Commands follow the grammar in [contract.md](../contract.md#usage): `ncly <domain> [<resource>] <verb>`, with singular nouns and one verb per action across domains.

**Why.** An agent explores one domain at a time, so `ncly video --help` costs fewer tokens than a flat list of fifty verbs. Options stay specific to their domain. Names become predictable: after `ncly image convert`, an agent guesses `ncly video convert`. The verbs are those of `gh`, which agents already know.

**Rejected.** The verb-first form, such as `ncly convert`. Keeping the interface of a script that joins Nicely, such as `transcript list prompts`, because one mixed shape breaks the guess.
