# Mephisto

Mephisto is a distributed, replicated key-value storage engine written in Go from scratch. It is designed to demonstrate core principles of systems programming, storage engine durability, and consensus-based replication.

Currently, Mephisto implements a durable Write-Ahead Log (WAL) with binary record framing and corruption detection, laying the foundation for a Log-Structured Merge-tree (LSM) storage engine.

---

## Architecture Overview

Mephisto uses an LSM-tree (Log-Structured Merge-tree) design. Every write transaction is first serialized and appended to a Write-Ahead Log (WAL) to guarantee durability before updating the in-memory state.

```text
                  WRITE WORKFLOW
                  
            Client calls: Put(key, val)
                       │
                       ▼
             ┌──────────────────┐
             │  Record Encoding │ (Adds CRC32 & length-prefixes)
             └─────────┬────────┘
                       │
                       ▼
             ┌──────────────────┐
             │ Write-Ahead Log  │ (Appends to disk & calls fsync)
             └─────────┬────────┘
                       │
                       ▼
             ┌──────────────────┐
             │     Memtable     │ (In-memory sorted index)
             └──────────────────┘
```

### Key Technical Decisions

* **Binary Length-Prefixed Framing:** Rather than slow, text-based delimiters (like `\n` or JSON), Mephisto uses binary framing with a fixed 13-byte header containing checksums and payload lengths. This makes writes binary-safe, avoids escaping overhead, and enables fast, single-syscall reads via `io.ReadFull`.
* **CRC32 Checksum Validation:** Every record is signed with an IEEE CRC32 checksum over the operation, key, and value payload. If a bit flips on disk or the system loses power mid-write, Mephisto detects the corruption on recovery and rejects the corrupted write.
* **Strict Durability (Sync before Ack):** To guarantee no data loss, writes are synchronously flushed from the operating system's page cache to physical storage using the `fsync` system call (`file.Sync()`) before client acknowledgement.

---

## Directory Structure

```text
mephisto/
├── docs/
│   └── implementation_plan.md   # Project roadmaps & checkpoints
├── storage/
│   ├── record.go                # Binary record framing logic
│   ├── record_test.go           # Unit tests for framing & corruption
│   ├── wal.go                   # Write-Ahead Log engine
│   └── wal_test.go              # Tests for WAL durability & recovery
├── go.mod
└── README.md
```

---

## Running the Tests

To verify that record framing, checksum verification, and log recovery work correctly on your machine, run:

```bash
go test -v ./storage/...
```

---

## Project Roadmap

- [x] **Phase 0:** Project initialization & layout.
- [x] **Phase 1 (Active):** Durable storage engine.
  - [x] Binary Record framing & serialization.
  - [x] WAL implementation with synchronous `fsync`.
  - [ ] Memtable (In-memory sorted index).
  - [ ] SSTable writer (Flush memory to disk).
- [ ] **Phase 2:** Raft Consensus Replication (Log replication & leader election).
- [ ] **Phase 3:** Sharding & horizontal scaling.
