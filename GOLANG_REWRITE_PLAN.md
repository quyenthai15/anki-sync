# Go Rewrite Learning Plan — AnkiSync

This plan sequences Go learning so every concept is practiced immediately in the project.
Each session has: what to learn → what to build → resources.

---

## How to use this plan

1. Read the **Learn** material first (1–2 hrs)
2. Build the **Project Task** in a new repo: `anki-sync-go/`
3. Move to the next session only when the task compiles and runs correctly

---

## Phase 1 — Go Fundamentals (no project code yet)

### Session 1 — The Go toolchain & your first program ✅

**Learn**
- How Go modules work (`go mod init`, `go.mod`, `go.sum`)
- `package main`, `func main()`, `go run`, `go build`
- Variables, types, `:=` vs `var`, zero values
- `fmt.Println`, `fmt.Printf`, `fmt.Errorf`

**Resources**
- [A Tour of Go — Basics](https://go.dev/tour/basics/1) (sections 1–14)
- [How to Write Go Code](https://go.dev/doc/code) (official guide, 10 min read)

**Practice Task**
```
mkdir anki-sync-go && cd anki-sync-go
go mod init github.com/yourname/anki-sync-go
```
Write `main.go` that prints `"AnkiSync v0.1"` and the current date using `time.Now()`.

---

### Session 2 — Structs, methods, and interfaces

**Learn**
- Structs and field access
- Methods with value vs pointer receivers
- Interfaces (implicit satisfaction — the key Go idea)
- `error` type and returning errors

**Resources**
- [A Tour of Go — Methods and Interfaces](https://go.dev/tour/methods/1) (sections 1–20)
- [Go by Example: Structs](https://gobyexample.com/structs), [Interfaces](https://gobyexample.com/interfaces), [Errors](https://gobyexample.com/errors)

**Practice Task**
Define a `Config` struct matching `config.json`:
```go
type Config struct {
    SheetCSVURL         string            `json:"sheet_csv_url"`
    AnkiDeck            string            `json:"anki_deck"`
    AnkiModel           string            `json:"anki_model"`
    Mapping             map[string]string `json:"mapping"`
    AudioField          string            `json:"audio_field"`
    SentenceAudioField  string            `json:"sentence_audio_field"`
    SentenceSourceCol   string            `json:"sentence_source_col"`
    ForceSync           bool
    DryRun              bool
}
```
Write a `String() string` method on it and print it.

---

### Session 3 — JSON, file I/O, and OS args

**Learn**
- `encoding/json`: `json.Unmarshal`, struct tags
- `os.ReadFile`, `os.Args`
- `log.Fatal` for early exits
- Slice iteration with `range`

**Resources**
- [Go by Example: JSON](https://gobyexample.com/json)
- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Go by Example: Command-Line Arguments](https://gobyexample.com/command-line-arguments)

**Project Task — `config.go`**
Port `lib/config.js` → `internal/config/config.go`.
- Read `config.json` with `os.ReadFile` + `json.Unmarshal`
- Check `os.Args` for `--force` and `--dry-run` flags
- Return `(Config, error)` — no `process.exit` in library code

---

## Phase 2 — HTTP and AnkiConnect

### Session 4 — HTTP client and JSON round-trips

**Learn**
- `net/http`: `http.Post`, `http.NewRequest`, `http.Client`
- Read response body: `io.ReadAll`
- `json.Marshal` / `json.Unmarshal` on request/response
- Custom `http.Transport` (equivalent to the Node `http.Agent`)

**Resources**
- [Go by Example: HTTP Client](https://gobyexample.com/http-client)
- [net/http docs](https://pkg.go.dev/net/http)
- [Go by Example: JSON](https://gobyexample.com/json) (revisit the decode side)

**Project Task — `anki.go`**
Port `lib/anki.js` → `internal/anki/anki.go`.
- Define `AnkiRequest` and `AnkiResponse[T]` structs (use generics or `any`)
- Write `ankiAction(action string, params any) (any, error)`
- Wire up a custom transport with `DisableKeepAlives: true` (matches current Node behaviour; you can relax this later)
- Test it: call `deckNames` and print the result

---

### Session 5 — Generics (light touch)

**Learn**
- Why generics: avoid `any` casts everywhere
- `func Foo[T any](...)`, type constraints
- When NOT to use generics (most of the time)

**Resources**
- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [Go by Example: Generics](https://gobyexample.com/generics)

**Project Task — improve `anki.go`**
Replace the `any` return with a typed helper:
```go
type AnkiResponse[T any] struct {
    Result T      `json:"result"`
    Error  string `json:"error"`
}

func ankiAction[T any](action string, params any) (T, error)
```
Use this to get a `[]string` back from `deckNames` without casting.

---

## Phase 3 — CSV, diff logic, and planning

### Session 6 — Maps, slices, and the standard library

**Learn**
- `map[K]V`: creation, lookup, delete, iteration order
- Slice patterns: `append`, spread (`s...`), copy
- `strings` package: `TrimSpace`, `Contains`, `ReplaceAll`
- `unicode/utf8` and `golang.org/x/text/unicode/norm` for NFC normalization

**Resources**
- [A Tour of Go — Maps](https://go.dev/tour/moretypes/19)
- [Go by Example: Maps](https://gobyexample.com/maps)
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text/unicode/norm) (add with `go get`)

**Project Task — `diff.go` + `anki.go` (getAnkiDataMap)**
Port `lib/diff.js` → `internal/diff/diff.go`:
- `BuildNoteFields(row map[string]string, mapping map[string]string) map[string]string`
- `GetFieldChanges(fields map[string]string, existing map[string]string) map[string][2]string` (returns `[old, new]` pairs or nil)

Port `getAnkiDataMap` in `anki.go`:
- Return `map[string]ExistingNote` keyed by the primary field value
- Keep sequential chunks for now (you'll parallelize in Phase 5)

---

### Session 7 — HTTP GET and CSV parsing

**Learn**
- `net/http` GET requests
- `encoding/csv`: `csv.NewReader`, `Reader.ReadAll`, `Reader.Read`
- Turning CSV rows + headers into `[]map[string]string`

**Resources**
- [Go by Example: HTTP Client](https://gobyexample.com/http-client)
- [encoding/csv docs](https://pkg.go.dev/encoding/csv)

**Project Task — `fetcher.go`**
Port the CSV fetch from `sync.js` lines 17–23 → `internal/fetcher/fetcher.go`:
- `FetchCSV(url string) ([]map[string]string, error)`
- Returns records as `[]map[string]string` (column name → value), trimmed and NFC-normalised

---

### Session 8 — Validate and planner

**Learn**
- Multiple return values
- Named return values (use sparingly)
- Wrapping errors: `fmt.Errorf("context: %w", err)` and `errors.Is`

**Resources**
- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values)
- [Go blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)

**Project Tasks**
- Port `lib/validate.js` → `internal/validate/validate.go`
- Port `lib/planner.js` → `internal/planner/planner.go`
  - Define `NoteToAdd` and `NoteToUpdate` structs
  - `BuildSyncPlan(records []map[string]string, cfg config.Config, ankiDataMap map[string]ExistingNote) ([]NoteToAdd, []NoteToUpdate, error)`

---

## Phase 4 — Audio, execution, and the CLI

### Session 9 — String manipulation and URL encoding

**Learn**
- `net/url`: `url.QueryEscape`, `url.Values`
- `regexp`: `MustCompile`, `ReplaceAllString`
- `path/filepath` for safe filenames

**Resources**
- [Go by Example: Regular Expressions](https://gobyexample.com/regular-expressions)
- [net/url docs](https://pkg.go.dev/net/url)

**Project Task — `audio.go`**
Port `lib/audio.js` → `internal/audio/audio.go`:
- `GetTTSUrl(text, lang string) string`
- `SanitizeFilename(name, prefix string) string`
- `GetAudioSyncPlan(cfg config.Config, fields map[string]string, existing *ExistingNote) AudioPlan`

---

### Session 10 — CLI flags

**Learn**
- `flag` package: `flag.Bool`, `flag.String`, `flag.Parse`
- Subcommand pattern if needed
- `os.Exit` and when to use it (only in `main`)

**Resources**
- [Go by Example: Command-Line Flags](https://gobyexample.com/command-line-flags)
- [flag package docs](https://pkg.go.dev/flag)

**Project Task — wire up `main.go`**
Replace the `os.Args` check in `config.go` with proper `flag` parsing in `main.go`.
The binary should support: `anki-sync [--force] [--dry-run]`

---

### Session 11 — Executor and reporter

**Learn**
- Variadic functions
- `fmt.Sprintf` formatting verbs
- ANSI escape codes in terminal output (same as Node — just strings)

**Resources**
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)

**Project Tasks**
- Port `lib/executor.js` → `internal/executor/executor.go`
  - `ExecuteUpdates(notes []NoteToUpdate) error`
  - `ExecuteAdditions(notes []NoteToAdd) error`
  - Drop the `canAddNotesWithErrorDetail` pre-check (see performance notes)
- Port `lib/reporter.js` → `internal/reporter/reporter.go`
  - `PrintDryRunSummary(toAdd []NoteToAdd, toUpdate []NoteToUpdate)`

---

## Phase 5 — Concurrency (the Go payoff)

### Session 12 — Goroutines and channels

**Learn**
- `go func()` — spawning goroutines
- Channels: unbuffered vs buffered, `<-ch`, `ch <-`
- `sync.WaitGroup`
- The rule: **don't communicate by sharing memory, share memory by communicating**

**Resources**
- [A Tour of Go — Concurrency](https://go.dev/tour/concurrency/1) (all sections)
- [Go by Example: Goroutines](https://gobyexample.com/goroutines), [Channels](https://gobyexample.com/channels), [WaitGroups](https://gobyexample.com/waitgroups)

**Project Task — parallelize `getAnkiDataMap`**
Replace the sequential `notesInfo` chunk loop with parallel fetches:
```go
// Fan out: launch one goroutine per chunk
// Fan in: collect results into the map with a mutex or channel
```
This is the biggest real-world speedup in the project.

---

### Session 13 — `errgroup` and bounded concurrency

**Learn**
- `golang.org/x/sync/errgroup`: cleaner goroutine error propagation than `WaitGroup`
- Semaphore pattern to limit concurrent requests (don't hammer AnkiConnect)
- `context.Context`: cancellation and passing through call chains

**Resources**
- [errgroup docs](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Go blog: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Go by Example: Context](https://gobyexample.com/context)

**Project Task — parallelize audio downloads**
In `executor.go`, separate `storeMediaFile` calls from `updateNoteFields`:
1. First: fan out all `storeMediaFile` actions in parallel (bounded to N=5 concurrent)
2. Then: send `updateNoteFields` batches

This unblocks field updates from slow Google TTS downloads.

---

## Phase 6 — Polish and final integration

### Session 14 — Testing in Go

**Learn**
- `testing` package: `func TestXxx(t *testing.T)`
- Table-driven tests (idiomatic Go pattern)
- `t.Run` for subtests
- Mocking HTTP with `net/http/httptest`

**Resources**
- [Go by Example: Testing](https://gobyexample.com/testing)
- [How to write table-driven tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [httptest docs](https://pkg.go.dev/net/http/httptest)

**Project Task — port the test suite**
- Table-driven tests for `BuildNoteFields` and `GetFieldChanges` (pure functions, easy to test)
- HTTP mock test for `ankiAction` using `httptest.NewServer`

---

### Session 15 — Project structure and final wiring

**Learn**
- Standard Go project layout: `cmd/`, `internal/`
- `internal/` visibility rules (packages only importable within the module)
- `go build ./...`, `go test ./...`
- Cross-compilation: `GOOS=darwin GOARCH=arm64 go build`

**Resources**
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) (reference, not a rule)
- [Go Modules reference](https://go.dev/ref/mod)

**Project Task — final integration**
Final directory structure:
```
anki-sync-go/
├── cmd/
│   └── sync/
│       └── main.go          ← entry point, flag parsing, spinner output
├── internal/
│   ├── anki/
│   │   └── anki.go          ← ankiAction, getAnkiDataMap
│   ├── audio/
│   │   └── audio.go
│   ├── config/
│   │   └── config.go
│   ├── diff/
│   │   └── diff.go
│   ├── executor/
│   │   └── executor.go
│   ├── fetcher/
│   │   └── fetcher.go
│   ├── planner/
│   │   └── planner.go
│   ├── reporter/
│   │   └── reporter.go
│   └── validate/
│       └── validate.go
├── config.json
├── go.mod
└── go.sum
```
Wire everything together. Run `go build ./...` — it should produce a single binary.

---

## Progress tracker

| Session | Topic                          | Status |
|---------|-------------------------------|--------|
| 1       | Toolchain & first program      | [ ]    |
| 2       | Structs, methods, interfaces   | [ ]    |
| 3       | JSON, file I/O, os args        | [ ]    |
| 4       | HTTP client + ankiAction       | [ ]    |
| 5       | Generics (typed ankiAction)    | [ ]    |
| 6       | Maps, slices, diff.go          | [ ]    |
| 7       | HTTP GET + CSV parsing         | [ ]    |
| 8       | Validate + planner             | [ ]    |
| 9       | String/URL + audio.go          | [ ]    |
| 10      | CLI flags                      | [ ]    |
| 11      | Executor + reporter            | [ ]    |
| 12      | Goroutines + parallel chunks   | [ ]    |
| 13      | errgroup + bounded concurrency | [ ]    |
| 14      | Testing                        | [ ]    |
| 15      | Final wiring                   | [ ]    |

---

## Key Go concepts map (JS → Go)

| JavaScript             | Go equivalent                          |
|------------------------|----------------------------------------|
| `require('./lib/x')`   | `import "module/internal/x"`           |
| `async/await`          | goroutines + channels / `errgroup`     |
| `Promise.all`          | `errgroup.Go` + wait                   |
| `try/catch`            | `if err != nil { return err }`         |
| `JSON.parse`           | `json.Unmarshal`                       |
| `process.argv`         | `os.Args` / `flag` package             |
| `process.exit(1)`      | `log.Fatal` or `os.Exit(1)` in main    |
| `console.log`          | `fmt.Println`                          |
| `Map`                  | `map[K]V`                              |
| `Array.filter/map`     | explicit `for range` loop              |
| `axios.post`           | `http.NewRequest` + `client.Do`        |
| `null`                 | pointer `*T` or zero value check       |
| `undefined`            | zero value (Go has no undefined)       |
| `interface {}`         | `any` (Go 1.18+)                       |
