# UAT ledger

This file answers one question: which UATs ran, with which verdict and measurements, and what stays open? Add one row per UAT, newest first. The issue holds the detail.

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
| #346 | A fresh agent found `ncly` only after a detour. One `AGENTS.md` line now names `ncly skill view nicely`, and a fresh agent must retest it | #346 |

## Economy measurements

Add one row per UAT that measured the common commands of the [budgets](north-star.md#budgets), newest first. Each cell gives the larger size in bytes, with `--json` or without it. The error cell gives stderr; the others give stdout.

| UAT | Extension | `skill view nicely` | `skill view <slug>` | `describe` | `describe <cmd>` | `doctor <slug>` | Dry run | Run | Error |
|---|---|---|---|---|---|---|---|---|---|
| Reference, `052b3184` | video-convert v0.1.2 | 2,071 | 1,278 | 2,036 | 3,743 | 181 | 241 | 169 | 200 |

The reference row set the budgets on 2026-10-29, on the Linux test host, with a 2-second clip that ffmpeg made. Its `describe video convert --json` counts 1,002 tokens with `o200k_base`. #375 measured 3,830 bytes for it on `71125d9`.
