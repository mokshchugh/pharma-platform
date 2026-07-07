# ADR-0016: PostgreSQL-Backed Machine and Tag Store

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform previously loaded machine (PLC) and tag configuration from YAML files (`config/plcs.yaml` and `config/tags.yaml`). This approach had several limitations:

* YAML files needed manual synchronization with the actual PLC inventory
* No runtime configuration changes possible
* Each deployment required maintaining separate YAML files
* No relational integrity between machines and tags

The PLC inventory data from the factory survey (`PLC_Inventory_Tag_Mapping.xlsx`) contains 128 tags across 11 Ethernet-capable machines spanning 4 protocols (MC, FINS, OPC UA, EtherNet/IP).

## Decision

Store machine and tag configuration in PostgreSQL with the following schema:

### `machines` Table

```sql
CREATE TABLE machines (
    id              SERIAL PRIMARY KEY,
    machine_name    TEXT NOT NULL,
    brand           TEXT NOT NULL,
    model           TEXT NOT NULL,
    protocol        TEXT NOT NULL,       -- mc, fins, opcua, ethernetip
    connection_type TEXT DEFAULT 'ethernet',
    ip_address      TEXT,
    port            INTEGER,
    notes           TEXT,
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
```

### `tags` Table

```sql
CREATE TABLE tags (
    id              SERIAL PRIMARY KEY,
    machine_id      INTEGER REFERENCES machines(id) ON DELETE CASCADE,
    tag_name        TEXT NOT NULL,
    description     TEXT,
    data_type       TEXT NOT NULL DEFAULT 'float64',
    scale_factor    REAL DEFAULT 1.0,    -- x0.1, x0.01 scaling
    unit            TEXT,
    address         TEXT NOT NULL,       -- native PLC address string
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
```

### Store Pattern

The `internal/store/` package implements the `PLCStore` and `TagStore` interfaces previously defined in `internal/api/handlers/`:

```go
type PLCStore interface {
    GetPLCs() []models.PLC
    GetPLC(id string) *models.PLC
    GetTagsByPLC(plcID string) []models.Tag
}

type TagStore interface {
    GetTags() []models.Tag
    GetTag(id string) *models.Tag
}
```

This allows the API handlers to remain unchanged — only the backing implementation changes from in-memory YAML to PostgreSQL queries.

### Seed Data

Seed data from the factory survey is maintained as SQL files:

- `deploy/postgres/init/002_seed_machines.sql` — 11 machines
- `deploy/postgres/init/003_seed_tags.sql` — 128 tags

These are idempotent (`WHERE NOT EXISTS`) and run automatically by `cmd/dev-mode/main.go` and `cmd/api/main.go` when the `machines` table is empty. The production binary (`cmd/pharma-platform/main.go`) does not seed — machines are configured via the dashboard or API at deployment time.

## Consequences

### Positive

* Configuration is now relational, queryable, and runtime-modifiable
* No YAML files to maintain and synchronize
* Automatic migration on startup via `internal/store/migrate.go`
* Same interfaces across all binaries (prod, dev, api, simulator)

### Negative

* PostgreSQL is now a hard dependency for all binaries (except collector-sim which reads tags)
* Migration runner must handle both PostgreSQL and QuestDB DDL
