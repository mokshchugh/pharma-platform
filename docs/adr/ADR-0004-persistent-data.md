# ADR-0004: Persistent Data Storage Strategy

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-06-28

## Context

The platform stores production telemetry, aggregated manufacturing metrics, application metadata, and future dashboard configurations.

These datasets are operationally critical and must survive:

* Application restarts
* Container recreation
* Docker image upgrades
* Host system reboots
* Application redeployments

## Decision

Persistent application data will be stored using host bind mounts in a `persistent/` directory at the project root.

The `persistent/` directory is gitignored and contains subdirectories for each database service:

```text
pharma-platform/
    Source Code
    Documentation
    Configuration
    Deployment

pharma-platform/persistent/
    postgres/       → PostgreSQL PGDATA
    questdb/        → QuestDB data files
```

The `persistent/` directory lives inside the project root but is excluded from version control. This keeps the setup self-contained while ensuring data survives container rebuilds.

### Docker Compose Integration

```yaml
services:
  postgres:
    volumes:
      - ../persistent/postgres:/var/lib/postgresql/data

  questdb:
    volumes:
      - ../persistent/questdb:/var/lib/questdb
```

## Alternatives Considered

### External Data Directory

Previously the design specified a separate `pharma-platform-data/` directory outside the project root. This was changed to `persistent/` inside the project root to:

- Simplify setup (single clone + `docker compose up`)
- Keep related data co-located with the project
- Reduce configuration surface area

### Docker Named Volumes

Pros: Docker-managed, easy to create, portable across Compose deployments.

Cons: Data location abstracted by Docker, less transparent, backup requires Docker tooling.

### Host Bind Mounts (Selected)

Pros: Data location fully controlled, easy inspection, straightforward backup, survives container recreation, clear separation between code and data.

## Rationale

Binding data directories inside the project root as gitignored directories provides the best balance of simplicity and data integrity. The entire platform can be deployed with a single `docker compose up` without pre-creating external directories.

## Consequences

### Positive

* Persistent storage across redeployments
* Simple backup and restore operations
* Self-contained project setup
* No external directory management

### Negative

* `persistent/` must be gitignored
* Large data directories could slow git status operations
