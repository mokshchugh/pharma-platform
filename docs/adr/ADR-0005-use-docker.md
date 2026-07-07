# ADR-0005: Containerized Deployment Using Docker Compose

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-06-28

## Context

The platform consists of multiple independent services including time-series storage, relational storage, backend services, and future visualization and monitoring components.

## Decision

Adopt Docker as the containerization platform and Docker Compose as the deployment orchestrator.

Each major component will execute as an independent container with a single responsibility.

### Compose File Location

The Docker Compose file lives at `runtime/docker-compose.yml` (previously `deploy/compose.yaml`) to keep deployment artifacts alongside their Dockerfiles:

```
pharma-platform/
    runtime/
        docker/
            Dockerfile
            entrypoint.sh
        docker-compose.yml
```

### Infrastructure Services

* PostgreSQL (official `postgres:17-alpine`)
* QuestDB (official `questdb/questdb:9.1.0`)
* Application (built from `runtime/docker/Dockerfile`)

### Storage Strategy

Host bind mounts from `persistent/postgres/` and `persistent/questdb/` provide data persistence. See ADR-0004 for details.

### Networking

Containers communicate over Docker's internal network using service discovery. External ports are only published for database access during development and the API server for dashboard access.

## Alternatives Considered

Native Host Installation, Kubernetes. Docker Compose selected for simplicity and sufficient orchestration for single-host deployments.

## Consequences

### Positive

* Reproducible deployments
* Simplified onboarding
* Isolated services
* Persistent data independent of container lifecycle

### Negative

* Docker runtime is a deployment dependency
* Multi-host scaling would require different orchestration
