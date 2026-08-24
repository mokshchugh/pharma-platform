# Software Requirements Specification (SRS)

## Pharmaceutical Industrial Data Acquisition & Analytics Platform

**Version:** 0.4 (Post-cleanup)

**Status:** Implementation Phase

For a complete, file-by-file inventory, see [`../../CODEBASE_AUDIT.md`](../../CODEBASE_AUDIT.md).
This document stays high-level; the audit doc is the source of truth for "what exists."

---

# 1. Technology Stack

* Go (module root at `project/`)
* QuestDB (time-series telemetry)
* PostgreSQL (persistent configuration and business data)
* React SPA (`web/`), also embedded into the Go binary for single-binary deployment
* Docker + Docker Compose

# 2. Repository Layout

See [`../../CODEBASE_AUDIT.md` §2](../../CODEBASE_AUDIT.md#2-repository-layout) for the
current tree — not duplicated here to avoid the two copies drifting apart.

# 3. High-Level Architecture

```
                   Users
                      |
           ┌──────────┴──────────┐
           │                     │
           ▼                     ▼
      Go API Server         React SPA (dev) / embedded SPA (prod)
           │                     │
           └──────────┬──────────┘
                      │
              ┌───────┴───────┐
              │               │
              ▼               ▼
          QuestDB         PostgreSQL
              ▲
              │
   telemetry.Tracker (derives alarms/run-state/production/downtime)
              ▲
              │
   Simulator (make simulate) or a PLC driver pipeline (planned, make real)
```

# 4. Database Responsibilities

### QuestDB (project/deploy/questdb/init/)
- `plc_samples` — raw telemetry (`machine_id`, `machine_name`, `tag_name`, `value`, `quality`)
- `plc_samples_1m` / `_1h` / `_1d` / `_1w` — materialized rollup views
- `alarms` — alarm events
- `events` — batch and machine events
- `logs` — system logs
- `production_counts`, `machine_state` — per-machine production/state snapshots

### PostgreSQL (project/deploy/postgres/init/)
- `machines`, `tags` — machine/PLC inventory and tag definitions
- `production_runs`, `downtime_events`, `oee_targets` — OEE/production tracking
- `machine_control_state` — start/stop/setpoint/mode control state
- `alarm_acks` — alarm acknowledgement state (QuestDB has no UPDATE, so acks live here and get joined at query time)

# 5. Entry Points (project/cmd/)

| Binary | Postgres | QuestDB | Seed | Telemetry source | API |
|---|---|---|---|---|---|
| `pharma-platform` (`make real`) | Schema | Tables | No | None yet — no machines configured | Yes |
| `dev-mode` (`make simulate`) | Schema+Seed | Tables | If empty | Built-in simulation engine | Yes |
| `seed` | Schema+Seed | No | Always | — | No |
| `migrate` | Schema only | Tables | No | — | No |

All invoked via `make` (or `./run.sh`) from the repository root.

# 6. Configuration

Single file: `project/config/bootstrap.yaml`

```yaml
postgres:    # host, port, database, user, password
questdb:     # host, port, batch_size, flush_interval
api:         # host, port
collector:   # workers, queue_size
plant:       # name, location, timezone
```

# 7. Current Design Decisions (ADR Summary)

1. ADR-0001: QuestDB for time-series
2. ADR-0002: Go for backend
3. ADR-0003: PostgreSQL for business data
4. ADR-0004: `persistent/` directory, `project/` for go module
5. ADR-0005: Docker Compose at `project/runtime/docker-compose.yml`
6. ADR-0006: OPC UA-first driver strategy
7. ADR-0007: Protocol-agnostic PLC driver interface
8. ADR-0008: Collector with scheduler + worker pool (design decision retained for when real driver polling is rebuilt; the original implementation was removed as dead code alongside `cmd/collector-sim` — see roadmap Phase 4)
9. ADR-0011: Initial REST API
10. ADR-0012: Dashboard API v1
11. ADR-0013: Embedded SPA frontend
12. ADR-0014: Collector pause/resume
13. ADR-0015: Dev-mode with DB-backed mock data
14. ADR-0016: PostgreSQL store for machines and tags
15. ADR-0017: Bootstrap configuration
16. ADR-0018: Identity field refactoring (plc_id/tag_id → machine_id/machine_name/tag_name)

# 8. Identity Model

Telemetry samples in QuestDB use the following identity columns:

- `machine_id` (SYMBOL) — stable machine identifier e.g. `"1"`
- `machine_name` (SYMBOL) — human-readable machine name e.g. `"Fluid Bed Dryer"`
- `tag_name` (SYMBOL) — technical tag name e.g. `"Inlet_Air_Temp"`

The API surface still exposes URL parameters as `plc_id` and `tag_id` for backward compatibility, but maps them internally to `machineID` and `tagName`.

# 9. Data Flow

```
Simulator (internal/simulation)          PLC Driver (planned: internal/plc/registry
   or, eventually, a real driver              + internal/plc/drivers/opcua, others TBD)
              │                                          │
              └───────────────────┬──────────────────────┘
                                   ▼
                       models.Sample{MachineID, MachineName, TagName, Value, Quality}
                                   │
                                   ▼
                  telemetry.Tracker.Observe() — derives alarms, machine_state,
                  production_runs, downtime_events from the raw sample stream
                                   │
                                   ▼
                  QuestDB Writer (ILP over TCP, double-buffered, bounded retry)
                                   │
                                   ▼
                  QuestDB (plc_samples + materialized views)  +  PostgreSQL
                                   │
                                   ├──► Reader (REST HTTP API) ──► Go API Server ──► Dashboard
                                   │
                                   └──► business.RealEngine (/api/v2/analytics/*)
```

Both `internal/simulation.Simulator` and any future real driver pipeline feed the
*same* `telemetry.Tracker`, so simulated runs exercise the real alarm/OEE/production
derivation logic rather than a simulation-only shortcut.
