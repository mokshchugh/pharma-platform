# ADR-0002: Use PostgreSQL for Aggregated Business Data

**Status:** Accepted

**Date:** 2026-06-28

## Context

Initially, the architecture only considered a time-series database for storing telemetry.

During the design process, it became clear that the platform also requires persistent storage for business entities and aggregated information that are not naturally modeled as time-series data.

Examples include:

* Production summaries
* OEE calculations
* Shift reports
* Batch information
* User management
* Plant configuration
* Audit records

## Decision

Introduce PostgreSQL as the primary relational database alongside QuestDB.

QuestDB stores raw telemetry.

PostgreSQL stores processed, relational, and business-oriented data.

## Alternatives Considered

### QuestDB Only

Pros

* Simpler architecture
* Single database

Cons

* Not optimized for relational business entities
* Less suitable for transactional data
* Would mix telemetry with application data

### PostgreSQL Only

Pros

* Single database
* Mature ecosystem

Cons

* Raw telemetry ingestion would become the primary workload
* Less specialized for high-frequency time-series storage

### Dual Database Architecture (Selected)

QuestDB

* Raw telemetry
* Recent measurements
* Time-series analytics

PostgreSQL

* Aggregated metrics
* Business entities
* Configuration
* Authentication
* Reporting

## Rationale

Separating operational telemetry from business data allows each database to perform the workload it is optimized for while keeping the architecture modular and scalable.

## Consequences

### Positive

* Clear separation of responsibilities
* Better query performance for both workloads
* Easier long-term scalability
* Cleaner application architecture

### Negative

* Two databases to operate
* Aggregation service required to synchronize processed data

## Future Considerations

Additional databases (for example, Redis for caching or object storage for documents) may be introduced if future requirements justify them.
