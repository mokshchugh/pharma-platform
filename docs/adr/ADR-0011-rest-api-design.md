# ADR-0011: REST API Design for Telemetry Access

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform requires an HTTP API to expose telemetry data to dashboards, external systems, and future frontend applications. The API must provide access to live and historical PLC readings without exposing the underlying database topology or query language.

At the time of design, the only data store is QuestDB. PostgreSQL is reserved for aggregated business metrics that will be produced by the future aggregation service.

Key requirements:

* Return the latest value for every PLC tag.
* Return historical values for a specific tag over a time range.
* Simple enough to be consumed without a client library.
* Built with minimal dependencies.

## Decision

Implement a REST API using the **go-chi/chi** router with three endpoints:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness check |
| `GET` | `/telemetry/latest` | Latest sample per PLC/tag |
| `GET` | `/telemetry/history` | Historical samples for a specific tag |

### Query Parameters (`/telemetry/history`)

| Parameter | Type | Description |
|-----------|------|-------------|
| `plc_id` | string | PLC identifier |
| `tag_id` | string | Tag identifier |
| `start` | RFC 3339 | Start of time range |
| `end` | RFC 3339 | End of time range |

### Response Format

All endpoints return JSON. Array endpoints return a JSON array of `Sample` objects directly (no envelope).

### Architecture

The API server is a standalone binary (`cmd/api/main.go`) separate from the Collector. It connects to QuestDB in read-only mode using the REST HTTP reader. This separation allows:

* Independent scaling of ingestion and serving capacity.
* API restarts without interrupting data collection.
* Read-only database access from the API process.

```
┌──────────┐     HTTP     ┌──────────┐     SQL      ┌──────────┐
│  Client  │ ──────────►  │  API     │ ──────────►  │  QuestDB │
│ (curl/GUI)│ ◄────────── │ (:8081)  │ ◄──────────  │ (:9000)  │
└──────────┘             └──────────┘              └──────────┘
```

Handler injection follows a simple dependency pattern: `TelemetryHandler` receives a `*questdb.Reader` at construction and is wired into the router.

## Alternatives Considered

### Embedded API in Collector Binary

Pros

* Single deployment artifact
* Shared database connections

Cons

* API restart requires stopping data collection
* API bugs risk ingestion stability
* Cannot scale read and write capacity independently

---

### GraphQL Endpoint

Pros

* Flexible querying
* Self-documenting schema

Cons

* Additional dependency and complexity
* Overkill for two query patterns
* Steeper learning curve for operational tooling

---

### Standard REST with chi Router (Selected)

Pros

* Minimal dependencies (chi is the only framework dependency).
* Familiar URL patterns for operators.
* Easy to test with `curl`.
* Chi's lightweight router adds negligible overhead.
* Clean separation of handlers, server, and routes.

Cons

* Handlers must manually parse query parameters and encode JSON.
* No built-in API documentation (future OpenAPI integration would be manual).

## Rationale

The API starts with two telemetry query patterns because they are the only read requirements validated by the current roadmap. A minimal surface reduces maintenance burden and allows the API to evolve based on real consumption patterns rather than speculative design.

go-chi is chosen over the standard library's `http.ServeMux` (which at the time of implementation lacks path parameter support) and over heavier frameworks like Gin for its idiomatic Go approach, composable middleware, and zero reflection magic.

Keeping the API as a standalone binary enforces a clean architectural boundary between data ingestion and data serving at minimal operational cost.

## Consequences

### Positive

* API and Collector can be deployed, updated, and scaled independently.
* API binary is small and starts quickly.
* Adding new endpoints is straightforward — add a handler method and register a route.
* Existing telemetry handlers serve as a clear template for future handlers (PLC CRUD, Tag CRUD).

### Negative

* Two separate binaries to build and deploy.
* No read-through caching — every API call queries QuestDB directly.
* Query parameter parsing and validation are manual in each handler.

## Future Considerations

As the platform grows, the API should gain:

* OpenAPI specification for client code generation.
* Pagination for history queries with large result sets.
* Authentication and authorization middleware.
* A caching layer (for example, Redis) for latest values to reduce QuestDB query load.
* Handler implementations for PLC and Tag CRUD operations (currently placeholder stubs).
