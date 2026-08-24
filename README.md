# Pharma Platform — Industrial Telemetry Platform

A platform for collecting, storing, and visualizing telemetry from pharmaceutical
manufacturing equipment: OEE, production runs, downtime, alarms, and analytics, on top of
Go, QuestDB, PostgreSQL, and a React dashboard.

For a full file-by-file breakdown of the codebase, see [`CODEBASE_AUDIT.md`](CODEBASE_AUDIT.md).
For architecture decisions, see [`docs/adr/`](docs/adr/).

## Architecture

```
                 ┌────────────────────────┐
   PLC Network ─►│  PLC Drivers (planned)  │
                 └───────────┬─────────────┘
                             │
   Simulator ────────────────┤   (make simulate — no PLCs needed)
                             ▼
                 ┌────────────────────────┐
                 │  telemetry.Tracker      │  derives alarms / run-state /
                 │  (producer-agnostic)    │  production-runs / downtime
                 └───────────┬─────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼                             ▼
        QuestDB                        PostgreSQL
   (raw samples, alarms,          (machines, tags, production_runs,
    machine_state, rollups)       downtime_events, controls, alarm_acks)
              │                             │
              └──────────────┬──────────────┘
                              ▼
                        Go API Server
                              │
                              ▼
                   React SPA / Embedded Dashboard
```

**Key design decisions** (see `docs/adr/` for the full record):
- QuestDB for high-ingestion-rate time-series storage with built-in materialized rollups
- PostgreSQL for relational/business data (machines, tags, production runs, OEE targets, controls)
- Human-readable identity columns (`machine_id`/`machine_name`/`tag_name`) on every telemetry row — no cross-database joins for display (ADR-0018)
- A single producer-agnostic derivation pipeline (`internal/telemetry.Tracker`) turns raw tag samples into alarms/production-runs/downtime regardless of whether the samples came from the simulator or a real driver
- Protocol-agnostic `plc.Driver` interface (`internal/plc/driver.go`) — real protocol drivers are not yet implemented (see Status below)
- Single-binary deployment: the compiled SPA is embedded into the Go binary via `go:embed` (ADR-0013)

## Quick Start

```bash
./run.sh          # installs dependencies, starts Docker infra, prompts for a run mode
```

Or manually:

```bash
make up          # start Postgres + QuestDB
make simulate    # build + run dev-mode (simulated telemetry) + dashboard
```

Open `http://localhost:5173` (Vite dev server, proxies API calls to `:8081`) or
`http://localhost:8081` directly.

## Run modes

| Command | Binary | Data source | Pre-configured machines? |
|---|---|---|---|
| `make simulate` | `cmd/dev-mode` | Built-in simulation engine (`internal/simulation`) — realistic per-machine production, faults, and recovery | Yes (11 seeded machines/tags) |
| `make real` | `cmd/pharma-platform` | None yet — no driver-polling pipeline is wired up (see Status) | No — empty database, dashboard is ready but has nothing to show until machines/drivers exist |

`make simulate` is the mode to use to see the product working end to end: realistic,
differentiated OEE per machine, alarms, production runs, and downtime events, all derived
from a simulated but structurally-real telemetry stream. See
[`CODEBASE_AUDIT.md` §1](CODEBASE_AUDIT.md#1-how-the-run-modes-work) for exactly how each
mode boots and what it wires together.

## Other Make targets

| Command | Purpose |
|---|---|
| `make up` | Start Postgres + QuestDB containers |
| `make up-all` | Start Postgres + QuestDB + the containerized app |
| `make down` | Stop all containers |
| `make logs` | Tail container logs |
| `make migrate` | Run schema migrations only |
| `make seed` | Run schema migrations + seed data, no server |
| `make build` | `go build ./...` |

## API surface

Full endpoint list in [`CODEBASE_AUDIT.md` §4](CODEBASE_AUDIT.md#4-full-endpoint-reference).
Summary: telemetry (`/telemetry/*`), machine/PLC/tag config (`/plcs`, `/tags`,
`/api/v1/machines`), alarms (`/alarms*`), controls (`/api/v1/controls/*`), production/OEE/
downtime (`/api/v1/{production,oee,downtime}*`), and the business analytics layer
(`/api/v2/analytics/*` — overview, production, quality, machines, energy, alarms,
correlations, maintenance, insights).

## Storage model

**PostgreSQL** — relational/business data: `machines`, `tags`, `production_runs`,
`downtime_events`, `oee_targets`, `machine_control_state`, `alarm_acks`.

**QuestDB** — time-series data: `plc_samples` (raw telemetry, partitioned by day),
`alarms`, `events`, `logs`, `production_counts`, `machine_state`, plus materialized rollup
views `plc_samples_1m`/`1h`/`1d`/`1w`.

## Configuration

Runtime config lives in `project/config/bootstrap.yaml`:

```yaml
postgres:
  host: localhost
  port: 5433   # mapped off the default 5432 to avoid clashing with a local Postgres install
  database: pharma
  user: postgres
  password: postgres

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

plant:
  name: Pharma Platform
  location: Manufacturing Facility
  timezone: Asia/Kolkata
```

## Development

**Prerequisites:** Go, Docker + Docker Compose, Node.js/npm (for the frontend dev server).

```bash
make up                              # infra
cd project && go build ./...         # backend build check
cd web && npm install && npm run dev # frontend dev server (if not using `make simulate`)
```

Backend tests: `cd project && go test ./...`

## Status

This project has a working simulated demo (`make simulate`) with a full dashboard, real
OEE/production/alarm calculation, and a real backend for everything except energy metrics
(no power/vibration sensor data exists in the schema, so that section is explicitly flagged
as synthetic in its API response).

Real PLC connectivity is partially built: the driver interface (`internal/plc/driver.go`)
is defined, the OPC UA driver (`internal/plc/drivers/opcua`) is fully implemented and unit
tested, and a driver registry (`internal/plc/registry`) + connection manager
(`internal/plc/manager.go`) exist to construct and own driver instances — OPC UA is wired,
every other protocol returns an explicit "not implemented" error rather than doing nothing
silently. What's still missing: a real polling collector wired to the registry, and a
dashboard-driven machine-configuration UI to actually connect `make real` to equipment. See
[`CODEBASE_AUDIT.md` §5](CODEBASE_AUDIT.md#5-known-gaps-and-what-changed-this-session) for
the full list of known gaps, and `docs/roadmap.md` for planned work.

## License

There is no `LICENSE` file in this repository — no license terms have been finalized.
Do not treat this project as licensed for reuse until that's resolved.
