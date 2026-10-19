# Everyday domains

Pascal picked neither candidate in [M17](../M17-inventory.md) on 2026-10-19.
Both wait for the [extension template](../M18-extension-template.md).

## Markdown view

Today, `just sk-read <name>` in Pascal's SKILLS_MONO repository prints a skill's raw Markdown.
The candidate renders one Markdown file in the terminal.
Revisit as a default extension that wraps Glow once the extension template milestone is done.

## Screen-recording conversion

`video-archive` in Pascal's dotfiles covers this task today:

```
video-archive convert --input ./clip.mov -o ./clip.mp4
```

The candidate produces one converted file.
Revisit once the extension template exists and Pascal picks this task as an extension.

## Other deferred scope

Parked from M17: exhaustive inventories of two machines, histories and packages, and the rule
that every writing command supports files, folders, globs, batches, progress, records, and
resume. A safe one-file transformation does not inherit all of those capabilities.

Writing commands still need dry run, destination validation, overwrite consent, and truthful retry behavior for their actual effects.

The unselected video and image suites of M19 and M20 are removed from the delivery queue.
Keep their ideas here: video encode, cut, and extract audio; image encode, resize, and compress.
Revisit one operation when a concrete daily job warrants it, and write only its contract and
one fixture. Neither domain implies that the whole command suite ships together.
