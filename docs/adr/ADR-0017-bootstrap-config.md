# ADR-0017: Bootstrap Configuration

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform previously used six separate YAML configuration files:

- `config/plcs.yaml`
- `config/tags.yaml`
- `config/plant.yaml`
- `config/collector.yaml`
- `config/api.yaml`
- `config/aggregation.yaml`

This required each file to be parsed independently and kept in sync. With the migration of PLC and tag configuration to PostgreSQL, only connection and runtime settings remain as configuration.

## Decision

Consolidate all configuration into a single `config/bootstrap.yaml` file:

```yaml
postgres:
  host: localhost
  port: 5432
  database: pharma
  user: postgres
  password: postgres
  max_open_conns: 20
  max_idle_conns: 10

questdb:
  host: localhost
  port: 9009
  batch_size: 1000
  flush_interval: 1s

api:
  host: 0.0.0.0
  port: 8081

collector:
  workers: 16
  queue_size: 10000

aggregator:
  interval: 1m

plant:
  name: Pharma Platform
  location: Manufacturing Facility
  timezone: Asia/Kolkata
```

The config loader reads this single file and returns a `Config` struct containing all settings:

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

## Rationale

- Single file is easier to manage than six
- No risk of cross-file configuration drift
- Load is a single function call
- All binaries use the same config format and loader
- Environment variable overrides can be added later without changing the file format

## Consequences

### Positive

- Simplified configuration management
- Consistent config loading across all binaries
- Clear, self-documented config file

### Negative

- All settings in one file (can be mitigated by clear section naming)
- Legacy `config/api.yaml` retained for reference but not loaded
