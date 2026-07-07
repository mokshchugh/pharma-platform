# ADR-0015: Development Mode Architecture

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-07-07

## Context

The production platform (`cmd/pharma-platform/main.go`) requires PostgreSQL and QuestDB databases, machine and tag configuration loaded from PostgreSQL, and real PLC drivers that connect to physical equipment.

During development, no real PLCs are available. Developers need a self-contained environment with simulated data.

## Decision

Create three development entry points:

### 1. `cmd/dev-mode/main.go` — All-in-one development

Runs migrations, seeds the database, starts a mock collector, writer, and API server in a single process:

```
dev-mode
    ├── PostgreSQL migration + seed (machines + tags from SQL files)
    ├── QuestDB migration (tables)
    ├── mockDriver (generates sine-wave values for all 128 tags)
    ├── Collector (reads tags from PostgreSQL DB via MachineStore/TagStore)
    ├── questdb.Writer (batch ILP to QuestDB)
    └── API server with embedded dashboard
```

### 2. `cmd/api/main.go` — Standalone API + dashboard

Runs migrations, seeds the database, starts the API server with a no-op collector. No data is generated — the dashboard shows the machine/tag list but no live telemetry.

### 3. `cmd/collector-sim/main.go` — Standalone simulator

Reads tags from PostgreSQL, runs QuestDB migrations, and writes mock data to QuestDB. No API server or dashboard. Useful for generating test data while iterating on the dashboard or API.

### 4. `cmd/seed/main.go` — Standalone seed

Runs PostgreSQL migrations and seeds the database with machines and tags. Useful after a fresh database reset.

## Architecture

```
cmd/dev-mode/main.go
    │
    ├── config.Load("config/bootstrap.yaml")
    ├── postgres.Connect → store.MigratePostgres(schema + seed)
    ├── questdb.Connect → store.MigrateQuestDB(tables)
    ├── store.NewMachineStore(postgresClient) → PLCStore
    ├── store.NewTagStore(postgresClient) → TagStore
    ├── questdb.NewWriter → QuestDB ILP
    ├── api.NewFull → HTTP server
    ├── Collector pause/resume via SIGUSR1/SIGUSR2
    └── Graceful shutdown
```

## Mock Driver

The mock driver reads tags from the database and generates realistic-looking sine-wave data:

```go
base = 42.0  // or 100.0 for integers, 1.0 for booleans
value = base + sin(time) * 10.0
```

## Alternatives Considered

### Hardcoded Mock PLCs (Previous approach)

Previously dev-mode hardcoded 3 PLCs with 25 tags each (75 total). This was replaced with DB-backed tags so:

- The same mock driver works with the full 128-tag inventory
- Tag metadata (units, scale factors, data types) comes from the real schema
- Changes to the inventory automatically flow to dev-mode

### Mock YAML Files

Requires maintaining separate mock configuration files. Rejected in favor of DB-backed approach.

## Consequences

### Positive

* Zero configuration — bootstrap.yaml is the only config file
* All 11 machines and 128 tags from the real inventory are available in dev
* Same store interfaces used in production and development
* Simulator can generate data independently of the API

### Negative

* Requires PostgreSQL running locally (or via docker compose)
* Wiring duplicated across four binaries
