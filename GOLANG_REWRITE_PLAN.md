# AnkiSync Go Rewrite — Implementation Plan

## Current Status

**Completed Sessions:**
- ✅ Go toolchain & basic program structure
- ✅ Structs, methods, and interfaces  
- ✅ JSON, file I/O, and command line args
- ✅ HTTP client and AnkiConnect integration
- ✅ Generics for typed API responses
- ✅ HTTP GET + CSV parsing
- ✅ Maps, slices, and diff logic
- ✅ String/URL handling and audio processing
- ✅ CLI flags and command structure
- ✅ Executor and basic reporter modules
- ✅ Configuration validation (on startup)
- ✅ Interactive Sync Preview (TUI)
- ✅ Default "Always Update" behavior

**Remaining Sessions:**
- 🔄 Batch processing with goroutines
- 🔄 Bounded concurrency with errgroup
- 🔄 Comprehensive test suite
- 🔄 Final integration and polish

---

## Enhancement Roadmap

#### 1. Enhanced CLI Experience
- ✅ **Interactive Sync Preview**: Before executing changes, show a beautifully rendered summary of what will be added or updated.
- ✅ **Default Update Behavior**: By default, the tool updates existing notes (equivalent to previous `--force`).
- 🔄 **Opt-out Updates**: Introduce an `--add-only` flag to skip updating existing notes and only add new ones.
- 🔄 **Progress bars**: Real-time progress indication for long-running operations (currently using spinners).
- 🔄 **Colored output**: ANSI color codes for different message types (mostly implemented in TUI).
- 🔄 **Verbose mode**: Detailed logging with `-v` flag.
- ✅ **Interactive confirmations**: Prompt before destructive operations.

#### Bubbletea approach

I've analyzed the current codebase and identified that the Syncer logic is linear and uses standard output for logging. To enhance the UX with a Terminal UI (TUI), I suggest using the Bubble Tea framework by Charmbracelet. It's the industry standard for modern Go TUIs, offering a clean Model-View-Update (MVU) architecture that fits perfectly with your project's goals.

**Proposed Approaches**

1. Interactive Sync Dashboard (Done)
Transform the linear log output into a live dashboard.
* Progress Bars: Use bubbles/progress to show real-time progress for fetching sheet data and executing Anki updates.
* Spinners: Use bubbles/spinner during validation steps (e.g., "Checking Anki connection...").
* Live Log: A scrolling area at the bottom showing the last few operations.

2. Pre-Sync Preview & Confirmation
Before execution, show a structured summary of what will happen.
* **Diff Preview Table**: Use `bubbles/table` or custom lipgloss styles to list exactly which notes will be added vs. updated, highlighting field changes.
* Interactive Confirmation: may drop --dry-run since we already have [Y/n] prompt


3. Selection & Configuration Menu
If the app grows, you could add an entry menu:
* Deck Selection: Choose which Anki deck to sync with if multiple are configured.
* Profile Switcher: Quickly switch between different sync configurations.

### 2. Batch Processing & Concurrency
- **Parallel Anki data fetching**: Replace sequential `notesInfo` calls with concurrent goroutines.
- **Concurrent audio downloads**: Separate media storage from field updates using worker pools.
- **Bounded concurrency**: Use semaphores (buffered channels) to limit concurrent AnkiConnect requests.
- **Error handling**: Implement error collection using channels or `sync.Mutex` for clean error propagation.

**Implementation Hints:**
- **Batching**: Use `s[i:min(i+batchSize, len(s))]` to process slices in chunks.
- **Goroutines**: Remember to pass loop variables as arguments to avoid closure captures.
- **Semaphores**: Use a buffered channel `make(chan struct{}, N)` to limit concurrent tasks.

**Helpful Resources:**
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [Go by Example: Worker Pools](https://gobyexample.com/worker-pools)
- [A Tour of Go: Concurrency](https://go.dev/tour/concurrency/1)

### 3. Record Management Features
- **Count limit flag**: `-n` flag to process only N records (useful for testing)
- **Resume capability**: Track processed records and resume from last checkpoint
- **Batch size control**: Configurable batch sizes for Anki operations
- **Rate limiting**: Configurable delays between API calls

### 4. Comprehensive Test Suite
- **Unit tests**: Table-driven tests for core logic (diff, validation, planning)
- **Integration tests**: HTTP mock tests for AnkiConnect interactions
- **Benchmark tests**: Performance testing for concurrent operations
- **Test coverage**: Target 80%+ code coverage

### 5. Additional Suggested Enhancements

#### Performance & Reliability
- **Connection pooling**: Verify/Optimize reuse of HTTP connections for AnkiConnect
- **Retry logic**: Exponential backoff for failed API calls
- **Caching**: Cache deck names and field templates
- **Memory optimization**: Streaming processing for large datasets

#### User Experience
- **Diff preview**: Enhance to show detailed changes before execution

#### Distribution
- Compile win / mac binaries for others to run

---

## Implementation Priority

**Phase 1 (Immediate)**
1. Implement interactive preview table (TUI)
2. Change default behavior to "Always Update" and add `--add-only` flag
3. Implement progress bars for sheet fetching and Anki updates

**Phase 2 (Short-term)**
1. Implement batch processing & concurrency with goroutines
2. Add comprehensive test suite
3. Implement record count limiting (`-n` flag)


---

## Technical Implementation Notes

### Concurrency Strategy
```go
// Use errgroup for bounded parallel processing
g, ctx := errgroup.NewContext(ctx)
g.SetLimit(5) // Max 5 concurrent requests

for _, chunk := range chunks {
    chunk := chunk
    g.Go(func() error {
        return processChunk(ctx, chunk)
    })
}
```

### CLI Enhancement Structure
```go
// Enhanced flag parsing
var (
    dryRun    = flag.Bool("dry-run", false, "Show changes without executing")
    addOnly   = flag.Bool("add-only", false, "Only add new notes, do not update existing ones")
    verbose   = flag.Bool("verbose", false, "Enable detailed logging")
    limit     = flag.Int("n", 0, "Process only N records (0 = all)")
    batchSize = flag.Int("batch-size", 50, "Batch size for Anki operations")
)
```

### Testing Strategy
- **Unit tests**: Focus on pure functions (validation, diff logic)
- **Integration tests**: Mock AnkiConnect server for API testing
- **Benchmark tests**: Compare sequential vs concurrent performance
- **End-to-end tests**: Full sync workflow with test data
