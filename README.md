# ARCHITECTURE

```bash
pharma-platform/
├── cmd
│   ├── aggregator
│   ├── api
│   ├── collector
│   └── simulator
├── config
│   ├── aggregation.yaml
│   ├── api.yaml
│   ├── collector.yaml
│   ├── plant.yaml
│   ├── plcs.yaml
│   └── tags.yaml
├── deploy
│   ├── compose.yaml
│   ├── grafana
│   ├── postgres
│   │   └── init
│   ├── questdb
│   └── scripts
├── docs
│   ├── adr
│   │   ├── ADR-0001-use-quest.md
│   │   ├── ADR-0002-use-go.md
│   │   ├── ADR-0003-use-postgre.md
│   │   ├── ADR-0004-persistent-data.md
│   │   └── ADR-0005-use-docker.md
│   ├── repo_layout.md
│   └── srs
│       └── architecture_V1.md
├── go.mod
├── internal
│   ├── aggregator
│   ├── api
│   ├── collector
│   ├── common
│   │   ├── logger
│   │   ├── retry
│   │   └── utils
│   ├── config
│   ├── models
│   ├── plc
│   │   ├── drivers
│   │   ├── manager
│   │   └── protocols
│   ├── postgres
│   └── questdb
├── LICENSE
└── README.md

32 directories, 17 files
```