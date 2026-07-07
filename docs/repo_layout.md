```bash
pharma-platform/
│
├── project/
│   │
│   ├── cmd/
│   │   ├── pharma-platform/main.go       # Production binary
│   │   ├── dev-mode/main.go              # Development all-in-one
│   │   ├── api/main.go                   # Standalone API server
│   │   ├── collector-sim/collector-sim.go # Standalone simulator
│   │   └── seed/main.go                  # Standalone DB seeder
│   │
│   ├── internal/
│   │   ├── store/                        # PostgreSQL-backed stores
│   │   ├── collector/                    # Telemetry collector
│   │   ├── plc/                          # PLC driver interface + drivers
│   │   ├── questdb/                      # QuestDB client
│   │   ├── postgres/                     # PostgreSQL client
│   │   ├── config/                       # Bootstrap config loader
│   │   ├── api/                          # REST API handlers + server
│   │   ├── models/                       # Domain models
│   │   └── aggregator/                   # Aggregation service
│   │
│   ├── config/
│   │   └── bootstrap.yaml                # Single config file
│   │
│   ├── deploy/
│   │   ├── postgres/init/                # PostgreSQL schema + seed SQL
│   │   └── questdb/init/                 # QuestDB table DDL
│   │
│   ├── runtime/
│   │   ├── docker/
│   │   │   ├── Dockerfile
│   │   │   └── entrypoint.sh
│   │   └── docker-compose.yml
│   │
│   ├── go.mod
│   └── go.sum
│
├── persistent/                           # (git-tracked skeleton) Docker bind mounts
│   ├── postgres/
│   └── questdb/
│
├── docs/
│   ├── adr/
│   ├── srs/
│   ├── repo_layout.md
│   └── roadmap.md
│
├── Makefile                              # Wraps all common commands
├── .gitignore
├── LICENSE
└── README.md
```
