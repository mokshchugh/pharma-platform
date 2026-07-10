```bash
pharma-platform/
│
├── project/
│   │
│   ├── cmd/
│   │   ├── pharma-platform/main.go       # Production binary
│   │   ├── dev-mode/main.go              # Development all-in-one
│   │   ├── api/main.go                   # Standalone API server
│   │   ├── collector-sim/main.go         # Standalone simulator
│   │   ├── seed/main.go                  # Standalone DB seeder
│   │   └── migrate/main.go              # Standalone migration runner
│   │
│   ├── internal/
│   │   ├── store/                        # PostgreSQL-backed stores
│   │   ├── collector/                    # Telemetry collector
│   │   ├── plc/                          # PLC driver interface + drivers
│   │   ├── questdb/                      # QuestDB client + ILP writer
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
│   │   ├── postgres/
│   │   │   ├── init/                     # PostgreSQL schema DDL
│   │   │   └── seed/                     # Seed data SQL
│   │   └── questdb/init/                 # QuestDB table DDL + views
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
├── web/                                  # React SPA frontend
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
