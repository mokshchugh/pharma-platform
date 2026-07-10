# ADR-0009: QuestDB Write Pipeline Using InfluxDB Line Protocol

**Status:** Accepted

**Date:** 2026-07-07

## Context

The Collector emits individual `Sample` values on a channel as they are read from PLCs. These samples must be written to QuestDB for time-series storage and later retrieval.

QuestDB supports multiple ingestion methods:

* PostgreSQL wire protocol (PG wire)
* InfluxDB Line Protocol (ILP) over TCP
* REST HTTP API
* CSV upload

The telemetry workload involves sustained high-frequency writes (potentially thousands of samples per second). The write path must batch samples for efficiency, handle transient network interruptions, and avoid blocking the collector workers.

## Decision

Write samples to QuestDB using the **InfluxDB Line Protocol over TCP** on port 9009, with a **double-buffer asynchronous writer**.

### Wire Protocol: ILP over TCP

ILP is QuestDB's recommended path for high-throughput ingestion. Each sample is serialized as a single line:

```
plc_samples,machine_id=<escaped>,machine_name=<escaped>,tag_name=<escaped> value=<encoded>,quality=<int>i <nanoseconds>
```

The encoder handles symbol escaping (spaces, commas, equals signs) and value encoding across all supported data types (bool, int variants, float variants, string).

### Writer Architecture: Double-Buffer Pipeline

The writer uses a pool of reusable buffers and two goroutines:

**Accumulator goroutine:**
* Pulls samples from the input channel.
* Appends them to a buffer until `BatchSize` is reached or `FlushInterval` elapses.
* Hands the full buffer to the flush channel and acquires an empty buffer from the free pool.

**Flush goroutine:**
* Dequeues full buffers from the flush channel.
* Encodes the batch to ILP text.
* Writes the byte slice to the TCP connection using `writeAll` (retries partial writes).
* Returns the buffer to the free pool for reuse.

```
samples ──► Accumulator ──► flushBuf ──► Flush ──► TCP (port 9009)
                 ▲                        │
                 │                        ▼
              freeBuf ──────────────── Buffer Pool
```

### Reconnection

On write failure, the writer closes and re-establishes the TCP connection, then retries the write once. This handles temporary QuestDB restarts without collapsing the pipeline.

### Metrics

A separate goroutine logs the sustained write throughput (samples/second) every second using an atomic counter.

## Alternatives Considered

### Synchronous Per-Sample Write

Pros

* Simplest implementation
* No buffering delay

Cons

* Extremely poor throughput (one TCP round-trip per sample)
* Would bottleneck the collector workers
* Unusable at production scale

---

### REST HTTP Batch Write

Pros

* HTTP is easier to debug
* Standard tooling

Cons

* Higher per-request overhead
* TLS termination adds latency
* Slower than raw TCP for high-frequency ingestion

---

### PG Wire Protocol

Pros

* SQL interface
* Familiar to PostgreSQL users

Cons

* Higher per-message overhead
* Not optimized for time-series ingestion
* Less well-documented for QuestDB

---

### Double-Buffer ILP over TCP (Selected)

Pros

* Maximum ingestion throughput (QuestDB's recommended path)
* Configurable batch size and flush interval
* Buffer reuse minimizes GC pressure
* Asynchronous pipeline decouples write latency from collector workers
* Reconnection logic handles transient failures

Cons

* More complex than synchronous approaches
* ILP format requires careful encoding (symbol escaping, type encoding)
* Buffering adds a small delay before data is visible in queries

## Rationale

ILP over TCP is QuestDB's native high-performance ingestion path. The double-buffer architecture ensures that the accumulator never blocks on network I/O and the flusher never blocks on sample arrival. Reusable buffers reduce allocation pressure under sustained throughput.

The trade-off of slight ingestion-to-query latency (bounded by `FlushInterval` and `BatchSize`) is acceptable for an industrial telemetry platform where durability and throughput matter more than sub-second query visibility.

## Consequences

### Positive

* Sustained high write throughput.
* Collector workers are never blocked on database writes.
* Configurable batching allows tuning for latency vs. throughput.
* Buffer pooling reduces memory allocation.

### Negative

* Data is not immediately visible in queries until the buffer is flushed.
* ILP encoding must be kept in sync with the QuestDB table schema.
* Two goroutines and channel coordination add internal complexity.

## Future Considerations

If the platform adds a second QuestDB node or migrates to QuestDB Enterprise with clustering, the writer can be extended to shard writes or use a load-balanced TCP connection pool without changing the accumulator or encoder.

The `samples/sec` metric output can be consumed by a centralized monitoring system (for example, Prometheus via a future metrics endpoint).
