# KV Store

This repository is a small WAL-backed key-value store written in Go.

The main purpose of the project is learning Go basics by building a real, small system step by step instead of only reading language syntax in isolation.

Current goals:
- learn core Go syntax and structure
- understand packages, structs, methods, errors, slices, and maps
- practice building a small storage system with a write-ahead log
- get comfortable writing and running Go tests

## What It Does

The project currently supports:
- `set <key> <value>`
- `get <key>`
- `delete <key>`

Data is persisted using a write-ahead log (`wal.log`):
- every write is appended to the WAL
- the WAL is synced to disk before the operation is considered successful
- on startup, the store replays the WAL to rebuild in-memory state

## Why This Project

This is not meant to be a production database.

It is a playground for learning:
- how Go programs are structured
- how a CLI talks to internal packages
- how storage state can be rebuilt from an append-only log
- how to add tests while evolving behavior

## Project Layout

- `cmd/kvstore`
  - CLI entrypoint
- `internal/store`
  - store implementation and WAL encoding/decoding
- `go.mod`
  - module definition

## Running

From the project root:

```bash
GOCACHE=/tmp/go-build-cache go run ./cmd/kvstore ./data set name mahmoud
GOCACHE=/tmp/go-build-cache go run ./cmd/kvstore ./data get name
GOCACHE=/tmp/go-build-cache go run ./cmd/kvstore ./data delete name
```

You can also print usage with:

```bash
GOCACHE=/tmp/go-build-cache go run ./cmd/kvstore ./data help
```

## Testing

Run the full test suite with:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```

## Current Features

- WAL-backed persistence
- replay on startup
- empty-key validation
- rejection of unknown WAL operations
- unit tests for store behavior and CLI behavior

## Next Ideas

- detect truncated WAL records more explicitly
- improve CLI ergonomics further
- add compaction or snapshots later
