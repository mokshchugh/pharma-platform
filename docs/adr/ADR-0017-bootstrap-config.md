# ADR-0017: Bootstrap Configuration

**Status:** Accepted

**Date:** 2026-07-07

## Context

Previously six separate YAML configuration files. Consolidated to a single file when PLC/tag config moved to PostgreSQL.

## Decision

Single `project/config/bootstrap.yaml` contains all settings:

```yaml
postgres:    # host, port, database, user, password, pool settings
questdb:     # host, port, batch_size, flush_interval
api:         # host, port
collector:   # workers, queue_size
aggregator:  # interval
plant:       # name, location, timezone
```

Config loader reads this single file into a `Config` struct:

```go
type Config struct {
    Plant      PlantConfig
    Collector  CollectorConfig
    API        APIConfig
    Aggregator AggregatorConfig
    Postgres   postgres.Config
    QuestDB    questdb.Config
}
```

All binaries load from the same file with `config.Load("config/bootstrap.yaml")`.

## Consequences

### Positive
* Single file management
* Consistent loading across all binaries
* No cross-file configuration drift

### Negative
* All settings in one file (mitigated by clear section naming)
