# Acceptance cases

This part says how a scenario declares an acceptance case, and what the test kit checks, as COR-104-T5 of [D068](../decisions/D068-verification-map.md) decides.
The extension keeps the facts of its domain: its made inputs, its reader, its references, and its bad variants.
The kit runs the reader, compares, and searches.

## A case

A case is a scenario of `testdata/script/`. It runs the command, then calls `reads`:

```
# spec: header
ncly notes export header.md --json
reads -case 'Keeps data that a reader needs' -private 'Ana Lima' header.txt reference.txt bad.txt sh show.sh

-- header.md --
encoding: iso-8859-1
author: Ana Lima

café
```

The archive also holds `show.sh`, `reference.txt`, and `bad.txt`. The made input is the file `header.md`, in ISO-8859-1.

`reads [-case <case>] [-limit <id>] [-private <text>]... <output> <reference> <bad> <reader> [<arg>...]` takes these parts:

- **Output**: the primary output that the command wrote. When the answer is the output, `cp stdout answer.json` keeps it in a file first
- **Reference**: the file that holds what the reader must print, byte for byte. The author writes it from the made input, without the extension
- **Bad**: a bad variant of the output, which a handler with the bug writes. For a block that a reader needs, it keeps the raw block
- **Reader**: the program that reads the output, and its arguments. The word `{}` stands for the path of the file that it reads. Without that word, the kit adds the path as the last argument
- `-private`: a text of the made input that the output must not hold. Give it once for each text
- `-case`: the case of the Output group of [Generic cases](generic-cases.md) that the case proves, by its name

Output, Reference, and Bad are paths in the folder of the scenario.
The scenario makes each file that the case reads.
A made input is a file of its archive, or a program of the scenario writes it, such as a generator of a video.
The tests never copy a file of the user.

## The reader answer

The kit gives one reader, `answer <key>...`. It reads one JSON object, such as an answer.
It prints the line `<key>: <value>` for each key, with the value in compact JSON.
A key that the object lacks gives the line `<key>: missing`.
The kit compares the whole value, so an answer of 30 lines never passes for 3. At the parent of 86d24e83, a prefix match of `acceptance.sh` passed it.

## What reads checks

1. A call of `ncly` or `ncly-<slug>` ran before it in the scenario.
2. Output, Reference, and Bad are paths in the folder of the scenario.
3. The reader is not `ncly`, `ncly-<slug>`, or a program in `requires.bins` of a command of the manifest.
4. The reader exits 0 on the output, and prints the reference.
5. No private text is in the output, or in the text of the reader.
6. The bad variant fails. With a private text, the bad variant holds it. Without one, the reader fails on the bad variant, or prints other text.
7. The cases `Keeps data that a reader needs` and `Keeps private data of the input` give private text or a validated `-limit`.
8. The case of `-case` is an Output case, and `## Case matrix` of `spec.md` names it.

Step 6 proves that the reader, the reference, and the search can refuse a bad output.
For #425, a handler that keeps the raw block fails, because its output holds the private text.

## Output cases

After all scenarios, the kit reads the cells Source of `## Case matrix` in `spec.md`.
A cell names an Output case as `Output: <case>`, such as `Output: Keeps data that a reader needs`.
Each Output case that the matrix names needs a `reads -case` of that case that passed.
`Appears during the run, without --force` needs none, because the obligation `appears` of [Publication declaration](publication.md) proves it.
`just scenario` skips this report, as it skips coverage.

Before the review of round 1, the matrix names each family, and the round ends with `just check`.
So the case of each Output case that the matrix names exists before that review.
The reviewer judges these facts:

- The matrix names each Output case that applies
- The reader and the reference share no bug of the handler
- The made input uses a value that is not the default

## Accepted privacy limits

An accepted privacy limit can retain metadata that the command does not remove.
Its case still needs a passing independent reader, the complete reference, and a bad variant that fails.

1. Declare the limit under **Limits.**, with its ID, state `limit`, description, and the user's acceptance record.
   Use the columns ID, State, Limit, and Accepted, in that order. Empty cells and placeholder text supply no acceptance.
   The Limit cell also declares each affected `Output: <case>`, separated by semicolons before its description.
2. Name that ID in backticks in Accepted for each matrix row that names the applicable Output case.
   Source uses the complete `Output: <case>` binding. Separate multiple bindings with semicolons.
3. Use `reads -case '<case>' -limit <id> <output> <reference> <bad> <reader> [<arg>...]`.
4. Make the reader and reference check each retained field and the input's nondefault display declaration.
5. Keep `-private` for text that must stay absent. Every such search remains strict, including with `-limit`.

Only the two privacy Output cases accept `-limit`.
The ID must identify exactly one accepted limit whose Limit cell names the case, and apply in every matrix row that names it.
An unknown ID, missing acceptance, wrong declared case, unbound matrix row, or ambiguous binding fails.
The reviewer rejects a description unrelated to the declared case, and checks that the reference covers every retained field.
A declaration alone proves no case and grants no coverage exemption.
