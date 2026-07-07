```bash
pharma-platform/
│
cmd/
├── pharma-platform/main.go       # Production binary
├── dev-mode/main.go              # Development all-in-one
├── api/main.go                   # Standalone API server
├── collector-sim/collector-sim.go # Standalone simulator
└── seed/main.go                  # Standalone DB seeder
│
internal/
│
├── store/
│   ├── migrate.go                # PostgreSQL + QuestDB migration runner
│   ├── machine.go                # PostgreSQL PLCStore
│   └── tag.go                    # PostgreSQL TagStore
│
├── collector/
│   ├── collector.go              # Collector (scheduler + workers)
│   ├── scheduler.go              # Tag scheduling loop
│   ├── worker.go                 # PLC read worker
│   ├── errors.go
│   ├── bench_collector_test.go
│   └── diag_test.go
│
├── plc/
│   ├── driver.go                 # Driver interface
│   └── drivers/
│       ├── opcua/                # OPC UA driver
│       ├── mc/                   # MC Protocol driver (future)
│       ├── fins/                 # FINS/TCP driver (future)
│       └── ethernetip/           # EtherNet/IP driver (future)
│
├── questdb/
│   ├── client.go                 # TCP ILP client
│   ├── config.go
│   ├── writer.go                 # Batch writer
│   ├── reader.go                 # REST HTTP reader
│   ├── sql.go                    # SQL executor for DDL
│   ├── encoder.go
│   ├── errors.go
│   ├── bench_encode_test.go
│   └── bench_pipeline_test.go
│
├── postgres/
│   ├── client.go                 # PostgreSQL connection
│   ├── config.go
│   ├── writer.go                 # Aggregated data writer
│   └── errors.go
│
├── config/
│   ├── loader.go                 # Bootstrap YAML loader
│   ├── types.go                  # Config structs
│   ├── validator.go              # Config validation
│   └── errors.go
│
├── models/
│   ├── plc.go                    # PLC struct
│   ├── tag.go                    # Tag struct
│   ├── sample.go                 # Sample struct
│   ├── datatype.go               # DataType enum
│   ├── driver_type.go            # DriverType enum
│   ├── quality.go                # Quality enum
│   └── doc.go                    # Package documentation
│
├── api/
│   ├── server.go                 # HTTP server
│   ├── routes.go                 # Chi router setup
│   ├── api.go
│   ├── handlers/
│   │   ├── telemetry.go          # Telemetry endpoints
│   │   ├── plc.go                # PLC endpoints
│   │   ├── tag.go                # Tag endpoints
│   │   ├── alarms.go             # Alarm endpoints
│   │   ├── collector.go          # Collector control endpoints
│   │   ├── system.go             # System status endpoint
│   │   └── health.go             # Health check endpoint
│   ├── middleware/
│   └── responses/
│
├── aggregator/
│   ├── aggregator.go             # Aggregation service
│   ├── aggregate.go              # Aggregation logic
│   ├── config.go
│   └── errors.go
│
└── common/
    ├── logger/
    ├── retry/
    └── utils/
│
config/
├── bootstrap.yaml                # All configuration
└── api.yaml                      # Legacy (kept for reference)
│
deploy/
├── postgres/init/
│   ├── 001_schema.sql            # machines + tags tables
│   ├── 002_seed_machines.sql     # 11 machines
│   └── 003_seed_tags.sql         # 128 tags
└── questdb/init/
    ├── 001_plc_samples.sql       # plc_samples table
    └── 002_events.sql            # alarms, events, logs
│
runtime/
├── docker/
│   ├── Dockerfile                # Multi-stage Go build
│   └── entrypoint.sh
├── docker-compose.yml            # Postgres + QuestDB + App
└── logs/
│
persistent/                       # (gitignored) Docker bind mounts
├── postgres/
└── questdb/
│
docs/
├── adr/                          # Architecture Decision Records
├── srs/                          # Software Requirements Spec
├── repo_layout.md
└── roadmap.md
```
