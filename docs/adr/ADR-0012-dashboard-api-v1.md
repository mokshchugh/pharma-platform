# ADR-0012: Dashboard API v1 Endpoint Contract

**Status:** Accepted

**Date:** 2026-07-07

## Context

The earlier ADR-0011 defined a minimal 3-endpoint REST API (`/health`, `/telemetry/latest`, `/telemetry/history`) sufficient for basic telemetry access. Building a dashboard in React (or any frontend framework) requires a complete set of read-only endpoints covering all domain objects the UI renders.

Adding ad-hoc endpoints per dashboard component would lead to duplication, inconsistent response formats, and unnecessary round-trips. A single design pass for the v1 API surface was needed before frontend development began.

## Decision

Define an 18-endpoint REST API contract covering all dashboard read operations plus collector control:

### Telemetry

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/telemetry/latest` | Latest sample per machine/tag pair |
| `GET` | `/telemetry/latest/{plc_id}` | Latest per tag scoped to a machine |
| `GET` | `/telemetry/latest/{plc_id}/{tag_id}` | Single latest sample |
| `GET` | `/telemetry/history` | Raw samples within a time window |
| `GET` | `/telemetry/aggregate` | Time-bucketed min/max/avg |

### PLCs

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/plcs` | All configured PLCs |
| `GET` | `/plcs/{plc_id}` | Single PLC |
| `GET` | `/plcs/{plc_id}/status` | Machine connectivity and tag stats |
| `GET` | `/plcs/{plc_id}/tags` | Tags belonging to a machine |

### Tags

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/tags` | All configured tags |
| `GET` | `/tags/{tag_id}` | Single tag |

### Alarms

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/alarms` | Full alarm history |
| `GET` | `/alarms/active` | Currently active alarms |

### System & Collector

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/system/status` | Overall platform health summary |
| `GET` | `/collector/status` | Collector runtime state |
| `POST` | `/collector/pause` | Pause data collection |
| `POST` | `/collector/resume` | Resume data collection |
| `GET` | `/health` | Liveness check |

### Aggregate Query

`/telemetry/aggregate` uses QuestDB's `SAMPLE BY` SQL extension. Supported intervals map to tables managed by the aggregation service:

| Interval | Table |
|----------|-------|
| `1m` | `plc_samples_1m` (or raw `SAMPLE BY 1m`) |
| `1h` | `plc_samples_1h` (or raw `SAMPLE BY 1h`) |
| `1d` | `plc_samples_1d` (or raw `SAMPLE BY 1d`) |
| `1w` | `plc_samples_1w` (or raw `SAMPLE BY 1w`) |

### Response Format

All responses are JSON arrays or objects with no envelope. Error responses are plain text with the appropriate HTTP status code.

## Alternatives Considered

### Endpoint-Per-Component

Pros

* Tightly coupled to UI needs
* Potentially simpler queries

Cons

* Poor reusability
* Dashboard changes require API changes
* No stable contract for third-party consumers

### GraphQL

Pros

* Flexible querying
* Self-documenting

Cons

* Additional runtime dependency
* Over-engineered for the current query patterns
* Steeper learning curve for operators

### Flat 18-Endpoint REST (Selected)

Pros

* Every domain object has a stable URL
* Predictable response shapes
* Easy to test with curl
* Dashboard components fetch exactly what they need
* No envelope ceremony — raw JSON arrays

Cons

* More endpoints to maintain
* History and aggregate queries require parameter validation

## Rationale

The v1 contract is designed to be exhaustive for the dashboard's read needs while avoiding speculative endpoints. Each domain (telemetry, PLCs, tags, alarms, system) gets its own resource path, making the API navigable without documentation.

Aggregate telemetry is the most important endpoint — it enables the dashboard's chart views without requiring the frontend to compute aggregations from raw data or understand QuestDB SQL.

## Consequences

### Positive

* Frontend and backend can be developed in parallel against the contract.
* All dashboard pages have a purpose-built endpoint.
* No server-side rendering or templating is needed.
* Adding auth middleware later is straightforward (per-route or per-group).

### Negative

* 18 endpoints require 18 handler implementations.
* Some endpoints (PLC status, system status) compute derived values that add backend complexity.

## Future Considerations

When authentication is added, all endpoints except `/health` will require role-based access. Collector control endpoints (`/collector/pause`, `/collector/resume`) should be restricted to admin roles with audit logging.

The aggregate endpoint may be optimized to query pre-aggregated materialized views (e.g., `plc_samples_1h`) once the aggregation service populates them, falling back to `SAMPLE BY` on raw data until then.
