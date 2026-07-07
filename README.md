# Pharma Platform — Industrial Telemetry Platform

A production-grade platform for collecting, storing, and visualizing telemetry from pharmaceutical manufacturing equipment.

## Layout

```
pharma-platform/
├── persistent/          # Docker bind-mount volumes (git-tracked skeleton)
│   ├── postgres/
│   └── questdb/
├── project/             # Go module root + everything that runs
│   ├── cmd/
│   ├── internal/
│   ├── config/
│   ├── deploy/
│   ├── runtime/
│   │   └── docker-compose.yml
│   ├── go.mod
│   └── ...
├── docs/                # ADRs, SRS, roadmap
├── Makefile             # Wraps all commands
└── LICENSE
```

## Quick Start

```bash
# 1. Start infrastructure
make up

# 2. Seed the database (once, or after reset)
make seed

# 3. Run dev mode (mock collector + API + dashboard)
make dev
```

Open http://localhost:8081/

## Commands

| `make ...` | What it does |
|------------|-------------|
| `setup` | Creates persistent/ directories |
| `up` | setup + docker compose up (postgres + questdb) |
| `down` | docker compose down |
| `dev` | run cmd/dev-mode (migrate+seed+mock collector+API) |
| `api` | run cmd/api (migrate+seed+API only) |
| `sim` | run cmd/collector-sim (mock data to QuestDB) |
| `seed` | run cmd/seed (schema + seed SQL) |
| `build` | go build ./... inside project/ |
| `prod` | run cmd/pharma-platform (production) |
