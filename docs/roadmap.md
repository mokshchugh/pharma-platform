# Development Roadmap

## Phase 1 — Core Models & Configuration (done)
- [x] Domain models (PLC, Tag, Sample, DataType, Quality)
- [x] Bootstrap configuration (project/config/bootstrap.yaml)
- [x] Config validation

## Phase 2 — Storage Layer (done)
- [x] PostgreSQL schema (machines, tags tables)
- [x] QuestDB schema (plc_samples, alarms, events, logs)
- [x] Migration runner (auto-create tables on startup + standalone `cmd/migrate`)
- [x] Separation of Postgres schema DDL from seed data (init/ vs seed/)
- [x] Seed data from plant inventory (11 machines, 128 tags)
- [x] PostgreSQL-backed MachineStore and TagStore

## Phase 3 — Infrastructure (done)
- [x] Docker Compose (project/runtime/docker-compose.yml)
- [x] Dockerfile (project/runtime/docker/Dockerfile)
- [x] Persistent storage (persistent/, bind mounts)
- [x] Makefile (root, wraps all commands)

## Phase 4 — Collector (done)
- [x] Collector architecture (scheduler + worker pool)
- [x] Buffered sample channel
- [x] QuestDB batch writer (ILP over TCP)
- [x] Pause/resume (atomic bool + SIGUSR1/SIGUSR2)
- [x] Graceful shutdown protocol

## Phase 5 — API & Dashboard (done)
- [x] REST API (18 endpoints)
- [x] Embedded SPA dashboard (vanilla HTML/JS)
- [x] React SPA frontend (separate web/ directory)
- [x] Dev-mode with mock collector
- [x] Standalone API binary
- [x] Standalone simulator binary
- [x] Standalone seed binary
- [x] Standalone migrate binary

## Phase 6 — Identity Schema Refactor (done)
- [x] Replace plc_id/tag_id with machine_id/machine_name/tag_name in QuestDB
- [x] Update Go models, ILP encoder, SQL queries
- [x] Update reader, writer, aggregator for new schema
- [x] Separate Postgres seed data from schema DDL
- [x] Preserve API backward compatibility for URL params
- [x] Document decision in ADR-0018

## Phase 7 — PLC Driver Development (next sprint)
- [ ] MC Protocol (SLMP 3E Frame) driver
- [ ] FINS/TCP driver
- [ ] EtherNet/IP (CIP) driver
- [ ] Complete OPC UA driver (Read/Close)
- [ ] Multi-driver collector support

## Phase 8 — Real PLC Integration
- [ ] Machine configuration via dashboard/API
- [ ] Connection health monitoring
- [ ] Automatic reconnection
- [ ] Connectivity dashboard

## Phase 9 — Advanced Features
- [ ] Aggregation service (OEE, shift reports)
- [ ] Alarm management
- [ ] User authentication & authorization
- [ ] Audit logging
- [ ] Production reporting
