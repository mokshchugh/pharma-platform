# ADR-0015: Development Mode Architecture

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-07-07

## Context

The production platform requires PostgreSQL and QuestDB databases. During development, no real PLCs are available.

## Decision

Four development entry points under `project/cmd/`:

### 1. `cmd/dev-mode/main.go` — All-in-one dev
Runs migrations + seed + mock collector + writer + API in one process. Tags loaded from PostgreSQL via MachineStore/TagStore.

### 2. `cmd/api/main.go` — Standalone API + dashboard
Migrations + seed + API with no-op collector. No data generation.

### 3. `cmd/collector-sim/main.go` — Standalone simulator
Reads tags from PostgreSQL, writes mock data to QuestDB.

### 4. `cmd/seed/main.go` — Standalone seed
Runs PostgreSQL schema + seed SQL files.

All accessed via root `Makefile`:
```bash
make up     # start postgres + questdb
make seed   # seed database
make dev    # run all-in-one dev mode
make api    # run dashboard only
make sim    # run simulator
```

## Architecture

```
Makefile (root)
    │
    ├── make up → docker compose -f project/runtime/docker-compose.yml up
    ├── make dev → cd project && go run cmd/dev-mode/main.go
    └── make sim → cd project && go run cmd/collector-sim/main.go
```

## Mock Driver

Generates sine-wave data for all 128 tags loaded from PostgreSQL. Data types (bool, int16, int32, float32) are respected with appropriate base values.

## Consequences

### Positive
* Zero configuration — `make up && make dev` starts everything
* All 11 machines and 128 tags available
* Same store interfaces in production and development

### Negative
* Requires PostgreSQL running locally
* Wiring duplicated across four binaries
