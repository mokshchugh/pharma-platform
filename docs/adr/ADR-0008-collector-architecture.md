# ADR-0008: Telemetry Collector Architecture

**Status:** Accepted

**Date:** 2026-07-07

## Context

The Collector is the central data ingestion component. It must continuously read hundreds or thousands of tags from one or more PLCs, each with its own configured poll interval, and deliver samples downstream to the QuestDB write pipeline.

Key requirements:

* Support tags with different poll intervals (for example, critical process values every 100 ms, environmental readings every 5 s).
* Prevent duplicate reads of the same tag while one is in flight.
* Decouple tag scheduling from network I/O so that a slow PLC read does not block the scheduling loop.
* Gracefully handle PLC disconnections and partial read failures.
* Support throughput benchmarking without modification to application code.

## Decision

Implement the Collector as a two-stage pipeline:

**Scheduler (single goroutine)**
* Runs a fixed 100 ms ticker.
* On each tick, iterates all enabled tags and dispatches those whose `PollInterval` has elapsed (with a 2 ms tolerance for timer jitter).
* Maintains a `lastPoll` map to track per-tag timing.
* Uses an `inFlight` set to prevent dispatching a tag whose prior read is still outstanding.
* Pushes eligible tags onto a buffered `workQueue`.

**Worker Pool (N goroutines)**
* Each worker dequeues tags from `workQueue`.
* Calls `driver.Read(ctx, tag)` — a blocking network call.
* Removes the tag from `inFlight` after the read completes (success or failure).
* Sends the resulting `Sample` to the output `samples` channel.
* Failed reads are logged and skipped; they will be retried on the next scheduled tick.

```
                     ┌────────────┐
                     │ 100ms Tick │
                     └─────┬──────┘
                           │
                     ┌─────▼──────┐
                     │ Scheduler  │
                     │ (1 goroutine)
                     └─────┬──────┘
                           │ workQueue
               ┌───────────┼───────────┐
               │           │           │
          ┌────▼───┐ ┌────▼───┐ ┌────▼───┐
          │Worker 1│ │Worker 2│ │Worker N│
          └────┬───┘ └────┬───┘ └────┬───┘
               │           │           │
               └───────────┼───────────┘
                           │ samples
                     ┌─────▼──────┐
                     │  Writer    │
                     └────────────┘
```

## Alternatives Considered

### Single Worker Per Tag

Pros

* Simple mental model
* No contention

Cons

* Too many goroutines at scale
* No sharing of limited PLC connections
* Hard to bound resource usage

---

### Poll-Interval Ticker Per Tag

Pros

* Precise timing per tag

Cons

* Unbounded goroutine count
* All tickers fire simultaneously on startup causing a read burst
* No built-in rate limiting

---

### Centralized Scheduler + Worker Pool (Selected)

Pros

* Bounded concurrency (configurable worker count)
* Tags with infrequent intervals do not consume a goroutine while idle
* In-flight deduplication handles slow PLC reads naturally
* Worker pool absorbs transient latency spikes
* Single source of timing truth simplifies debugging

Cons

* Scheduler tick resolution (100 ms) limits scheduling precision for intervals below 100 ms.
* All tags are inspected on every tick — negligible at expected tag counts but O(n) in the tag list size.

## Rationale

The two-stage design separates concerns: the scheduler owns timing and deduplication logic, while workers own network I/O. This prevents a blocked PLC read from delaying the scheduling of other tags and keeps the scheduler loop lightweight.

Configurable worker count allows tuning concurrency per deployment (matching PLC connection limits, CPU cores, or network bandwidth).

The fixed scheduler tick (100 ms) is a pragmatic choice: most industrial tags are polled at intervals of 100 ms or higher, and the 2 ms tolerance avoids unnecessary wake-ups on sub-millisecond timer drift.

## Consequences

### Positive

* Concurrent reads from multiple workers improve throughput on multi-core hosts.
* Slow PLCs do not starve fast PLCs.
* In-flight dedup prevents redundant reads of the same tag.
* Scheduler and worker pool are independently testable.
* Throughput is predictable and configurable via `Workers` and `QueueSize`.

### Negative

* Poll intervals below 100 ms are not supported precisely.
* Worker count must be tuned per deployment to avoid overwhelming the PLC or the database writer.
* Failed tag reads are silently retried — no alerting mechanism at the collector level.

## Future Considerations

If the platform expands to manage hundreds of PLCs with thousands of tags each, the scheduler may need sharding (one scheduler per PLC) to keep the per-tick iteration bounded.

A circuit breaker per PLC could prevent repeated reads against a disconnected device from wasting worker capacity.
