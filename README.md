# anki-sync

Create language flashcards for Anki with an AI agent. You describe the words, the agent drafts the cards (readings, example sentences, translations), you review the draft, and the CLI adds them to Anki through AnkiConnect.

This repo contains:

- `anki-sync`: the command-line tool (Windows and macOS).
- `.agents/skills/anki-sync/`: the agent skill that drives the workflow.
- `data/fr_sample.apkg`: a French A1 deck with the note type the profiles expect.
- `.profiles.example.json`: language profiles (deck, note model, fields).

## Requirements

- [Anki desktop](https://apps.ankiweb.net/) with the [AnkiConnect](https://ankiweb.net/shared/info/2055492159) add-on. Anki must be open while importing.
- An AI coding agent that supports skills (for example Claude Code). The skill is optional; the CLI works on its own.

## Setup

### 1. Import the deck (once)

In Anki, use **File > Import** and select `data/fr_sample.apkg`. This creates the note type the profile uses. Do not import it again, or every note will be duplicated.

### 2. Get the profile file

```
copy .profiles.example.json .profiles.json      (Windows)
cp .profiles.example.json .profiles.json        (macOS, or: make setup)
```

`.profiles.json` is your local copy and is git-ignored. Edit it to change decks or fields.

### 3. Get the binary

The binaries ship inside the skill folder, at `.agents/skills/anki-sync/bin/`:

- **Windows:** `bin\windows-amd64\anki-sync.exe`. If SmartScreen warns, click **More info > Run anyway**.
- **macOS (Apple Silicon):** `bin/darwin-arm64/anki-sync`. Make it executable and clear the quarantine flag:

```
chmod +x .agents/skills/anki-sync/bin/darwin-arm64/anki-sync
xattr -d com.apple.quarantine .agents/skills/anki-sync/bin/darwin-arm64/anki-sync
```

Intel Macs are not built yet.

### 4. Check the connection

With Anki open:

```
anki-sync check
```

Expected output: `AnkiConnect is running on localhost:8765`.

## Commands

| Command                              | What it does                                              |
| ------------------------------------ | --------------------------------------------------------- |
| `anki-sync check`                    | Confirms AnkiConnect is reachable                         |
| `anki-sync profiles`                 | Lists the configured languages and their decks            |
| `anki-sync import --dry-run <draft>` | Validates a draft and previews the notes (writes nothing) |
| `anki-sync import <draft>`           | Adds the notes to Anki                                    |

`<draft>` defaults to `.anki-draft.yaml`. The `-n` flag is a short alias for `--dry-run`.

### Draft format

```yaml
profile: french
cards:
  - Expression: la lune
    Reading: /lyn/
    Meaning: the moon
    Type: Nom (f)
    Sentence_FR: La lune est très belle ce soir.
    Sentence_EN: The moon is very beautiful tonight.
```

- `profile` must match a key from `anki-sync profiles`.
- `cards` must be a non-empty list. Use only the fields listed in that profile.

## Using the skill

The skill tells an agent how to draft cards, validate them, and import only after you confirm.

Copy the skill folder to your agent's skills directory:

```
Windows:  %USERPROFILE%\.claude\skills\anki-sync
macOS:    ~/.claude/skills/anki-sync
```

Then ask your agent something like _"Add たまご to my Japanese deck"_ or _"Make flashcards for these 5 French verbs."_ Review the preview, then say **import**.

## Troubleshooting

| Problem                              | Fix                                                                                                       |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------- |
| `AnkiConnect not reachable`          | Open Anki and check that AnkiConnect is enabled.                                                          |
| `failed to find profiles file`       | Create `.profiles.json` from `.profiles.example.json` in the folder you run the command from.             |
| `Import failed` with a profile error | Check that the draft's `profile` matches a key in `anki-sync profiles`.                                   |
| Duplicate notes after import         | Re-importing the `.apkg` or re-running the same draft creates duplicates. Delete the extra notes in Anki. |
| macOS: "cannot be opened"            | Run the `xattr` command in step 3.                                                                        |
| Windows: SmartScreen warning         | Click **More info > Run anyway**.                                                                         |

## Development

Requires Go 1.24+.

```
make build      # build ./anki-sync for this machine
make dist       # build macOS and Windows binaries into the skill's bin/, then zip the skill into dist/
make check      # run the commands above through make
make dry-run DRAFT=.anki-draft.yaml
make import  DRAFT=.anki-draft.yaml
make clean
```

Run `make help` for the full list.
