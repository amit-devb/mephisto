# mephisto — implementation plan

A distributed, replicated key-value store in Go. Own Raft implementation,
own log-structured storage engine, sharded across multiple Raft groups.

**Goal:** ship a CV-grade distributed systems project in ~4 months, to land
a role working on distributed systems / storage / cloud infra.

**Repo:** github.com/<you>/mephisto

---

## How to use this file

- Check off tasks as you finish them. Don't check off something you don't
  understand — the box is a proxy for "I could explain this in an
  interview," not "the code compiles."
- Each phase has a **shippable** — a working artifact you'd still have if
  everything after this phase got cut. If you're ever unsure what to work
  on, work on the current phase's shippable.
- "Explain it" checkpoints are mandatory stops. Don't move to the next
  section until you can answer the question out loud, from memory, without
  looking at your code.
- If you drift onto a tangent (a new project, a rabbit-hole optimization,
  a totally different tech), come back here and ask: does this serve the
  current phase's shippable? If not, park it in "Parking lot" below and
  return to the plan.

---

## Phase 0 — Setup

- [ ] Go installed (1.22+), `go version` confirmed
- [ ] Repo created and named (avoid collisions with existing projects —
      e.g. not "flink")
- [ ] `go mod init github.com/<you>/<name>`
- [ ] Minimal layout: `go.mod`, `.gitignore`, `README.md`, `storage/`
- [ ] First commit pushed

**Shippable:** an empty but real Go module on GitHub.

---

## Phase 1 — Storage engine (weeks 1–3)

Single machine, single process. Answers: *how do I not lose data if the
power dies mid-write?*

- [ ] Design `Record` struct (op, key, value, ...)
- [ ] Solve record framing (how a reader splits a byte stream back into
      exact records)
- [ ] WAL: append-only writer with `fsync` before ack
- [ ] WAL: reader/replay on startup, rebuilds in-memory state
- [ ] Memtable: in-memory sorted structure (map or skip list) serving
      reads/writes
- [ ] Flush: memtable → sorted on-disk file (SSTable) past a size threshold
- [ ] Get / Put / Delete over a simple local API (no network yet)
- [ ] Crash-recovery test: kill the process mid-write, restart, verify no
      data loss and no corruption
- [ ] **Explain it:** why fsync before ack, not after? Why is sequential
      append fast and in-place update slow? What's the failure mode if you
      skip fsync?

**Shippable:** a durable embedded KV store — a library, no server yet.

---

## Phase 2 — Raft replication (weeks 4–8)

Multiple machines, one group. Answers: *how do N copies agree on what
happened, even when one crashes or the network partitions?*

- [ ] `cmd/` created — first runnable binary (server wrapping the engine)
- [ ] Leader election (terms, votes, timeouts)
- [ ] Log replication (leader → followers, majority commit)
- [ ] Safety: election restriction, commit rules, log matching property
- [ ] Persistent Raft state survives a restart
- [ ] Snapshotting + log compaction (the part most people skip — don't
      skip it)
- [ ] Wire Raft's replicated log to the Phase 1 storage engine
- [ ] Cluster of 3–5 nodes running locally (processes or Docker)
- [ ] Fault test: kill the leader mid-write, cluster elects a new leader,
      no committed write is lost
- [ ] **Explain it:** why does a majority (not all nodes) need to
      acknowledge before commit? What's split-brain, and how does Raft's
      term mechanism prevent it? What breaks if you skip log compaction?

**Shippable:** a fault-tolerant replicated KV store. This is the core CV
artifact — if the 4 months ran out here, you'd still have something real.

---

## Phase 3 — Sharding (weeks 9–12)

Multiple Raft groups. Answers: *what happens when one group of 5 machines
isn't enough?*

- [ ] Keyspace partitioning scheme (range-based or hash-based — pick one
      and justify it)
- [ ] Run multiple independent Raft groups
- [ ] Shard controller: tracks which group owns which range
- [ ] Client-side or gateway routing to the correct group
- [ ] Shard rebalancing / migration (move a shard between groups without
      losing writes)
- [ ] **Explain it:** what has to happen for a shard migration to be safe
      mid-traffic? What are the tradeoffs of range vs. hash partitioning?

**Shippable:** horizontal scalability — the piece that separates a toy
from a real system.

---

## Phase 4 — Make it interview-legible (weeks 13–16)

- [ ] Compaction/merge for the LSM engine, if deferred from Phase 1
- [ ] MVCC or a linearizability test suite
- [ ] README: architecture diagram, design-decisions-and-tradeoffs section
- [ ] Benchmarks: throughput/latency, behavior under node failure
- [ ] A headline test: kill the leader mid-write, show zero data loss —
      this is the single most convincing artifact in the repo
- [ ] Clean up: dead code, TODOs, consistent naming
- [ ] (Optional, if time) gRPC or HTTP API polish, basic metrics/logging

**Shippable:** something a hiring manager can skim in 5 minutes and trust.

---

## Non-goals (for this project, deliberately out of scope)

Keep this list visible — it's here so scope creep has somewhere to go
that isn't "into the plan."

- Multi-region / geo-replication
- Pluggable consensus algorithms (Raft only)
- A SQL layer or query planner
- Production-grade auth/security
- A UI/dashboard beyond basic CLI or metrics endpoint

---

## Parking lot

Ideas, tangents, and "ooh interesting" detours that showed up mid-build.
Write them here instead of chasing them. Revisit only after a phase ships.

-

---

## Log

One line per session: date, what got done, what's next. Keeps momentum
visible and makes it easy to pick back up.

- YYYY-MM-DD — Bootstrapped repo, wrote this plan. Next: Phase 1, WAL
  record framing.