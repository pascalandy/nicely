# UAT ledger

This file answers one question: which UATs ran, with which verdict, and what stays open? Add one row per UAT, newest first. The issue holds the detail.

| UAT | Subject | Template | Verdict | Date | Follow-ups |
|---|---|---|---|---|---|
| [#346](https://github.com/pascalandy/nicely/issues/346) | video-convert v0.1.2 on a real video, from a fresh agent | v0.23.1 | accepted with reservations | 2026-10-28 | [#350](https://github.com/pascalandy/nicely/issues/350) closed, [#351](https://github.com/pascalandy/nicely/issues/351), [#352](https://github.com/pascalandy/nicely/issues/352) |
| [#195](https://github.com/pascalandy/nicely/issues/195) | Forms and grants of M07 and M08 | n/a | pending | | |

The Template column gives `_commit` from the `.copier-answers.yml` of the extension.

## Lessons for the template

| UAT | Lesson | Issue |
|---|---|---|
| #346 | The dry run accepted a fake file that the run refused with `RUNTIME` | #350 |
| #346 | The wrapped program accepts a still image that the domain should refuse | #351 |
| #346 | The skill does not say where the error goes with `--json` | #352 |
| #346 | A fresh agent found Nicely only after its `AGENTS.md` named `ncly skill view nicely` | #346 |
