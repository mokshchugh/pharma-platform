# ADR-0001: Use QuestDB for Raw Telemetry Storage

**Status:** Accepted

**Date:** 2026-06-28

## Context

The platform continuously receives high-frequency telemetry from multiple PLCs. The database must efficiently ingest timestamped industrial data while supporting fast analytical queries over recent and historical measurements.

Apache IoTDB was initially considered because of its mature open-source ecosystem and focus on IoT and Operational Technology workloads.

## Decision

Use QuestDB as the primary time-series database for raw telemetry.

## Alternatives Considered

### Apache IoTDB

Pros

* Purpose-built for IoT workloads
* Rich open-source ecosystem
* Hierarchical time-series model
* Strong industrial use cases

Cons

* Additional operational complexity
* Smaller community than traditional SQL databases
* Less familiar PostgreSQL ecosystem

### PostgreSQL + TimescaleDB

Pros

* Familiar SQL ecosystem
* Mature tooling
* Good relational support

Cons

* Additional extension dependency
* Less optimized for the intended ingestion workload compared to QuestDB

### QuestDB (Selected)

Pros

* Extremely high ingestion performance
* Native SQL support
* PostgreSQL wire protocol compatibility
* Excellent time-series query performance
* Simple Docker deployment
* Good fit for industrial telemetry

Cons

* Smaller ecosystem than PostgreSQL
* Fewer industrial examples than Apache IoTDB

## Rationale

Although Apache IoTDB provides a compelling industrial feature set, QuestDB offers excellent ingestion performance, SQL compatibility, operational simplicity, and straightforward integration with the rest of the platform.

The decision prioritizes efficient ingestion and simple deployment while maintaining flexibility through SQL-based tooling.

## Consequences

### Positive

* High write throughput
* Fast analytical queries
* Easy Docker deployment
* Familiar SQL interface
* PostgreSQL-compatible client connectivity

### Negative

* Smaller community than PostgreSQL
* Fewer industrial reference implementations than Apache IoTDB

## Future Considerations

If future requirements demand distributed clustering or specialized IoT features beyond QuestDB's capabilities, Apache IoTDB remains a viable alternative for evaluation.
