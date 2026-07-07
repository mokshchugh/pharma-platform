# ADR-0005: Containerized Deployment Using Docker Compose

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-06-28

## Decision

Docker Compose for container orchestration. Compose file at `project/runtime/docker-compose.yml`.

### Services

* PostgreSQL (official `postgres:17-alpine`)
* QuestDB (official `questdb/questdb:9.1.0`)
* Application (built from `project/runtime/docker/Dockerfile`)

### Storage

Host bind mounts from `persistent/postgres/` and `persistent/questdb/`. See ADR-0004.

### Networking

Containers communicate over Docker's internal network. External ports published for database access during development and API server for dashboard access.

## Consequences

### Positive
* Reproducible deployments
* Persistent data independent of container lifecycle

### Negative
* Docker runtime is a deployment dependency
