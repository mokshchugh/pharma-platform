# ADR-0007: Protocol-Agnostic PLC Driver Interface

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform collects telemetry data from industrial PLCs in pharmaceutical manufacturing equipment. Different PLC models and vendors use incompatible communication protocols, including OPC UA, Modbus TCP, Mitsubishi MC, Siemens S7, Omron FINS, and EtherNet/IP.

The Collector and higher-level services need to read tags from any PLC without being coupled to the specific wire protocol in use. Adding support for a new PLC protocol should not require changes to the Collector, Aggregator, API, or data pipeline.

A clean abstraction boundary between protocol-specific communication and application-level data processing was required before implementing the first driver.

## Decision

Define a minimal `Driver` interface in a dedicated `internal/plc` package that all protocol implementations must satisfy:

```go
type Driver interface {
    Connect(ctx context.Context) error
    Close() error
    Read(ctx context.Context, tag models.Tag) (models.Sample, error)
}
```

Each industrial protocol gets its own package under `internal/plc/drivers/`:

```
internal/plc/drivers/
    opcua/       # Implemented
    modbus/      # Reserved
    mc/          # Reserved
    s7/          # Reserved
    fins/        # Reserved
    ethernetip/  # Reserved
```

The Collector receives a `plc.Driver` through dependency injection and never inspects the concrete protocol type.

## Alternatives Considered

### Unified Driver with Protocol Enum

Pros

* Single package
* Centralized protocol selection logic

Cons

* Tight coupling between protocols
* Changes to one protocol risk affecting others
* Difficult to test or compile out unused protocols

---

### Driver Interface with Registry (Selected)

Pros

* Clean separation of concerns
* Easy to test with mock drivers
* New protocols added without modifying existing code
* Each driver package can have its own configuration, errors, and internal types
* Supports library-level integration testing per protocol

Cons

* More packages to maintain
* Driver discovery requires either registration or explicit wiring

## Rationale

A narrow interface with three methods provides a stable contract that is simple to implement for any industrial protocol. Connect/Close/Read maps naturally onto the lifecycle of a PLC connection and the unit of work—reading a single tag value.

Separating each protocol into its own package keeps dependencies isolated (for example, the OPC UA package imports `gopcua`, but Modbus will import a Modbus library without affecting any other package). This also allows compiling protocol support selectively.

## Consequences

### Positive

* The Collector, Aggregator, and API are fully decoupled from PLC protocols.
* Mock drivers can be injected for unit tests and simulations.
* Adding a new protocol is a single-package addition.
* Each driver can manage its own connection lifecycle, retries, and error semantics.

### Negative

* Wires protocol selection into application-level configuration rather than auto-detection.
* Each protocol driver must independently implement the same lifecycle patterns.

## Future Considerations

If the number of supported protocols grows significantly, a driver registry pattern (`internal/plc/registry.go`) can be introduced to allow plugins or config-driven driver selection without modifying `main.go`.

The interface may be extended with `ReadBatch(ctx, []Tag)` if bulk-read performance becomes a requirement for high-density PLCs.
