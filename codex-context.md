# Codex Context

## Project Summary

This repository is a Go learning playground built as a small WAL-backed key-value store.

The user is learning:
- Go basics
- Vim motions while editing real code
- basic storage-engine ideas like WAL replay and durability

Important collaboration style for future sessions:
- prefer guiding by showing the code to write rather than doing all edits immediately
- include Vim motion tips during implementation
- the user does not know Go well yet, so explain Go concepts simply and concretely
- when the user asks to "review", inspect what they wrote, fix mistakes if requested, then run tests or commands
- when committing, split logic and tests into separate commits when practical

## Current Architecture

### CLI

File:
- `cmd/kvstore/main.go`

Behavior:
- command entrypoint
- supports:
  - `set <key> <value>`
  - `get <key>`
  - `has <key>`
  - `delete <key>`
  - `help`
- `set` and `delete` are silent on success
- `get` prints only the value
- `has` prints `true` or `false`
- exit code policy:
  - `0` success
  - `1` not found
  - `2` usage error
  - `3` runtime/store error
- prints errors to stderr through `main()`
- command parsing is handled by:
  - `run(args []string)`
  - `runSet(...)`
  - `runGet(...)`
  - `runDelete(...)`

### Store

Files:
- `internal/store/store.go`
- `internal/store/wal.go`

Behavior:
- `Open(path)` creates the data directory if needed
- opens `wal.log`
- replays WAL records into an in-memory `map[string][]byte`
- `Set` appends a WAL record, calls `Sync()`, then updates memory
- `Delete` appends a WAL record, calls `Sync()`, then updates memory
- `Get` reads from memory
- `Has` checks key existence directly from memory
- `Close` closes the WAL file

### WAL Format

In `internal/store/wal.go`, each record is encoded as:
- 4 bytes: CRC32 checksum
- 1 byte: operation
- 4 bytes: key length
- 4 bytes: value length
- key bytes
- value bytes

Supported operations:
- `opSet = 1`
- `opDel = 2`

Unknown WAL ops are rejected during replay.
Checksum mismatches are rejected during replay.

## Important Behavior Implemented

### Persistence

- writes are appended to `wal.log`
- store state is rebuilt by replaying the WAL on startup
- persistence across reopen is covered by tests

### Durability

- `Set` and `Delete` call `s.wal.Sync()` after `Write()`
- in-memory state is updated only after the sync succeeds

### Validation

- empty keys are rejected
- `ErrEmptyKey` is returned for `Set`, `Get`, and `Delete` when `len(key) == 0`
- empty values are still allowed

### Error Handling

- `ErrNotFound` for missing keys
- `ErrEmptyKey` for empty keys
- unknown WAL op causes `Open()` to fail during replay
- truncated WAL records cause `Open()` to fail with a clearer replay error message
- corrupted WAL records with checksum mismatch cause `Open()` to fail during replay

## Tests Present

### Store Tests

File:
- `internal/store/store_test.go`

Covers:
- set/get
- direct `Has` behavior
- persistence after reopen
- delete persistence after reopen
- empty key rejection
- rejecting unknown WAL ops during replay
- rejecting truncated WAL records during replay
- rejecting corrupted WAL checksum during replay
- immediate persistence expectations after `Sync()`

### CLI Tests

File:
- `cmd/kvstore/main_test.go`

Covers:
- `set` then `get`
- `has` for present and missing keys
- `has` empty-key usage handling
- `delete` then `get` not found
- usage error for too few args
- help output
- not-found exit code behavior for `get` and `has`
- empty key rejection through CLI

## Notable Go Concepts Already Explained To User

These have already been discussed and can be referenced again:
- packages
- imports
- structs
- methods with receivers
- `[]byte`
- `map[string][]byte`
- explicit `error` returns
- `:=` vs `=`
- `defer`
- `make([]byte, n)`
- WAL record layout
- `Sync()` meaning and why durability differs from normal correctness

## Commit History So Far

Recent commits:
- `84aa4ed` Document project purpose and usage
- `3c699cf` Add tests for synced WAL persistence
- `89ddc95` Sync WAL writes before applying changes
- `0f9d723` Add test for unknown WAL operations
- `6a0e725` Reject unknown WAL operations
- `2f0b3d0` Add empty-key validation tests
- `5311bdd` Reject empty store keys
- `6c6a31b` Add CLI commands and tests
- `25a3f28` Add store persistence tests
- `b06ff91` Implement WAL-backed store

## Current Repo State Notes

- there is an unrelated untracked file: `text.txt`
- it has intentionally been left out of commits

## Recommended Next Steps

Most sensible next engineering step:
- compaction / snapshotting, to avoid unbounded WAL growth and teach storage lifecycle
- WAL format versioning or migration support, since checksum-based records are now incompatible with older WAL files
- a clearer recovery policy for corrupted WALs

Other reasonable next steps:
- improve CLI ergonomics further
- consider adding explicit exit-code semantics for `has` or `get` not-found cases if the CLI is pushed further
- add compaction or snapshotting later

## Working Style Reminder

When continuing this project in a future session:
- first inspect current repo state before assuming anything
- continue teaching through incremental changes
- prefer small scoped steps
- when the user asks for diffs only, provide diffs only
- when the user asks for review, findings first, then fix if requested
- update this `codex-context.md` file whenever meaningful project progress is made so the handoff stays current
- when adding a feature, also update the relevant sections here:
  - implemented behavior
  - tests present
  - recent commits if useful for context
  - recommended next steps
- when one of the listed next steps is implemented, remove it from the list and replace it with the next most relevant follow-up item
