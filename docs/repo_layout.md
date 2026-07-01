```bash
pharma-platform/

cmd/
│
├── collector/
│   └── main.go
│
├── api/
│   └── main.go
│
├── aggregator/
│   └── main.go
│
└── simulator/
    └── main.go

internal/
│
├── collector/
│
│   ├── app.go
│   ├── scheduler.go
│   ├── buffer.go
│   ├── writer.go
│   ├── metrics.go
│   └── health.go
│
├── plc/
│
│   ├── manager.go
│   ├── driver.go
│   │
│   ├── drivers/
│   │
│   │   ├── mitsubishi/
│   │   ├── omron/
│   │   ├── allenbradley/
│   │   ├── schneider/
│   │   ├── pilz/
│   │   └── br/
│   │
│   └── protocols/
│
│       ├── mcprotocol/
│       ├── ethernetip/
│       ├── fins/
│       ├── modbus/
│       └── opcua/
│
├── questdb/
│
│   ├── client.go
│   ├── writer.go
│   └── schema.go
│
├── postgres/
│
│   ├── client.go
│   └── migrations.go
│
├── config/
│
│   ├── loader.go
│   ├── validator.go
│   └── models.go
│
├── models/
│
│   ├── plc.go
│   ├── tag.go
│   ├── event.go
│   └── sample.go
│
└── common/
    ├── logger/
    ├── retry/
    ├── utils/
    └── errors/

config/

├── plant.yaml
├── plcs.yaml
├── tags.yaml
├── collector.yaml
├── api.yaml
└── aggregation.yaml

deploy/

├── compose.yaml
└── .env
```