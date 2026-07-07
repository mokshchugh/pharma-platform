# ADR-0004: Persistent Data Storage Strategy

**Status:** Accepted (Updated 2026-07-07)

**Date:** 2026-06-28

## Context

The platform stores production telemetry, aggregated manufacturing metrics, application metadata, and future dashboard configurations.

## Decision

Persistent application data will be stored using host bind mounts in a `persistent/` directory at the project root. The directory is git-tracked as a skeleton (empty subdirectories) and ignored by `go build ./...` by virtue of `go.mod` living inside `project/`.

```
pharma-platform/
├── persistent/           # git-tracked skeleton, docker bind-mounts
│   ├── postgres/         → /var/lib/postgresql/data
│   └── questdb/          → /var/lib/questdb
├── project/              # Go module root (go.mod lives here)
│   └── runtime/
│       └── docker-compose.yml
└── Makefile              # make up starts docker, make dev runs Go
```

### Docker Compose Integration

```yaml
services:
  postgres:
    volumes:
      - ../../persistent/postgres:/var/lib/postgresql/data

  questdb:
    volumes:
      - ../../persistent/questdb:/var/lib/questdb
```

## Rationale

Binding data directories inside the project root as git-tracked skeletons keeps the setup self-contained while ensuring data survives container rebuilds. The Go module is nested under `project/` to prevent Docker-created file permissions from interfering with Go tooling.

## Consequences

### Positive
* Persistent storage across redeployments
* Self-contained project setup
* No external directory management
* `go build ./...` never conflicts with Docker UIDs

### Negative
* Running Go commands requires `cd project` or the root Makefile
