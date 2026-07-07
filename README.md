# Pharma Platform — Industrial Telemetry Platform

A production-grade platform for collecting, storing, and visualizing telemetry from pharmaceutical manufacturing equipment.

## Architecture

```bash
pharma-platform/
├── cmd/
│   ├── pharma-platform/        # Production: migrate + API + dashboard
│   ├── dev-mode/               # Dev: migrate+seed + mock collector + API
│   ├── api/                    # Standalone API + dashboard
│   ├── collector-sim/          # Simulator: read tags from DB, write mock data
│   └── seed/                   # Standalone seed: populate DB from SQL files
├── runtime/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── entrypoint.sh
│   ├── docker-compose.yml
│   └── logs/
├── persistent/                  # Bind-mounted docker volumes (gitignored)
│   ├── postgres/
│   └── questdb/
├── deploy/
│   ├── postgres/init/
│   │   ├── 001_schema.sql       # CREATE TABLE IF NOT EXISTS machines, tags
│   │   ├── 002_seed_machines.sql # 11 machines from plant inventory
│   │   └── 003_seed_tags.sql     # 128 tags across all machines
│   └── questdb/init/
│       ├── 001_plc_samples.sql  # plc_samples table
│       └── 002_events.sql       # alarms, events, logs tables
├── internal/
│   ├── store/                   # PostgreSQL-backed stores
│   │   ├── migrate.go           # Run schema + seed migrations
│   │   ├── machine.go           # PLCStore impl
│   │   └── tag.go               # TagStore impl
│   ├── collector/               # Telemetry collector
│   ├── questdb/                 # QuestDB client + writer + reader
│   ├── postgres/                # PostgreSQL client + writer
│   ├── config/                  # Bootstrap config loader
│   ├── api/                     # REST API handlers + server
│   ├── models/                  # Domain models
│   └── plc/                     # PLC driver interface
├── config/
│   └── bootstrap.yaml           # Single config file: DB, API, collector
├── docs/
├── go.mod
└── README.md
```

## Quick Start (Development)

```bash
# Start infrastructure
docker compose -f runtime/docker-compose.yml up postgres questdb

# Terminal 1: Seed + mock collector + API (all-in-one)
go run cmd/dev-mode/main.go

# Or split:
# Terminal 1: API + dashboard (after seeding)
go run cmd/api/main.go

# Terminal 2: Simulator (writes mock data to QuestDB)
go run cmd/collector-sim/main.go
```

## Quick Start (Production)

```bash
docker compose -f runtime/docker-compose.yml up --build
```

## Storage

- **QuestDB**: Time-series telemetry (plc_samples, alarms, events, logs)
- **PostgreSQL**: Persistent data (machines, tags, users, config)
- **persistent/**: Bind-mounted volumes survive container rebuilds

## Entry Points

| Binary | Postgres | QuestDB | Seed | Collector | API |
|---|---|---|---|---|---|
| `pharma-platform` | Schema only | Schema + tables | No | Idle (no PLCs) | Yes |
| `dev-mode` | Schema + seed | Schema + tables | If empty | Mock driver | Yes |
| `api` | Schema + seed | Schema + tables | If empty | Stub | Yes |
| `collector-sim` | Read only | Schema + tables | No | Mock → QuestDB | No |
| `seed` | Schema + seed | No | Always | No | No |

## Config

Single `config/bootstrap.yaml` controls all settings:
- Postgres connection
- QuestDB connection
- API server
- Collector workers
- Aggregator interval
- Plant metadata
