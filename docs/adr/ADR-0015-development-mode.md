# ADR-0015: Development Mode Architecture

**Status:** Accepted

**Date:** 2026-07-07

## Context

The production platform (`cmd/pharma-platform/main.go`) requires real YAML configuration files defining PLCs, tags, plant metadata, and database connections. It also uses the OPC UA protocol driver, which connects to physical PLCs over the network.

During development and demonstration, no real PLCs are available. Running the full configuration pipeline requires creating mock YAML files that would need to be kept in sync with the real schema. Additionally, the OPC UA driver panics on Read/Close when no PLC is reachable.

A separate development entry point is needed that:

* Works without any configuration files.
* Generates simulated telemetry data for dashboard testing.
* Wires all components (collector, QuestDB writer, API server, frontend) into a single process.
* Supports pause/resume and graceful shutdown for developer convenience.

## Decision

Create `cmd/dev-mode/main.go` — a standalone binary that hardcodes all development configuration and uses a mock PLC driver.

### Architecture

```
cmd/dev-mode/main.go
    │
    ├── mockDriver (inline, implements plc.Driver)
    ├── 3 hardcoded PLCs (plc-1, plc-2, plc-3) with 25 tags each (75 total)
    ├── questdb.Writer → QuestDB at localhost:9009
    ├── api.NewFull → HTTP server on :8081 with embedded frontend
    ├── Collector pause/resume via SIGUSR1/SIGUSR2
    └── Graceful shutdown: collector stop → close samples → writer drain
```

### Mock Driver

The mock driver is defined in `main.go` (package `main`), not in the shared `internal/plc/drivers/` package. It returns a sine-wave value (`42 + sin(time) × 10`) that oscillates over time, making it easy to verify that queries return changing data.

### PLCs and Tags

| PLC ID | Name | Tags | Poll Interval |
|--------|------|------|-------|
| `plc-1` | Fluid Bed Dryer | 25 | 100 ms |
| `plc-2` | Tablet Press | 25 | 100 ms |
| `plc-3` | HVAC System | 25 | 100 ms |

Each tag has a unique ID following the pattern `{plc_id}-tag-{NN}` for easy querying by tag ID or prefix.

### Data Flow

```
mockDriver.Read()
    → Collector (scheduler + 16 workers)
    → samples channel (100,000 buffer)
    → questdb.Writer (ILP over TCP, 1000 batch, 1s flush)
    → QuestDB
    → questdb.Reader (REST HTTP)
    → API handlers
    → JSON responses or embedded SPA
```

## Alternatives Considered

### Mock YAML Configuration Files

Pros

* Reuses the real configuration pipeline
* No special code paths

Cons

* Requires maintaining mock YAML files
* Schema changes require updating both real and mock configs
* YAML changes don't take effect until restart
* Real config validation still applies (plant name, location, timezone, etc.)

---

### Feature Flag in Main Binary

Pros

* Single binary for all modes

Cons

* Production code path is polluted with development concerns
* Risk of shipping debug behavior to production
* Conditional logic spreads across the codebase

---

### Separate Dev-Mode Binary (Selected)

Pros

* Zero configuration — hardcoded values are always correct
* No risk of dev mode code leaking into production
* Production binary is completely unchanged
* Easy to add dev-specific features (different mock values, performance logging, etc.)
* Clear naming: `cmd/dev-mode/` is obviously not for production

Cons

* Duplicates the wiring logic (creating stores, handlers, server)
* Must be kept in sync if the component constructors change
* An extra binary to build and maintain

## Rationale

A separate binary is the cleanest separation of concerns. The production code path is never touched by development concerns. The dev-mode binary is self-documenting — anyone reading `cmd/dev-mode/main.go` can see exactly what mock configuration is used and how components are wired.

The trade-off of some wiring duplication is acceptable for the guarantee that production code remains untouched.

## Consequences

### Positive

* `go run cmd/dev-mode/main.go` starts a fully functional platform with simulated data.
* Production binary has no mock logic, no dev flags, no conditional branching.
* Easy to demonstrate the platform without real PLCs.
* New developers can start developing against a live system immediately.

### Negative

* Component wiring is duplicated between `cmd/pharma-platform/` and `cmd/dev-mode/`.
* Dev mode hardcodes QuestDB at `localhost:9009` — non-configurable without editing the source.
* Mock driver produces a synthetic sine wave, not realistic industrial telemetry patterns.

## Future Considerations

If the development requirements grow significantly (different mock data generators, simulated PLC failures, network latency simulation), the mock driver should be moved into a shared `internal/plc/drivers/mock` package to avoid duplication between `cmd/dev-mode/` and `cmd/collector-sim/`.

A future `--config` flag could be added to dev-mode to override defaults, but it would never require the full YAML configuration pipeline.
