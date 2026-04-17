# AnkiSync Go

A high-performance rewrite of the AnkiSync tool, built to learn Go idiomatically while fixing the bottlenecks of the original Node.js implementation.

## The Mission
- **Learn Go by Doing:** Focus on idiomatic patterns (composition over inheritance, implicit interfaces, concurrency).
- **Performance:** Replace slow sequential JS loops with Go's powerful concurrency (goroutines/channels).
- **Reliability:** Leverage Go's type system to ensure sync plans are valid before touching the Anki database.

## Architectural Spirit
- **Top-Down Design:** We define the "Skeleton" (interfaces and data flows) before the "Muscles" (implementation).
- **Surgical Logic:** Minimal dependencies, standard library first.

## Current Progress: The Skeleton Phase
- [x] **Project Structure:** Established `cmd/sync` for the binary and `internal/` for private logic.
- [x] **Config System:** JSON loading with strict validation and CLI flag injection.
- [x] **Domain Mapping:** Defined core types for Anki (`Note`, `Action`) and Sheets (`Record`).
- [x] **Contract Definition:** Established `AnkiClient` and `SheetFetcher` interfaces in `main.go`.
- [ ] **Anki Implementation:** `invoke` helper drafted (next up: HTTP client).
- [ ] **Sheet Implementation:** Pending (Session 7).
