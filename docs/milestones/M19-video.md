# M19 Video

Status: planned
Version: v0.19.0

## Demo

`ncly video convert clips/ --to mp4` converts a folder of videos through `ffmpeg`, after a dry run that lists every file it would write.

## Scope

A candidate everyday domain, to confirm in [M17](M17-inventory.md). `ncly video convert <input> --to mp4|webm|gif` wraps `ffmpeg`. It follows the [conventions for the everyday domains](M17-inventory.md#conventions-for-the-everyday-domains).

```
# A dry run on a folder lists every file it would write
exec ncly video convert clips/ --to mp4 --dry-run --json
stdout 'clips/a.mp4'
stdout 'clips/b.mp4'
! exists clips/a.mp4

# An existing file is not replaced without --force
exits 2 ncly video convert clips/a.mov --to mp4 --output existing.mp4 --json
stderr '"code":"CONFIRMATION_REQUIRED"'
```

## Open questions

- The spec of each command, in `extensions/video/spec.md`, and the cards

## Cards
