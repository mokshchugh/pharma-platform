# ADR-0016: PostgreSQL-Backed Machine and Tag Store

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform previously loaded machine and tag configuration from YAML files. This was replaced with PostgreSQL-backed stores to support runtime configuration and relational integrity.

## Decision

Store machine and tag configuration in PostgreSQL via `internal/store/` package.

### Schema (project/deploy/postgres/init/001_schema.sql)

```sql
CREATE TABLE IF NOT EXISTS machines (
    id              SERIAL PRIMARY KEY,
    machine_name    TEXT NOT NULL,
    brand           TEXT NOT NULL,
    model           TEXT NOT NULL,
    protocol        TEXT NOT NULL,
    connection_type TEXT DEFAULT 'ethernet',
    ip_address      TEXT,
    port            INTEGER,
    notes           TEXT,
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tags (
    id              SERIAL PRIMARY KEY,
    machine_id      INTEGER REFERENCES machines(id) ON DELETE CASCADE,
    tag_name        TEXT NOT NULL,
    description     TEXT,
    data_type       TEXT NOT NULL DEFAULT 'float64',
    scale_factor    REAL DEFAULT 1.0,
    unit            TEXT,
    address         TEXT NOT NULL,
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
```

### Seed Data (project/deploy/postgres/init/)

- `002_seed_machines.sql` — 11 Ethernet-capable machines
- `003_seed_tags.sql` — 128 tags across all machines

Seeds are idempotent (`WHERE NOT EXISTS`). Run automatically by `dev-mode`, `api`, and `seed` binaries when the `machines` table is empty. The production binary (`pharma-platform`) does not seed.

### Store Pattern

The `internal/store/` package implements the `PLCStore` and `TagStore` interfaces:

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

## Consequences

### Positive
* Relational, queryable, runtime-modifiable configuration
* No YAML files to maintain
* Automatic migration on startup

### Negative
* PostgreSQL is a hard dependency for most binaries
