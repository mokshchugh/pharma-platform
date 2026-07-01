# ADR-0002: Use Go for Backend Services

**Status:** Accepted

**Date:** 2026-06-28

## Context

The platform requires long-running backend services responsible for continuously collecting telemetry from industrial PLCs, writing high-frequency data to QuestDB, aggregating historical data into PostgreSQL, and exposing APIs for dashboards and analytics.

The initial implementation language considered was Python due to its extensive ecosystem, mature industrial communication libraries, rapid development experience, and familiarity within the team.

However, the Collector is expected to operate as a continuously running production service where reliability, predictable performance, deployment simplicity, and efficient concurrency are more important than rapid scripting.

## Decision

Use Go for all backend services, including:

* PLC Collector
* Aggregation Service
* REST API
* PLC Simulator
* Future background workers

## Alternatives Considered

### Python

Pros

* Rich ecosystem for industrial protocols
* Excellent data processing libraries
* Faster prototyping
* Large community support

Cons

* Higher memory usage
* Interpreter dependency
* Less predictable concurrency for long-running services
* Packaging and deployment complexity compared to a static binary

### Rust

Pros

* Excellent performance
* Strong memory safety
* Very low resource usage

Cons

* Steeper learning curve
* Longer development time
* Smaller ecosystem for industrial protocols
* Increased implementation complexity for the project's timeline

### Go (Selected)

Pros

* Excellent concurrency using goroutines
* Compiles to a single static binary
* Low memory footprint
* Fast startup
* Strong networking support
* Simple deployment in Docker
* Easy cross-compilation
* Well suited for continuously running backend services

Cons

* Smaller ecosystem for scientific computing
* Fewer industrial automation libraries than Python
* Simpler language with fewer abstraction features

## Rationale

Although Python provides a richer ecosystem for industrial communication, the primary workload of this platform is continuous network communication, scheduling, buffering, and database interaction rather than scientific computation.

Go's concurrency model, deployment simplicity, and operational characteristics make it a better fit for industrial data acquisition services.

## Consequences

### Positive

* Lightweight Docker images
* Single executable deployments
* Efficient concurrent polling of multiple PLCs
* Lower runtime resource consumption
* Consistent language across all backend services

### Negative

* Some protocol implementations may require more engineering effort.
* Smaller ecosystem compared to Python for certain industrial integrations.

## Future Considerations

Python remains a suitable choice for future machine learning, predictive maintenance, or offline analytics components if those become part of the platform.
