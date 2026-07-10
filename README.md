# Pharma Platform — Industrial Telemetry Platform

A production-grade platform for collecting, storing, and visualizing telemetry from pharmaceutical manufacturing equipment. Built with Go, QuestDB, PostgreSQL, and a React SPA dashboard.

## Architecture

```
PLC Network
    │
    ▼
PLC Drivers (OPC UA, MC, FINS, EtherNet/IP)
    │
    ▼
Collector (scheduler + worker pool + ILP writer)
    │
    ├──► QuestDB (time-series telemetry, materialized views)
    │
    └──► PostgreSQL (machine/tag configuration)
            │
            ▼
        Go API Server ────► React SPA / Embedded Dashboard
```

**Key design decisions:**
- QuestDB for high-ingestion-rate time-series storage with built-in aggregation views
- PostgreSQL for relational business data (machines, tags)
- Human-readable identity columns (`machine_name`, `tag_name`) embedded in every telemetry row — no cross-database JOINs for dashboard display
- Protocol-agnostic PLC driver interface — swap drivers without changing collection logic
- Double-buffered ILP writer — absorbs network latency without blocking the collector

## Quick Start

```bash
# 1. Start infrastructure (PostgreSQL + QuestDB)
make up

# 2. Seed the database (create schema + load plant data; once, or after reset)
make seed

# 3. Run dev mode (mock collector + API + embedded dashboard)
make dev
```

Open http://localhost:8081/

## Commands

| `make ...` | What it does |
|------------|-------------|
| `setup` | Creates persistent/ directories |
| `up` | setup + docker compose up (postgres + questdb) |
| `up-all` | Build and start everything via Docker Compose |
| `down` | docker compose down |
| `logs` | Tail docker compose logs |
| `dev` | run cmd/dev-mode (migrate + seed + mock collector + API) |
| `api` | run cmd/api (migrate + seed + API only) |
| `sim` | run cmd/collector-sim (mock data into QuestDB) |
| `seed` | run cmd/seed (schema + seed SQL for PostgreSQL) |
| `migrate` | run cmd/migrate (QuestDB tables/views + PostgreSQL schema) |
| `build` | go build ./... inside project/ |
| `prod` | run cmd/pharma-platform (production binary) |

## API Endpoints

### Telemetry

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/telemetry/latest` | Latest sample per machine/tag |
| `GET` | `/telemetry/latest/{plc_id}` | Per machine scoped latest |
| `GET` | `/telemetry/latest/{plc_id}/{tag_id}` | Single latest sample |
| `GET` | `/telemetry/history` | Historical samples (query: `plc_id`, `tag_id`, `start`, `end`) |
| `GET` | `/telemetry/aggregate/1m` | 1-minute aggregates |
| `GET` | `/telemetry/aggregate/1h` | 1-hour aggregates |
| `GET` | `/telemetry/aggregate/1d` | 1-day aggregates |
| `GET` | `/telemetry/aggregate/1w` | 1-week aggregates |

### Machines (PLCs)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/plcs` | List all machines |
| `GET` | `/plcs/{plc_id}` | Single machine details |
| `GET` | `/plcs/{plc_id}/status` | Machine connectivity and tag stats |
| `GET` | `/plcs/{plc_id}/tags` | Tags belonging to a machine |
| `PUT` | `/plcs/{plc_id}/pause` | Pause collection for a machine |
| `PUT` | `/plcs/{plc_id}/resume` | Resume collection for a machine |
| `GET` | `/plcs/status` | All machine statuses |
| `GET` | `/tags` | All tags across all machines |

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness check |

> Note: URL path parameters still use `plc_id`/`tag_id` for backward compatibility; internally they map to `machineID`/`tagName`.

## Project Structure

