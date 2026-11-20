# CLI Reference

## Locating the Binary

Binaries ship inside the skill at `<skill-dir>/bin/`:

| Platform | Detect | Path |
| --- | --- | --- |
| macOS, Apple Silicon | `uname -sm` prints `Darwin arm64` | `bin/darwin-arm64/anki-sync` |
| Windows 64-bit | `echo %PROCESSOR_ARCHITECTURE%` prints `AMD64` | `bin\windows-amd64\anki-sync.exe` |
| Anything else (e.g. Intel Mac, Linux) | | No binary. Report it and stop. |

Use the PATH binary first. On macOS, if the file is not executable, run `chmod +x`. If macOS blocks it as quarantined, tell the user to run `xattr -d com.apple.quarantine <path>`; do not run it with other flags.

## Commands

| Command | Purpose |
| --- | --- |
| `anki-sync check` | Verify AnkiConnect is reachable at `http://localhost:8765`. |
| `anki-sync profiles` | List configured profiles. Use this instead of reading profile files. |
| `anki-sync search --deck <deck> <words...>` | Check whether words already exist in a deck. |
| `anki-sync search --profile <profile> <words...>` | Same, using the profile's deck. |
| `anki-sync search --deck <deck> --field <field> <words...>` | Search one field for every word. |
| `anki-sync import --dry-run <draft>` (or `-n`) | Validate the draft and preview notes. Does not write to Anki. |
| `anki-sync import <draft>` | Add notes to Anki and generate TTS audio. |

## Searching Existing Notes

Use `search` before drafting when the user asks whether a word is already in a deck, or to avoid adding duplicates.

```bash
anki-sync search --profile french "le livre" flâner
anki-sync search --deck "Japanese::N5" 卵 たまご
anki-sync search --profile french --field Expression "le livre" flâner
```

- Pass exactly one of `--deck` or `--profile`; `--profile` resolves the deck from `.profiles.json`.
- Each word is checked independently. Output lists found note IDs or "not found" per word.
- Words are passed into an Anki search query sent directly to AnkiConnect. Plain terms are quoted as exact phrases and search all fields; terms containing `:` are passed through, so field queries like `Expression:"le livre"` or `Front:word` work if the field name is known.
- Add `--field <field>` to search one field for every word instead. The field name must match the note type (e.g. `Expression`, `Reading`, `Meaning`). Put flags before the words.
- Partial/token matches are possible depending on the query. Up to 10 note IDs are shown per word, then `+N more`.
- "Not found" is a normal result (exit `0`); command or AnkiConnect errors exit `1`.

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
