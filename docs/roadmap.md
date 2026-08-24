# Development Roadmap

For a full inventory of what currently exists, see [`../CODEBASE_AUDIT.md`](../CODEBASE_AUDIT.md).

## Phase 1 — Core Models & Configuration (done)
- [x] Domain models (PLC, Tag, Sample, DataType, Quality)
- [x] Bootstrap configuration (project/config/bootstrap.yaml)
- [x] Config validation

## Phase 2 — Storage Layer (done)
- [x] PostgreSQL schema (machines, tags, production_runs, downtime_events, oee_targets, machine_control_state, alarm_acks)
- [x] QuestDB schema (plc_samples, alarms, events, logs, production_counts, machine_state, rollup views)
- [x] Migration runner (auto-run on startup + standalone `cmd/migrate`)
- [x] Seed data (11 machines, ~100 tags)
- [x] PostgreSQL-backed MachineStore, TagStore, ProductionStore, ControlStore, AlarmAckStore

## Phase 3 — Infrastructure (done)
- [x] Docker Compose (project/runtime/docker-compose.yml)
- [x] Dockerfile (project/runtime/docker/Dockerfile)
- [x] Persistent storage (persistent/, bind mounts)
- [x] Makefile + `run.sh` single entrypoint

## Phase 4 — Simulation & telemetry derivation (done)
- [x] Per-machine simulation engine with fault/recovery scheduling (`internal/simulation`)
- [x] QuestDB batch writer (ILP over TCP)
- [x] Producer-agnostic derivation pipeline (`internal/telemetry.Tracker`) — turns raw samples into alarms, machine state, production runs, and downtime, regardless of whether the samples came from the simulator or (eventually) a real driver
- [x] Pause/resume (SIGUSR1/SIGUSR2)

Note: the earlier `internal/collector` scheduler+worker-pool package and its
`cmd/collector-sim` test harness were removed as dev-only/orphaned code once
`internal/simulation` + `internal/telemetry.Tracker` replaced them for the
simulate mode. A polling collector will be rebuilt against the driver
registry (Phase 7/8) when real PLC ingestion is implemented — it doesn't
need to look like the old one.

## Phase 5 — API & Dashboard (done)
- [x] Full REST API (telemetry, machines/PLCs/tags, alarms, controls, production/OEE/downtime, `/api/v2/analytics/*` business analytics)
- [x] Embedded SPA (single-binary deployment)
- [x] React SPA frontend (`web/`)
- [x] Two run modes: `make simulate` (dashboard + simulation engine) and `make real` (dashboard + backend, no pre-configured machines)

## Phase 6 — Identity Schema Refactor (done)
- [x] `machine_id`/`machine_name`/`tag_name` columns across QuestDB instead of opaque `plc_id`/`tag_id`
- [x] API backward compatibility for URL params
- [x] Documented in ADR-0018

## Phase 7 — PLC Driver Development (in progress)
- [x] OPC UA driver: `Connect`/`Read`/`Close` all implemented and unit-tested (`internal/plc/drivers/opcua`)
- [x] Driver registry (`internal/plc/registry`) and connection `Manager` (`internal/plc/manager.go`) — OPC UA wired, other protocols return explicit "not implemented" errors rather than silently doing nothing
- [ ] MC Protocol (SLMP 3E Frame) driver
- [ ] FINS/TCP driver
- [ ] EtherNet/IP (CIP) driver
- [ ] S7 driver
- [ ] A real polling collector wired to the registry, feeding `internal/telemetry.Tracker` the same way the simulator does

## Phase 8 — Real PLC Integration
- [ ] Machine configuration via dashboard/API (create a machine, pick a driver, set connection details)
- [ ] Connection health monitoring + automatic reconnection
- [ ] Wire `make real` to actually ingest from configured machines

## Phase 9 — Advanced Features
- [x] OEE calculation, production run/downtime tracking (`internal/store/production.go`, `internal/business`)
- [x] Alarm derivation + acknowledgement (`internal/telemetry.Tracker`, `internal/store/alarm_acks.go`)
- [ ] User authentication & authorization
- [ ] Audit logging
- [ ] Production reporting/export beyond the existing CSV telemetry export
