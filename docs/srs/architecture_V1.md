# Software Requirements Specification (SRS)

## Pharmaceutical Industrial Data Acquisition & Analytics Platform

**Version:** 0.3 (Schema Refactor)

**Status:** Implementation Phase

---

# 1. Technology Stack

* Go (module root at `project/`)
* QuestDB (time-series telemetry)
* PostgreSQL (persistent configuration and business data)
* Vanilla HTML/JS dashboard (embedded) + React SPA (separate)
* Docker + Docker Compose

# 2. Repository Layout

```
pharma-platform/
├── project/                # Go module root
│   ├── cmd/                # Entry points (6 binaries)
│   │   ├── migrate/        # Schema migration (QuestDB + Postgres)
│   │   └── ...
│   ├── internal/           # All Go packages
│   ├── config/             # bootstrap.yaml
│   ├── deploy/             # SQL migrations
│   │   ├── postgres/
│   │   │   ├── init/       # Postgres schema DDL
│   │   │   └── seed/       # Postgres seed data
│   │   └── questdb/init/   # QuestDB table DDL + materialized views
│   ├── runtime/            # Docker files + compose
│   ├── go.mod
│   └── go.sum
├── web/                    # React SPA (separate frontend)
├── persistent/             # Docker bind-mount volumes
├── docs/                   # ADRs, SRS, roadmap
├── Makefile                # Developer commands
└── .gitignore
```

# 3. High-Level Architecture

```
                   Users
                      |
           ┌──────────┴──────────┐
           │                     │
           ▼                     ▼
      Go API Server         React SPA (optional)
           │                     │
           └──────────┬──────────┘
                      │
              ┌───────┴───────┐
              │               │
              ▼               ▼
          QuestDB         PostgreSQL
              ▲
              │
        PLC Collector
              ▲
              │
         PLC Network
```

# 4. Database Responsibilities

### QuestDB (project/deploy/questdb/init/)
- `plc_samples` — raw telemetry (`machine_id`, `machine_name`, `tag_name`, `value`, `quality`)
- `plc_samples_1m` — 1-minute aggregated materialized view
- `plc_samples_1h` — 1-hour aggregated materialized view
- `plc_samples_1d` — 1-day aggregated materialized view
- `plc_samples_1w` — 1-week aggregated materialized view
- `alarms` — alarm events
- `events` — batch and machine events
- `logs` — system logs

### PostgreSQL (project/deploy/postgres/init/)
- `machines` — machine/PLC inventory
- `tags` — tag definitions per machine (references `machines.id`)

# 5. Entry Points (project/cmd/)

| Binary | Postgres | QuestDB | Seed | Collector | API |
|---|---|---|---|---|---|
| `pharma-platform` | Schema | Tables | No | Idle | Yes |
| `dev-mode` | Schema+Seed | Tables | If empty | Mock | Yes |
| `api` | Schema+Seed | Tables | If empty | Stub | Yes |
| `collector-sim` | Read tags | Tables | No | Mock→QuestDB | No |
| `seed` | Schema+Seed | No | Always | No | No |
| `migrate` | Schema only | Tables | No | No | No |

All invoked via `make` from the repository root.

# 6. Configuration

Single file: `project/config/bootstrap.yaml`

```yaml
postgres:    # host, port, database, user, password
questdb:     # host, port, batch_size, flush_interval
api:         # host, port
collector:   # workers, queue_size
aggregator:  # interval
plant:       # name, location, timezone
```

# 7. Current Design Decisions (ADR Summary)

1. ADR-0001: QuestDB for time-series
2. ADR-0002: Go for backend
3. ADR-0003: PostgreSQL for business data
4. ADR-0004: `persistent/` directory, `project/` for go module
5. ADR-0005: Docker Compose at `project/runtime/docker-compose.yml`
6. ADR-0007: Protocol-agnostic PLC driver interface
7. ADR-0008: Collector with scheduler + worker pool
8. ADR-0011: 18-endpoint REST API
9. ADR-0012: Dashboard API v1
10. ADR-0013: Embedded SPA frontend
11. ADR-0014: Collector pause/resume
12. ADR-0015: Dev-mode with DB-backed mock data
13. ADR-0016: PostgreSQL store for machines and tags
14. ADR-0017: Bootstrap configuration
15. ADR-0018: Identity field refactoring (plc_id/tag_id → machine_id/machine_name/tag_name)

# 8. Identity Model

Telemetry samples in QuestDB use the following identity columns:

- `machine_id` (SYMBOL) — stable machine identifier e.g. `"1"`
- `machine_name` (SYMBOL) — human-readable machine name e.g. `"Fluid Bed Dryer"`
- `tag_name` (SYMBOL) — technical tag name e.g. `"Inlet_Air_Temp"`

The API surface still exposes URL parameters as `plc_id` and `tag_id` for backward compatibility, but maps them internally to `machineID` and `tagName`.

# 9. Data Flow

```
PLC Network
    │
    ▼
PLC Driver (opcua, mc, fins, etc.)
    │
    ▼
Collector (scheduler + worker pool)
    │  models.Sample{MachineID, MachineName, TagName, Value, Quality}
    ▼
QuestDB Writer (ILP over TCP, double-buffer)
    │  plc_samples,machine_id=...,machine_name=...,tag_name=... value=...,quality=...i
    ▼
QuestDB (plc_samples table + materialized views)
    │
    ├──► Reader (REST HTTP API) ──► Go API Server ──► Dashboard
    │
    └──► Materialized Views (1m, 1h, 1d, 7d) ──► Aggregate API
```
