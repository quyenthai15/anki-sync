# CLI Reference

## Commands

| Command | Purpose |
| --- | --- |
| `anki-sync check` | Verify AnkiConnect is reachable at `http://localhost:8765`. |
| `anki-sync profiles` | List configured profiles. Use this instead of reading profile files. |
| `anki-sync import --dry-run <draft>` (or `-n`) | Validate the draft and preview notes. Does not write to Anki. |
| `anki-sync import <draft>` | Add notes to Anki and generate TTS audio. |

## Draft Format

`.anki-draft.yaml`:

```yaml
profile: french
cards:
  - Expression: la lune
    Reading: /lyn/
    Meaning: the moon (mặt trăng)
    Type: Nom (f)
    Sentence_FR: La lune est très belle ce soir.
    Sentence_EN: The moon is very beautiful tonight.
```

- `profile`: a key returned by `anki-sync profiles`.
- `cards`: a non-empty list. Each card has only the keys in that profile's fields.

## Profile Location

Profiles are read from `.profiles.json` in the current directory (or a parent). If a profile is missing or the file fails to parse, report the error verbatim.

## Error Handling

- Treat any non-zero exit or error output as a stop condition. Report the exact message.
- Do not call AnkiConnect directly, and do not edit Anki's data files.
- Do not retry unless the user asks or confirms the cause is fixed.
