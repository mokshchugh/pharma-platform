# ADR-0010: QuestDB Read Pipeline Using REST HTTP API

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform needs to retrieve telemetry data from QuestDB for API responses, dashboards, and the future aggregation service. The read patterns include:

* **Latest values**: For each PLC/tag pair, return the most recent sample.
* **Historical range**: For a specific PLC tag, return all samples within a time window.

Unlike the write path (high-throughput, continuous), the read path is request-driven with low-to-moderate query volume. Durability and throughput matter less here than simplicity, debuggability, and SQL expressiveness.

QuestDB exposes data through its PG wire protocol and a built-in REST HTTP API on port 9000.

## Decision

Implement reads against QuestDB using the **REST HTTP API** (`/exec?query=...`) rather than the PostgreSQL wire protocol.

A dedicated `Reader` struct wraps the HTTP client:

```go
type Reader struct {
    client *Client
}
```

Two query methods are provided:

### Latest

Executes QuestDB's `LATEST ON` SQL extension to retrieve the most recent sample per PLC/tag partition:

```sql
SELECT *
FROM plc_samples
LATEST ON timestamp PARTITION BY plc_id, tag_id
```

### History

Executes a filtered range query with RFC 3339 timestamps:

```sql
SELECT *
FROM plc_samples
WHERE plc_id = '<id>'
  AND tag_id = '<id>'
  AND timestamp BETWEEN '<start>' AND '<end>'
ORDER BY timestamp
```

The response is parsed from QuestDB's JSON dataset format (`{"dataset": [[...]]}`) and decoded into `[]models.Sample`.

## Alternatives Considered

### PostgreSQL Wire Protocol (PG Wire)

Pros

* Familiar SQL interface
* Standarized driver ecosystem
* Could reuse the same `database/sql` patterns as the PostgreSQL client

Cons

* QuestDB's PG wire implementation is read-only for the subset of SQL it supports
* Connection pooling adds complexity for occasional reads
* Less straightforward to debug than HTTP

---

### REST HTTP API (Selected)

Pros

* HTTP is trivially debuggable with `curl`.
* No connection pool to manage — uses `http.DefaultClient`.
* QuestDB's REST API exposes the full SQL engine.
* Stateless — each request is independent.
* JSON responses are easy to parse and inspect.
* No extra dependency beyond `net/http`.

Cons

* Higher per-request overhead than a persistent PG wire connection.
* No connection reuse optimizations (HTTP/1.1 keep-alive mitigates this).
* Large result sets require streaming support.

## Rationale

For the read path, operational simplicity outweighs raw throughput. The REST API allows any operator or developer to debug queries directly with `curl http://questdb:9000/exec?query=...` without needing a PostgreSQL client or connection configuration.

QuestDB's JSON response format maps cleanly to Go structs with standard `encoding/json`, keeping the reader implementation small and dependency-free.

The `LATEST ON` SQL extension provides a concise way to retrieve the per-tag latest value that would otherwise require a subquery or window function.

## Consequences

### Positive

* Reader implementation is under 150 lines of Go.
* No connection pooling, no driver management.
* Queries can be tested and debugged with standard HTTP tools.
* Easy to add new query methods (just write SQL and a decoder).

### Negative

* Higher latency per query compared to a persistent database connection.
* No streaming for large result sets — entire dataset is buffered in memory.
* HTTP port must be exposed or accessible within the Docker network.

## Future Considerations

If the query volume grows significantly, a read-through cache (for example, Redis for latest values, or in-memory caching in the API layer) can be introduced without changing the Reader.

For large history queries, the Reader should be extended with cursor-based pagination or QuestDB's `LIMIT`/`OFFSET` support.

If a dedicated analytics service is introduced, it may connect directly via PG wire for lower-latency OLAP-style queries.