```
pharma-platform/
├── project/                    # Go module root
│   ├── cmd/                    # Entry points (6 binaries)
│   │   ├── pharma-platform/    # Production binary
│   │   ├── dev-mode/           # Development all-in-one
│   │   ├── api/                # Standalone API server
│   │   ├── collector-sim/      # Standalone simulator
│   │   ├── seed/               # Standalone DB seeder
│   │   └── migrate/            # Standalone migration runner
│   ├── internal/
│   │   ├── api/                # REST API handlers + router
│   │   ├── collector/          # Scheduler + worker pool
│   │   ├── plc/                # Driver interface + implementations
│   │   ├── questdb/            # Writer (ILP) + Reader (REST)
│   │   ├── postgres/           # Connection pool + migration
│   │   ├── store/              # PostgreSQL-backed MachineStore, TagStore
│   │   ├── config/             # Bootstrap config loader
│   │   ├── models/             # Sample, MachineConfig, Tag, etc.
│   │   └── aggregator/         # Materialized view aggregation
│   ├── deploy/
│   │   ├── postgres/init/      # PostgreSQL schema DDL
│   │   ├── postgres/seed/      # Seed data (11 machines, 128 tags)
│   │   └── questdb/init/       # QuestDB DDL + materialized views
│   ├── runtime/                # Docker compose + Dockerfile
│   ├── config/bootstrap.yaml   # Single config file
│   └── go.mod
├── web/                        # React SPA frontend
├── persistent/                 # Docker bind-mount volumes (git-tracked skeleton)
├── docs/                       # ADRs, SRS, roadmap
├── Makefile                    # Developer command shortcuts
└── README.md
```

## Storage Model

### QuestDB — Telemetry

The `plc_samples` table uses three identity columns:

| Column | Type | Example |
|--------|------|---------|
| `machine_id` | SYMBOL | `"1"` |
| `machine_name` | SYMBOL | `"Fluid Bed Dryer"` |
| `tag_name` | SYMBOL | `"Inlet_Air_Temp"` |
| `value` | DOUBLE | `25.4` |
| `quality` | INT | `192` |
| `timestamp` | TIMESTAMP | `2026-07-10T12:00:00Z` |

Materialized views (`plc_samples_1m`, `1h`, `1d`, `1w`) aggregate by `machine_id, machine_name, tag_name`.

### PostgreSQL — Configuration

| Table | Purpose |
|-------|---------|
| `machines` | Plant equipment inventory |
| `tags` | Tag definitions per machine |

## Configuration

Single file: `project/config/bootstrap.yaml`

```yaml
postgres:
  host: localhost
  port: 5432
  database: pharma
  user: pharma
  password: pharma

questdb:
  host: localhost
  port: 8812  # PostgreSQL wire protocol (writer)
  http_port: 9000
  batch_size: 500
  flush_interval: 100ms

api:
  host: 0.0.0.0
  port: 8081

collector:
  workers: 4
  queue_size: 1000

aggregator:
  interval_as_seconds: 60

plant:
  name: "Pharma Plant"
  location: "Building A"
  timezone: "Asia/Kolkata"
```

## Development

### Prerequisites

- Go 1.22+
- Docker + Docker Compose
- Make

### Workflow

```bash
# Start databases
make up

# Run seed (first time or after reset)
make seed

# Start dev (migrate + seed + mock collector + API + dashboard)
make dev

# In another terminal, start the React frontend
cd web && npm install && npm run dev

# Or use the built-in embedded dashboard at http://localhost:8081
```

### Running tests

```bash
cd project && go test ./...
```

## Design Records

Architecture Decision Records (ADRs) are in `docs/adr/`:

| ADR | Title |
|-----|-------|
| 001 | QuestDB for Time-Series Storage |
| 002 | Go for Backend Implementation |
| 003 | PostgreSQL for Business Data |
| 004 | `persistent/` and `project/` Directory Layout |
| 005 | Docker Compose for Local Development |
| 007 | Protocol-Agnostic PLC Driver Interface |
| 008 | Collector with Scheduler + Worker Pool |
| 009 | QuestDB Write Pipeline (ILP over TCP) |
| 010 | QuestDB Read Pipeline (REST API) |
| 011 | REST API Design (go-chi/chi) |
| 012 | Dashboard API v1 |
| 013 | Embedded SPA Frontend |
| 014 | Collector Pause/Resume |
| 015 | Dev-Mode with DB-Backed Mock Data |
| 016 | PostgreSQL Store for Machines and Tags |
| 017 | Bootstrap Configuration |
| 018 | Identity Field Refactoring (plc_id/tag_id → machine_id/machine_name/tag_name) |

## License

MIT
