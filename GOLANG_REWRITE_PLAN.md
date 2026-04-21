# Go Rewrite Learning Plan — AnkiSync

This plan sequences Go learning so every concept is practiced immediately in the project.
Each session has: what to learn → what to build → resources.

---

## How to use this plan

1. Read the **Learn** material first (1–2 hrs)
2. Build the **Project Task** in a new repo: `anki-sync-go/`
3. Move to the next session only when the task compiles and runs correctly

---

## Phase 1 — Go Fundamentals

### ✅ Session 1 — The Go toolchain & your first program 

**Learn**
- How Go modules work (`go mod init`, `go.mod`, `go.sum`)
- `package main`, `func main()`, `go run`, `go build`

### ✅ Session 2 — Structs, methods, and interfaces

**Learn**
- Structs and field access
- Methods with value vs pointer receivers
- Interfaces (implicit satisfaction)

### ✅ Session 3 — JSON, file I/O, and OS args

**Learn**
- `encoding/json`: `json.Unmarshal`, struct tags
- `os.ReadFile`, `os.Args`

**Project Task — `config.go`**
- Port `lib/config.js` → `internal/config/config.go`.
- Read `config.json` with `os.ReadFile` + `json.Unmarshal`.

---

## Phase 2 — HTTP and AnkiConnect

### ✅ Session 4 — HTTP client and JSON round-trips

**Learn**
- `net/http`: `http.Post`, `http.NewRequest`, `http.Client`
- **Body Lifecycle**: Mandatory `resp.Body.Close()` and how it enables connection reuse.
- `json.Marshal` / `json.Unmarshal` on request/response.

**Project Task — `anki.go`**
- Port `lib/anki.js` → `internal/anki/anki.go`.
- Define `Action[T]` and `Response[T]` structs.
- Implement `invoke` (private helper) and public methods like `GetDeckNames`.

### ✅ Session 5 — Generics (light touch)

**Learn**
- `func Foo[T any](...)` for reusable API logic.

**Project Task — improve `anki.go`**
- Use generics in `invoke[T any]` to return typed results directly from AnkiConnect.

---

## Phase 3 — CSV, diff logic, and planning

### Session 6 & 7 — Maps, Closures, and the Streaming Diff

**Learn**
- **Maps in Depth**: Lookup with "comma ok" (`val, ok := m[key]`), `delete(m, key)`, and iteration.
- **Higher-Order Functions & Closures**: How to pass a `func` as an argument to implement the "Iterator" pattern.
- **Streaming IO**: Using `csv.Reader.Read()` in a loop to process infinite data with constant memory.
- **Resources**: 
    - [Go Blog: Go maps in action](https://go.dev/blog/maps)
    - [GobyExample: Closures](https://gobyexample.com/closures)
    - [Effective Go: Errors](https://go.dev/doc/effective_go#errors) (how to wrap errors in a stream)

**Project Task — The "Stream & Capture" Diff**
1.  **Refactor `internal/sheet/sheet.go`**: Implement `ForEachRecord(callback func(Record) error) error`.
2.  **Implement `internal/diff/diff.go`**: Logic to compare a single `Record` against an `AnkiNote`.
3.  **Wire `cmd/sync/main.go`**:
    - Load Anki notes into a map.
    - Stream the Sheet using `ForEachRecord`.
    - In the callback: Compare, track `toAdd`/`toUpdate`, and `delete` from the map to track "seen" notes.
    - Identify "Deletions" from the remaining map entries.

**Guiding Principles & Best Practices**
- **Memory Efficiency**: Keep the "Input side" (Sheet) O(1) memory and the "State side" (Anki) O(N).
- **Separation of Concerns**: `Fetcher` handles IO; `Planner` handles logic. The callback is the "glue".
- **Defensive Keys**: Trim spaces and normalize unicode (NFC) when using strings as map keys.
- **Batching**: Even though you process rows one-by-one, collect the *results* into a slice so you can send a single "Multi-Action" request to Anki at the end.

---

### Session 8 — Validate and planner

**Learn**
- Multiple return values.
- Wrapping errors: `fmt.Errorf("...: %w", err)`.

**Project Tasks**
- Port `lib/validate.js` → `internal/validate/validate.go`.
- Port `lib/planner.js` → `internal/planner/planner.go`.

---

## Phase 4 — Audio, execution, and the CLI

### Session 9 — String manipulation and URL encoding

**Learn**
- `net/url`: `url.QueryEscape`, `url.Values`
- `regexp`: `MustCompile`, `ReplaceAllString`
- `path/filepath` for safe filenames

**Project Task — `audio.go`**
- Port `lib/audio.js` → `internal/audio/audio.go`:
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
- Replace the `os.Args` check in `config.go` with proper `flag` parsing in `main.go`.
- The binary should support: `anki-sync [--force] [--dry-run]`

---

### Session 11 — Executor and reporter

**Learn**
- Variadic functions
- `fmt.Sprintf` formatting verbs
- ANSI escape codes in terminal output (same as Node — just strings)

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
- Table-driven tests for `BuildNoteFields` and `GetFieldChanges`.
- HTTP mock test for `ankiAction` using `httptest.NewServer`.

---

### Session 15 — Project structure and final wiring

**Learn**
- Standard Go project layout: `cmd/`, `internal/`
- `internal/` visibility rules
- `go build ./...`, `go test ./...`
- Cross-compilation: `GOOS=darwin GOARCH=arm64 go build`

**Resources**
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) (reference, not a rule)
- [Go Modules reference](https://go.dev/ref/mod)

**Project Task — final integration**
- Final directory structure wiring.
- Ensure `go build ./...` produces a single binary.

---

## Progress tracker

| Session | Topic                          | Status |
|---------|-------------------------------|--------|
| 1       | Toolchain & first program      | [x]    |
| 2       | Structs, methods, interfaces   | [x]    |
| 3       | JSON, file I/O, os args        | [x]    |
| 4       | HTTP client + ankiAction       | [x]    |
| 5       | Generics (typed ankiAction)    | [x]    |
| 6       | Maps, slices, diff.go          | [ ]    |
| 7       | HTTP GET + CSV parsing         | [x]    |
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
