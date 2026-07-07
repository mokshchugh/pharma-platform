# Software Requirements Specification (SRS)

## Pharmaceutical Industrial Data Acquisition & Analytics Platform

**Version:** 0.2 (Architecture Update)

**Status:** Implementation Phase

---

# 1. Introduction

## 1.1 Purpose

This document defines the software architecture, requirements, and design decisions for an Industrial Data Acquisition Platform intended for deployment in pharmaceutical manufacturing facilities.

---

# 2. Technology Stack

## Backend
* Go

## Databases
* QuestDB (time-series telemetry)
* PostgreSQL (persistent configuration and business data)

## Frontend
* Vanilla HTML/JS/CSS (embedded via `//go:embed`)

## Infrastructure
* Docker
* Docker Compose

---

# 3. High-Level Architecture

```
                  Users
                     |
              HTTPS / TLS
                     |
                Go API Server
          ┌──────────┴──────────┐
          │                     │
          ▼                     ▼
      QuestDB             PostgreSQL
          ▲
          │
    PLC Collector
          ▲
          │
        PLC Network
```

## Database Responsibilities

### QuestDB
- Raw telemetry (`plc_samples`)
- Time-series events (`alarms`, `events`, `logs`)
- High-frequency sensor values

### PostgreSQL
- Machine registry (`machines` table)
- Tag definitions (`tags` table)
- Users, roles, permissions (future)
- Batch information (future)
- Aggregated KPIs (future)

---

# 4. Core Components

## Entry Points

| Binary | Purpose | Postgres | QuestDB | Seed | Collector | API |
|---|---|---|---|---|---|---|
| `cmd/pharma-platform` | Production | Schema | Tables | No | Idle | Yes |
| `cmd/dev-mode` | Development | Schema+Seed | Tables | If empty | Mock | Yes |
| `cmd/api` | Dashboard only | Schema+Seed | Tables | If empty | Stub | Yes |
| `cmd/collector-sim` | Simulator | Read tags | Tables | No | Mock→QuestDB | No |
| `cmd/seed` | Standalone seed | Schema+Seed | No | Always | No | No |

## Configuration

Single `config/bootstrap.yaml` contains all settings:

```yaml
postgres:    # Connection + pool config
questdb:     # Connection + batch config
api:         # HTTP server host/port
collector:   # Workers, queue size
aggregator:  # Aggregation interval
plant:       # Name, location, timezone
```

---

# 5. Data Flow

```
Seed Flow (dev-mode, api, seed):
    deploy/postgres/init/001_schema.sql
    → PostgreSQL (machines + tags tables)
    → deploy/postgres/init/002_seed_machines.sql
    → deploy/postgres/init/003_seed_tags.sql

Migration Flow (all binaries):
    deploy/questdb/init/001_plc_samples.sql
    → QuestDB REST /exec
    → deploy/questdb/init/002_events.sql

Read Flow (all binaries with API):
    Frontend → API → QuestDB REST (telemetry)
    Frontend → API → PostgreSQL (machines, tags)
```

---

# 6. Repository Pattern

```
HTTP Handler (internal/api/handlers/)
    ↓
Store Interface (PLCStore, TagStore)
    ↓
PostgreSQL Implementation (internal/store/)
    ↓
Database
```

---

# 7. Deployment Model

```yaml
services:
  postgres:  # official postgres:17-alpine
  questdb:   # official questdb:9.1.0
  app:       # built from runtime/docker/Dockerfile
```

Persistent storage in `persistent/` (gitignored, bind-mounted).

---

# 8. Non-Functional Requirements

- 24×7 operation
- Continuous telemetry ingestion
- Automatic restart via Docker
- Health monitoring
- Modular services

---

# 9. Current Design Decisions (ADR Summary)

1. ADR-0001: QuestDB for time-series
2. ADR-0002: Go for backend
3. ADR-0003: PostgreSQL for business data
4. ADR-0004: `persistent/` directory for data
5. ADR-0005: Docker Compose at `runtime/docker-compose.yml`
6. ADR-0007: Protocol-agnostic PLC driver interface
7. ADR-0008: Collector with scheduler + worker pool
8. ADR-0011: 18-endpoint REST API
9. ADR-0012: Dashboard API v1
10. ADR-0013: Embedded SPA frontend
11. ADR-0014: Collector pause/resume
12. ADR-0015: Dev-mode with DB-backed mock data
13. ADR-0016: PostgreSQL store for machines and tags
