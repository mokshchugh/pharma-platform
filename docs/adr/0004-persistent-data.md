# ADR-0004: Persistent Data Storage Strategy

**Status:** Accepted

**Date:** 2026-06-28

## Context

The platform stores production telemetry, aggregated manufacturing metrics, application metadata, and future dashboard configurations.

These datasets are operationally critical and must survive:

* Application restarts
* Container recreation
* Docker image upgrades
* Host system reboots
* Application redeployments

The storage strategy must also simplify backup, inspection, migration, and disaster recovery.

## Decision

Persistent application data will be stored outside the project repository using host bind mounts.

Application source code and operational data will remain physically separated.

Project layout:

```text
Projects/

pharma-platform/
    Source Code
    Documentation
    Configuration
    Deployment

pharma-platform-data/
    PostgreSQL/
    QuestDB/
    Grafana/
    Logs/
    Backups/
```

Each service is assigned its own dedicated storage directory.

Example:

```yaml
volumes:
  - ${DATA_ROOT}/postgres:/var/lib/postgresql

volumes:
  - ${DATA_ROOT}/questdb:/var/lib/questdb
```

## Alternatives Considered

### Container Writable Layer

Pros

* No additional configuration

Cons

* Data lost when containers are removed
* Unsuitable for databases
* Difficult backup process

---

### Docker Named Volumes

Pros

* Docker-managed
* Easy to create
* Portable across Compose deployments

Cons

* Data location abstracted by Docker
* Less transparent to administrators
* More difficult to inspect manually
* Backup and migration require Docker tooling

---

### Host Bind Mounts (Selected)

Pros

* Data location fully controlled by administrators
* Easy inspection using standard operating system tools
* Straightforward backup and restore procedures
* Data survives container recreation
* Independent of Docker's internal storage implementation
* Clear separation between application code and operational data

Cons

* Host directory structure must be managed
* Requires correct filesystem permissions
* Slightly less portable between hosts without adjusting paths

## Rationale

The platform is intended for long-running industrial deployments where operational reliability and data integrity take precedence over deployment convenience.

Separating application code from operational data allows:

* Independent application upgrades
* Simple disaster recovery
* Transparent filesystem organization
* Easier system administration
* Predictable backup procedures

The chosen strategy aligns with the principle that application code is replaceable, while production data is not.

## Consequences

### Positive

* Persistent storage across redeployments
* Simple backup and restore operations
* Improved operational transparency
* Easier migration to new hardware
* Reduced risk of accidental data loss
* Independent lifecycle for code and data

### Negative

* Additional host directory management
* Storage paths become deployment-specific
* Backup responsibility remains with system administrators

## Future Considerations

As the platform evolves, additional persistent directories may be introduced for:

* Centralized logs
* Grafana dashboards
* Configuration snapshots
* Historical backups
* Machine learning datasets

The separation between application code and operational data should be maintained regardless of the deployment platform, including future migration to Kubernetes or cloud infrastructure.
