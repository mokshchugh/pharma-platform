# ADR-0014: Collector Pause/Resume and Graceful Shutdown

**Status:** Accepted

**Date:** 2026-07-07

## Context

The collector continuously reads tags from PLCs and pushes samples downstream to QuestDB. Two operational requirements emerged during development:

1. **Pause/Resume**: An operator or automated system needs to temporarily halt data collection without stopping the process — for example, during PLC maintenance, network changes, or to reduce database write load. The collector must resume from the paused state without restarting.

2. **Graceful Shutdown**: When the process receives a termination signal, in-flight samples must be flushed to QuestDB before the TCP connection is closed. Data already read from PLCs should not be lost.

## Decision

### Pause/Resume

Add an `atomic.Bool` field to the `Collector` struct with `Pause()`, `Resume()`, and `IsPaused()` methods:

```go
type Collector struct {
    ...
    paused atomic.Bool
}
```

The scheduler checks `paused` at the top of each 100 ms tick. When paused, the scheduler skips all tag dispatch. In-flight worker reads (already dispatched prior to the pause) complete normally — their samples flow through to the writer and are flushed to QuestDB. No reads are dropped mid-flight.

**Control paths:**

| Path | Mechanism |
|------|-----------|
| Signal | `SIGUSR1` pauses, `SIGUSR2` resumes |
| API | `POST /collector/pause`, `POST /collector/resume` |
| Programmatic | `collector.Pause()`, `collector.Resume()` |

### Graceful Shutdown Protocol

When a termination signal (SIGINT, SIGTERM) is received, the shutdown sequence is:

```
1. collector.Stop()    — stops scheduler + workers (stops producing)
2. close(samples)      — signals writer's accumulate goroutine to drain remaining buffered samples
3. writer.Stop()       — waits for accumulate & flushLoop to finish writing to TCP
4. cancel()            — cancels parent context (signals aggregator, metrics goroutines)
5. aggregator.Stop()   — waits for aggregation goroutine
6. server.Stop()       — shuts down HTTP server
```

The critical ordering is **stop producer → close channel → stop consumer**. Closing the `samples` channel causes the writer's accumulate goroutine to receive remaining buffered values from the channel (Go's channel semantics deliver buffered values before returning `ok=false`), flush its accumulated buffer to the flush queue, and exit. The flush loop drains the queue to TCP. `writer.Stop()` waits for all of this via `sync.WaitGroup`.

This eliminates data loss on shutdown.

## Alternatives Considered

### Context Cancel Only

Pros

* Simple — just cancel the context

Cons

* Accumulate goroutine exits via `<-ctx.Done()` without draining the samples channel
* Up to 100,000 buffered samples are silently discarded

---

### Pause via Context Cancellation

Pros

* Reuses existing context machinery

Cons

* Context cancellation is irreversible — cannot resume
* Would require recreating all goroutines on resume

---

### Atomic Bool Pause + Ordered Shutdown (Selected)

Pros

* Pause is reversible and instantaneous
* Shutdown ordering guarantees no data loss
* Signal and API control share the same code path
* No changes to worker or writer internals

Cons

* Two `select` cases to maintain in the scheduler (context cancel + pause check)
* In-flight reads complete after pause — brief lag before collection stops entirely

## Consequences

### Positive

* No data loss on clean shutdown.
* Pause/resume is instant and reversible.
* Dashboard controls and signal handlers share the same `Pause()`/`Resume()` methods.
* Shutdown protocol is explicit and auditable in code.

### Negative

* After pause, in-flight reads may take up to `driver.Read()` timeout to complete.
* Worker goroutines remain active during pause (idling on the work queue).
* The 100 ms scheduler tick continues running during pause — negligible CPU cost.

## Future Considerations

A future enhancement could add a `Drain()` method that pauses the collector, waits for all in-flight reads and writer buffers to flush, and reports completion. This would be useful for maintenance windows where the operator needs to know when the system is fully quiesced.

The shutdown protocol should be documented for any new component that joins the pipeline — each new goroutine must be stopped in the correct order relative to the channel it consumes from.
