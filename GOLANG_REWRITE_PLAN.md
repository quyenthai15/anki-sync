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

**Remaining Sessions:**
- 🔄 Enhanced CLI features (progress bars, colors)
- 🔄 Batch processing with goroutines
- 🔄 Bounded concurrency with errgroup
- 🔄 Comprehensive test suite
- 🔄 Final integration and polish

---

## Enhancement Roadmap

### 1. Enhanced CLI Experience
- **Progress bars**: Real-time progress indication for long-running operations
- **Colored output**: ANSI color codes for different message types (success, warning, error)
- **Verbose mode**: Detailed logging with `-v` flag
- **Interactive confirmations**: Prompt before destructive operations in non-dry-run mode

### 2. Batch Processing & Concurrency
- **Parallel Anki data fetching**: Replace sequential `notesInfo` calls with concurrent goroutines
- **Concurrent audio downloads**: Separate media storage from field updates using worker pools
- **Bounded concurrency**: Use semaphores to limit concurrent AnkiConnect requests (max 5)
- **Error handling**: Implement `errgroup` for clean error propagation across goroutines

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

---

## Implementation Priority

**Phase 1 (Immediate)**
1. Implement enhanced CLI features (progress bars, colors)
2. Implement batch processing & concurrency with goroutines
3. Add comprehensive test suite

**Phase 2 (Short-term)**
1. Implement record count limiting (`-n` flag)
2. Implement diff preview enhancements
3. Add retry logic and error handling improvements


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
    force     = flag.Bool("force", false, "Force update existing notes")
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
