# ADR-0005: Containerized Deployment Using Docker Compose

**Status:** Accepted

**Date:** 2026-06-28

## Context

The platform consists of multiple independent services including time-series storage, relational storage, backend services, and future visualization and monitoring components.

The system should be:

* Easy to deploy on a new machine
* Reproducible across development and production environments
* Simple to update
* Isolated from host operating system dependencies
* Capable of persisting data across container recreation and host reboots

A deployment strategy was therefore required before implementing application services.

## Decision

Adopt Docker as the containerization platform and Docker Compose as the deployment orchestrator.

Each major component will execute as an independent container with a single responsibility.

Initial infrastructure services include:

* PostgreSQL
* QuestDB

Future services include:

* Collector
* Aggregator
* API
* Grafana
* PLC Simulator

The complete platform will be deployed using a single Compose configuration.

## Alternatives Considered

### Native Host Installation

Pros

* No container runtime
* Direct access to host operating system

Cons

* Difficult dependency management
* Environment inconsistencies
* Complex upgrades
* Manual installation process
* Reduced deployment portability

---

### Kubernetes

Pros

* Highly scalable
* Advanced orchestration
* Self-healing workloads
* Enterprise deployment capabilities

Cons

* Significantly more operational complexity
* Unnecessary for the initial deployment size
* Higher learning and maintenance overhead

---

### Docker Compose (Selected)

Pros

* Simple deployment
* Infrastructure-as-Code
* Easy local development
* Consistent runtime environment
* Widely supported
* Straightforward migration to Kubernetes if required in the future

Cons

* Limited orchestration features compared to Kubernetes
* Intended primarily for single-host deployments

## Storage Strategy

Persistent application data will not be stored inside Docker-managed volumes.

Instead, host bind mounts will be used.

Example:

```yaml
${DATA_ROOT}/postgres:/var/lib/postgresql
```

Reasons:

* Data remains accessible outside Docker.
* Backup procedures are simplified.
* Data survives container recreation.
* Storage location remains under administrative control.
* Easier inspection and disaster recovery.

The repository and operational data remain physically separated.

Example:

```
Projects/

pharma-platform/
    Source code
    Documentation
    Deployment configuration

pharma-platform-data/
    PostgreSQL
    QuestDB
    Grafana
    Logs
    Backups
```

This separation allows the application repository to be updated, replaced, or re-cloned without affecting production data.

## Networking Strategy

Containers communicate over Docker's internal network using service discovery rather than fixed IP addresses.

Example:

```
Collector
      │
      ▼
pharma-questdb:9000
```

rather than

```
192.168.x.x
```

External ports are only published when access from outside Docker is required.

Internal communication remains isolated within the Docker network.

## Service Design

Each container is responsible for a single service.

Examples include:

* PostgreSQL
* QuestDB
* Collector
* Aggregator
* API
* Grafana

This improves:

* Isolation
* Maintainability
* Independent upgrades
* Debugging
* Fault isolation

## Rationale

Containerization provides a consistent runtime environment independent of the host operating system while simplifying deployment and future maintenance.

Docker Compose provides sufficient orchestration capabilities for the expected deployment scale while avoiding the operational overhead of Kubernetes.

Bind mounts align with the project's emphasis on data persistence, backup simplicity, and operational transparency.

## Consequences

### Positive

* Reproducible deployments
* Simplified onboarding
* Isolated services
* Persistent data independent of container lifecycle
* Infrastructure version-controlled alongside source code
* Easy migration between development and production systems

### Negative

* Docker runtime becomes a deployment dependency.
* Additional networking layer between services.
* Multi-host scaling would require a different orchestration platform.

## Future Considerations

As the platform evolves, Docker Compose may be supplemented or replaced by Kubernetes or another orchestration platform if deployment requirements expand beyond a single industrial server.

The internal service architecture is intentionally designed so that migration to a more advanced orchestrator requires minimal changes to the application itself.
