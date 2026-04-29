# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

A Go rewrite of an AnkiSync tool for learning idiomatic Go. Syncs Japanese vocabulary from a Google Sheet CSV into Anki via [AnkiConnect](https://git.sr.ht/~foosoft/anki-connect). The learning goal is clean, minimal, idiomatic Go — not feature creep.

## Commands

```bash
# Build
go build -o sync ./cmd/sync

# Run
./sync [--dry-run] [--force] [--skip-validate]

# Test
go test ./...
go test ./internal/plan/...        # single package
go test -run TestFunctionName ./internal/plan/...  # single test

# Cross-platform dist (mac-intel, mac-arm, win64)
make dist
```

## Architecture

```
cmd/sync/main.go          # Entry point: wires all components, starts TUI + goroutine
internal/
  config/   config.go     # Config struct, JSON load, validation (local, Anki, Sheet)
  anki/     anki.go       # AnkiConnect HTTP client, types, NormalizeAnkiNotes
  sheet/    sheet.go      # CSV fetcher over HTTP, streaming via ForEachRecord
  plan/
    plan.go               # PlanBuilder diffs sheet records against Anki notes
    audio.go              # TTSProvider interface, GoogleTTS, AudioProcessor
  executor/ executor.go   # Executes Plan: AddNotes/UpdateNotes via AnkiClient interface
  syncer/   syncer.go     # Orchestrates Validate → Run pipeline, reports progress
  status/   status.go     # Event interface + typed msgs (PhaseMsg, LogMsg, ProgressMsg, ConfirmMsg, ErrMsg); Reporter interface
  ui/       tui.go        # Bubbletea TUI: spinner, progress bar, confirmation prompt; MessageBus bridges goroutine ↔ Elm loop
```

### Data flow

1. `main` parses flags, loads config, wires `Syncer` with `AnkiClient`, `SheetFetcher`, `GoogleTTS`, `TUI`
2. Sync runs in a goroutine; TUI runs on the main thread via `tea.NewProgram`
3. `Syncer.Validate()` — local config → AnkiConnect (deck/model/fields) → Sheet headers
4. `Syncer.Run()` — fetches Anki notes into `NoteMap`, streams sheet records through `PlanBuilder`, asks user to confirm, then calls `Executor.Execute(plan)`
5. Progress reported via `status.Reporter` interface → `MessageBus` buffered channel → TUI event loop

### Key design decisions

- **Interface at the boundary**: `status.Reporter`, `plan.TTSProvider`, `executor.AnkiClient`, and `config.AnkiProvider` are interfaces — concrete types are injected in `main`. Packages don't import each other except through these contracts.
- **Plan as pure data**: `Plan` is a value type (`NewNotes []NoteToAdd`, `UpdatedNotes []NoteToUpdate`). Building it and executing it are separate steps.
- **Streaming sheet reads**: `ForEachRecord` streams CSV row-by-row rather than loading all records into memory.
- **TUI concurrency**: `MessageBus` owns two channels — `incoming chan status.Event` (buffered, 100) and `confirm chan bool`. `WaitForProgress()` returns a `tea.Cmd` that blocks on `incoming`, bridging the syncer goroutine and the Elm-style update loop. Non-blocking sends drop messages if the buffer is full (acceptable for UI progress).
- **Confirmation handshake**: Syncer calls `TUI.Confirm(summary)` which sends a `ConfirmMsg` into `incoming` then blocks on `confirm`. The TUI renders a `[y/N]` prompt; `y` sends `true` and resumes the event loop, `n`/`q` sends `false` and quits.
- **Step-driven rendering**: `View()` is split into `renderHeader`, `renderMainContent`, `renderLogs`, `renderFooter`. `renderMainContent` switches on `status.Step` (Idle → Validating → Fetching → Planning → Confirming → Executing → Done/Error).

### Known issues / deferred work
