# ADR-0006: Adopt an OPC UA-First Implementation Strategy

## Status

Accepted

---

## Context

The telemetry platform is intended to support communication with multiple PLC protocols, including:

- OPC UA
- Modbus TCP
- Mitsubishi MC Protocol
- Siemens S7
- Omron FINS
- EtherNet/IP

The original architecture was designed to remain protocol-agnostic so that additional PLC drivers can be introduced without modifying higher-level services such as the Collector, Aggregator, or API.

However, implementing multiple industrial communication protocols simultaneously would significantly increase the project's complexity, development time, and testing effort.

The initial deployment targets PLCs that expose OPC UA servers either natively or through OPC UA gateways.

---

## Decision

The platform will implement only the OPC UA protocol for the initial release.

The overall architecture will continue to reserve dedicated packages and interfaces for future protocol implementations, ensuring that support for additional PLC protocols can be introduced incrementally without requiring major architectural changes.

The PLC package will expose a protocol-independent Driver interface, while the first concrete implementation will be provided by the OPC UA driver.

Future protocol implementations will reside in:

- internal/plc/drivers/modbus
- internal/plc/drivers/mc
- internal/plc/drivers/s7
- internal/plc/drivers/fins
- internal/plc/drivers/ethernetip

These packages are intentionally reserved but will remain unimplemented until a concrete business requirement exists.

---

## Consequences

### Positive

- Reduces implementation complexity for the initial release.
- Enables early development of the Collector, database pipeline, and API.
- Allows testing against a single industrial communication protocol.
- Preserves a protocol-independent architecture for future expansion.
- Avoids premature implementation of protocols that are not currently required.

### Negative

- PLCs without OPC UA support will require an OPC UA gateway or future protocol implementation.
- Native support for proprietary industrial protocols is deferred to future releases.

### Future Considerations

As new deployment requirements emerge, additional protocol drivers can be implemented without modifying the Collector or other application services.

Each future driver will implement the common PLC Driver interface, allowing protocol-specific communication logic to remain isolated from the rest of the application.