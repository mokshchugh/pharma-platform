# Pharma Platform — Codebase Audit

This document is a complete inventory of the repository: what every file does and how the
two run modes (`simulate` and `real`) actually work end to end. It reflects the state of
`main` after a full cleanup pass (dead code removed, OPC UA driver completed, several real
bugs fixed and verified live). Anything listed here as a gap is a genuine, current gap — the
dead code and stub files that used to clutter this audit have been removed from the tree,
not just from this document.

For narrative/decision history see `docs/adr/` (18 ADRs) and `docs/roadmap.md`. This document
is the "what exists and what does it do" reference; the ADRs are the "why" reference.

---

## 1. How the run modes work

Two entry points, both invocable via `make` or `./run.sh`: `make simulate` and `make real`.
Both share the same config loader, Postgres/QuestDB clients, and migration logic — they
differ only in where telemetry samples come from.

### `make simulate` → `cmd/dev-mode/main.go`

This is the mode that demonstrates the full product without any real PLCs.

1. Loads `config/bootstrap.yaml`, connects to Postgres and QuestDB, runs schema migrations
   **and** seed data (`deploy/postgres/seed/*.sql` — 11 machines, ~100 tags), runs QuestDB
   schema migrations (`deploy/questdb/init/*.sql`).
2. Constructs `MachineStore`, `TagStore`, `ProductionStore`, `AlarmAckStore`, `ControlStore`
   (all thin Postgres wrappers in `internal/store/`).
3. Calls `productionStore.CloseStaleRunsAndDowntime()` — force-closes any `production_runs`/
   `downtime_events` left open by a previous crashed/restarted process, so stale rows can't
   get double-counted by `CalculateOEE`.
4. Creates two channels: `rawSamples` → **`telemetry.Tracker`** → `writerSamples` → QuestDB
   writer (writes to the `plc_samples` table).
5. `internal/simulation.Simulator` is the sample generator. A 100ms ticker calls `sim.Tick()`,
   which advances every machine's physical state (fault/recovery scheduling, per-machine
   production counters) once, then emits one sample per tag per machine (~100 tags × 10
   ticks/sec ≈ 1000 samples/sec) into `rawSamples`.
6. `internal/telemetry.Tracker.Observe()` runs on every sample as it passes through: it
   watches `Run_Status`/`Alarm_Status`/`Good_Count`/`Reject_Count` tags and derives
   higher-level state — QuestDB `alarms` and `machine_state` rows, and Postgres
   `production_runs`/`downtime_events` rows (with production counts computed as **deltas
   from each run's start**, since the underlying counters are lifetime PLC-style registers
   that don't reset when a machine stops).
7. All HTTP handlers (`internal/api/handlers/*`, wired into `api.NewBackend`) read back from
   Postgres/QuestDB on demand — they never talk to the simulator directly. The simulator and
   the API are fully decoupled through the database, exactly as a real deployment would be.
8. `SIGUSR1`/`SIGUSR2` pause/resume the simulator (`sim.Pause()`/`sim.Resume()` — `Tick()`
   becomes a no-op while paused).
9. The Makefile target also starts `npm run dev` (Vite) in `web/`, and kills the Go process
   when Vite exits.

**Why this is a faithful stand-in for real production data**: the derivation logic in
`internal/telemetry.Tracker` doesn't know or care whether a sample came from the simulator
or a real PLC driver — it only looks at `models.Sample` values on a channel. When a real
driver-polling pipeline is built against the new registry (see §5), it can feed the exact
same `Tracker.Observe()` pipeline and get identical alarm/OEE/production behavior for free.

### `make real` → `cmd/pharma-platform/main.go`

Same boot sequence as `dev-mode`, except:
- Postgres migration runs with **no seed data** — the machines table starts empty.
- No simulator, no `rawSamples`/`telemetry.Tracker` wiring — there is no live sample
  producer in this mode yet, because there's no dashboard-driven machine/driver
  configuration feature (see §5, Phase 8 in the roadmap).
- Uses a no-op `dummyCollector` purely to satisfy the `CollectorStatusProvider`/
  `CollectorHandle` interfaces so the same handlers can be reused.
- Also calls `productionStore.CloseStaleRunsAndDowntime()` on boot, for the same reason.
- Also starts `npm run dev` alongside it, same shutdown pattern.

Today, `make real` boots to a working but empty dashboard — this is intentional groundwork,
not a bug: it's the target for the machine-configuration feature described in the roadmap.

### `./run.sh` — single entrypoint

Installs Go/npm dependencies, checks Docker is installed and running (attempts
`systemctl start docker` if not), starts Postgres + QuestDB (`make up`), then either takes
`simulate`/`real` as an argument or prompts interactively. Postgres is mapped to host port
**5433**, not 5432, specifically to avoid clashing with a locally-installed Postgres — see
`project/runtime/docker-compose.yml` and `project/config/bootstrap.yaml`.

### Other `cmd/` entry points

- **`cmd/migrate/main.go`** — runs schema migrations only (`make migrate`), no seed, no server.
- **`cmd/seed/main.go`** — runs schema + seed migrations only (`make seed`), no server.

---

## 2. Repository layout

```
pharma-platform/
├── Makefile                     # simulate / real / migrate / seed / build / infra
├── run.sh                       # single entrypoint: deps + docker + mode select
├── README.md                    # top-level product README
├── docs/                        # ADRs, roadmap, SRS
├── persistent/                  # bind-mounted Postgres/QuestDB data (gitignored contents)
├── project/                     # Go module (module name: pharma-platform)
│   ├── cmd/                     # 4 binaries, see §1
│   ├── config/                  # bootstrap.yaml, api.yaml
│   ├── deploy/                  # SQL schema/seed files for Postgres + QuestDB
│   ├── internal/                # all application code, see §3
│   └── runtime/                 # docker-compose.yml, Dockerfile, entrypoint.sh
└── web/                         # React + TypeScript SPA (Vite)
    └── src/
        ├── components/          # one file per page + shared UI primitives
        └── lib/
```

Note: there is no `LICENSE` file. The project has not been licensed yet — don't assume MIT
or any other terms.

---

## 3. File-by-file reference

### `project/cmd/` — entry points

| File | Purpose |
|---|---|
| `dev-mode/main.go` | Simulation mode: seeded DB + `internal/simulation.Simulator` + `internal/telemetry.Tracker` + full API + frontend. See §1. |
| `pharma-platform/main.go` | Real mode: empty DB, no sample producer yet, full API + frontend. Built by `runtime/docker/Dockerfile`. See §1. |
| `migrate/main.go` | Runs Postgres + QuestDB schema migrations only. |
| `seed/main.go` | Runs schema migrations + seed data only. |

### `project/internal/simulation/` — the simulation engine

- **`simulator.go`** — `Simulator` (holds `map[int]*MachineSim`, a `samplesChan`, pause/tick/dispatch counters) and `MachineSim` (per-machine `running`/`faulted`/`cycle`/`speed`/`rejectRate`/`goodCount`/`rejectCount`/`downTicksLeft`). `Tick()` advances every known machine's physical state once (`advance()`: fault injection ~every 5 min of running time per machine, 5–60s auto-recovery, production counting scaled by each machine's own `speed`/`rejectRate`), then emits one sample per tag via `simulateValue()` (mostly `base + sin(cycle*freq) + noise` sensor curves, gated on `running`). `allTags()` is a hardcoded, cached catalog of ~100 tags across the 11 seeded machines, mirroring `deploy/postgres/seed/003_seed_tags.sql`.

### `project/internal/telemetry/` — sample-derivation pipeline

- **`tracker.go`** — `Tracker.Observe(ctx, sample)`: the producer-agnostic pipeline described in §1. Derives QuestDB `alarms`/`machine_state` rows and Postgres `production_runs`/`downtime_events` rows from raw `Run_Status`/`Alarm_Status`/`Good_Count`/`Reject_Count` samples. Tracks per-machine baselines (`baseGood`/`baseBad`) so run counts are correct deltas, not raw lifetime-counter values.

### `project/internal/business/` — analytics engine (`/api/v2/analytics/*`)

- **`engine.go`** — `Engine` interface (`GetOverview`, `GetProduction`, `GetQuality`, `GetMachines`, `GetEnergy`, `GetAlarmAnalytics`, `GetCorrelations`, `GetMaintenance`, `GetInsights`); `NewEngine(RealEngineConfig) Engine` constructs a `RealEngine`.
- **`real_engine.go`** — the only implementation. Every method queries Postgres/QuestDB on demand (no caching, no background ticker) — `GetOverview` aggregates per-machine metrics computed from `ProductionStore.CalculateOEE` plus live QuestDB tag reads for temperature/faulted status.
- **`types.go`** — response DTOs (`ExecutiveOverview`, `BusinessMetrics`, `EnergyAnalytics`, etc.). `EnergyAnalytics` and `ExecutiveOverview` each carry a `Simulated bool` field.
- **`energy_fallback.go`** — the one deliberately-synthetic exception: there is no power/current/voltage/vibration sensor data anywhere in the schema, so `GetEnergyAnalytics()` returns a clearly-flagged synthetic profile (`Simulated: true`) instead of pretending to have real data.

### `project/internal/store/` — Postgres-backed domain stores

| File | Purpose |
|---|---|
| `machine.go` | `MachineStore`: PLC/machine listing (`GetPLCs`, `GetAllMachines`, `GetMachine`, `TogglePLCEnabled`), shared `parseDataType`/`machineIDFromString` helpers (also used by `tag.go`). Hardcodes `PollInterval = 100ms` for every tag regardless of DB config. |
| `tag.go` | `TagStore`: tag listing/lookup, reuses `machine.go`'s helpers. |
| `production.go` | `ProductionStore`: production runs, downtime events, OEE targets/calculation (`CalculateOEE`), dashboard summary, production/quality aggregation by hour/day/week/shift/machine/batch, and `CloseStaleRunsAndDowntime()` (startup reconciliation). `ListRuns`/`ListDowntime` branch on `machineID > 0` to support both filtered and unfiltered queries. |
| `controls.go` | `ControlStore`: `machine_control_state` CRUD (`EnsureDefaults`, `List`, `Get`, `Start`, `Stop`, `SetSetpoint`, `SetMode`) backing the Controls page. |
| `alarm_acks.go` | `AlarmAckStore`: acknowledgement state for alarms (QuestDB has no UPDATE, so acks live in Postgres and get joined at query time). `machine_id` is stored as `INTEGER REFERENCES machines(id)`; the store converts to/from the string form used elsewhere in the telemetry pipeline at its API boundary. |
| `migrate.go` | `MigratePostgres`/`MigrateQuestDB`: runs `*.sql` files from a directory in filename order; Postgres seed only runs if the `machines` table is empty. |

### `project/internal/api/` and `internal/api/handlers/`

- **`server.go`** — `NewFull`/`NewBackend` construct `*http.Server` wrappers around the route sets below.
- **`routes.go`** — defines the `Handlers` aggregate struct and a single shared `registerRoutes(r, h)` helper; `Routes(h)` calls it and additionally mounts the embedded SPA at `/*` (full/`real` mode), `RoutesBackend(h)` calls it without the SPA mount (`simulate` mode, frontend served separately by Vite). Full endpoint list: see §4.
- **`handlers/idutil.go`** — `parseTrailingID(s string) int`, the one shared helper for "extract the trailing numeric ID from `machine-5` or a bare `5`" — used by `machine.go`, `plc.go`, and `analytics.go`.
- **`handlers/health.go`** — `Health(w,r)` → `{"status":"ok"}`.
- **`handlers/telemetry.go`** — the largest handler: `Latest`/`LatestByPLC`/`LatestByPLCAndTag`/`History`/`Aggregate`/`DataStream`/`DataStreamCSV`. `resolutionAllowlist` maps `raw`/`1m`/`1h`/`1d`/`1w` to real QuestDB table names — a security-relevant allowlist preventing arbitrary table-name injection.
- **`handlers/machine.go`** — machine list/detail, combines Postgres rows with QuestDB "latest sample" data. `Status`/`CollectionStatus` fields are hardcoded placeholders (`"UNKNOWN"`/`"COLLECTING"`), never computed from real state (see §5).
- **`handlers/plc.go`** — PLC config list/detail/status/tags/toggle. `GetStatus.Connected` just echoes `plc.Enabled`, not a real connectivity check.
- **`handlers/tag.go`** — tag list/lookup, optional `machine_id` filter.
- **`handlers/analytics.go`** — per-machine telemetry + analytics aggregation (`GetTelemetry`, `GetAnalytics`), backs `MachineDetailPage`.
- **`handlers/alarms.go`** — `AlarmStore` (composite key `machineID|tagName|timestampRFC3339Nano` derived live from QuestDB, joined against Postgres `alarm_acks`), `AlarmHandler`. `Acknowledge` unescapes the path segment (`net/url.PathUnescape`) before parsing it, since the ID contains `|`/`:` characters that chi's `URLParam` does not decode — fixed this session, see §5's changelog.
- **`handlers/system.go`** — `/system/status` aggregate (PLC/alarm/collector counts).
- **`handlers/dashboard.go`** — `/api/v1/dashboard` → `ProductionStore.GetDashboardSummary()`.
- **`handlers/oee.go`** — `/api/v1/oee`, `/api/v1/oee/{id}` → `ProductionStore.CalculateOEE`.
- **`handlers/production.go`** — production run + downtime CRUD/list endpoints.
- **`handlers/controls.go`** — start/stop/setpoint/mode endpoints over `ControlStore`.
- **`handlers/collector.go`** — `CollectorHandle` interface (`IsPaused`/`Pause`/`Resume`/`TickCount`/`DispatchSum`), satisfied directly by `internal/simulation.Simulator` (in `simulate` mode) or a no-op `dummyCollector` (in `real` mode).
- **`handlers/business_analytics.go`** — thin wrapper exposing `internal/business.Engine` over HTTP (`/api/v2/analytics/*`).

### `project/internal/models/` — domain types (data only, no logic)

`sample.go` (`Sample`), `tag.go` (`Tag`), `plc.go` (`PLC`), `datatype.go` (`DataType` enum + `String()`), `quality.go` (`Quality` enum, no `String()` — asymmetric with `datatype.go`), `driver_type.go` (`DriverType` enum — only `opcua` has a driver implementation; the rest are recognized but explicitly unimplemented, see §5), `production.go` (OEE/production/control/dashboard DTOs), `doc.go` (package doc only).

### `project/internal/config/`

`types.go` (`Config` + nested `PlantConfig`/`CollectorConfig`/`APIConfig`), `loader.go` (`Load(path)` — every `cmd/` hardcodes the literal path `"config/bootstrap.yaml"`), `validator.go` (`Validate` — defaults Collector/API/Postgres/QuestDB fields, including `QuestDB.BatchSize`/`FlushInterval`).

### `project/internal/postgres/` and `project/internal/questdb/`

- **`postgres/client.go`** — connect-with-retry wrapper over `database/sql` + `lib/pq`.
- **`questdb/client.go`** — raw TCP client to QuestDB's ILP ingestion port, with reconnect-on-failure logic.
- **`questdb/encoder.go`** — `[]models.Sample` → ILP line-protocol text. All numeric types get coerced to floats on the wire (`"%d.0"`), so QuestDB's `value` column is always float64 regardless of the tag's real type.
- **`questdb/writer.go`** — double-buffered batching writer (`accumulate`/`flushLoop` goroutines). On a flush failure it retries every batch in a bounded pending queue (`maxPendingBatches = 5`) on subsequent flush cycles; only drops the oldest batch (with a clear log line) once that backlog is exceeded — no longer a silent single-retry-then-drop (fixed this session).
- **`questdb/sql.go`** — HTTP-based `ExecSQL`/`MigrateDir` (hardcodes QuestDB's HTTP port `9000`, separate from the configurable ILP port). `splitStatements` naively splits SQL on `;` with no awareness of string literals/comments.
- **`questdb/reader.go`** — all the read-side queries: `ListAlarms`, `InsertAlarm`, `MachineStateHistory`, `InsertMachineState`, `SeriesFromView`, `StreamRaw(All)`, `StreamAggregate(All)`.

### `project/internal/plc/` — driver framework

- **`driver.go`** — the `Driver` interface (`Connect`/`Close`/`Read`) every protocol driver must satisfy.
- **`manager.go`** — `Manager`: owns the lifecycle of connected drivers keyed by PLC ID (`Add`/`Get`/`Remove`/`CloseAll`). Only deals in the abstract `Driver` interface, so it doesn't need to import any concrete driver package.
- **`registry/registry.go`** (separate package, to avoid an import cycle with driver implementations that import `internal/plc`) — `New(models.PLC) (plc.Driver, error)`: constructs a real `opcua.Client` for `DriverOPCUA`; every other `DriverType` returns a clear "not implemented yet" error rather than silently doing nothing.
- **`drivers/opcua/`** — the OPC UA driver against `github.com/gopcua/opcua`. `Connect`, `Read`, and `Close` are all implemented (`Read`/`Close` were fixed this session — they used to `panic("not implemented")` despite `reader.go`'s `readTag()` already containing a complete implementation that was simply never called). Unit-tested (`client_test.go`) for the pre-connect error paths; not tested against a live OPC UA server since none is available in this environment.

### `project/internal/web/`

`embed.go` — `go:embed static/*` serves the compiled frontend SPA with fallback-to-`index.html` for client-side routing. Only mounted by `Routes()` (`real` mode), not `RoutesBackend()` (`simulate` mode, where Vite serves the frontend instead).

### `project/deploy/postgres/init/` — Postgres schema

| File | Tables |
|---|---|
| `001_schema.sql` | `machines`, `tags` (FK→machines, cascade delete) |
| `002_oee_schema.sql` | `production_runs`, `downtime_events`, `oee_targets` |
| `003_controls_schema.sql` | `machine_control_state` (one row per machine) |
| `004_alarm_acks_schema.sql` | `alarm_acks` (`machine_id INTEGER REFERENCES machines(id)` — fixed this session, was `TEXT`) |

### `project/deploy/postgres/seed/`

`002_seed_machines.sql` seeds 11 machines (mixed protocols: Mitsubishi FX5U/MC, Omron CJ2M/FINS, B&R X20/OPC UA, Allen Bradley MicroLogix/EtherNet-IP — only guarded to run once, on an empty table). `003_seed_tags.sql` seeds ~100 tags across those machines using PLC-realistic addressing (`D100`, `M100`, etc).

### `project/deploy/questdb/init/` — QuestDB schema

| File | Tables/Views |
|---|---|
| `001_plc_samples.sql` | `plc_samples` (raw telemetry, partitioned by DAY) |
| `002_events.sql` | `alarms`, `events`, `logs` |
| `003_aggregation_views.sql` | `plc_samples_1m/1h/1d/1w` materialized views (avg/min/max/count) |
| `004_production.sql` | `production_counts`, `machine_state` |

### `project/runtime/`

- **`docker-compose.yml`** — `postgres` (17-alpine, **host port 5433** → container 5432, to avoid clashing with a local Postgres install), `questdb` (9.1.0, ports 9000/8812/9009), `app` (built from local `Dockerfile`, port 8081).
- **`docker/Dockerfile`** — multi-stage build; builds `./cmd/pharma-platform` only.
- **`docker/entrypoint.sh`** — trivial `exec ./pharma-platform`; not actually referenced by the Dockerfile's `CMD` (which invokes the binary directly) — effectively dead, kept for anyone building a custom image on top.

### `project/config/`

`bootstrap.yaml` — the real runtime config (Postgres on port 5433, QuestDB, API host/port, collector workers, plant metadata). `api.yaml` — minimal legacy leftover (just host/port), unused by any current entry point.

### `web/src/` — frontend

`App.tsx` (client-side routing via a `useState` switch, no router library — 9 pages), `main.tsx` (Vite/React 19 entrypoint).

| Page | Endpoints it calls |
|---|---|
| `HomePage.tsx` | `GET /api/v1/dashboard`, `GET /system/status`, `GET /api/v2/analytics/overview` |
| `AnalyticsPage.tsx` | all 9 `/api/v2/analytics/*` endpoints (renders via `echarts` + `echarts-for-react/lib/core`) |
| `AlarmsPage.tsx` | `GET /alarms/active`, `GET /alarms`, `POST /alarms/acknowledge/{id}` (now `encodeURIComponent`-escaped — fixed this session alongside the backend decode fix) |
| `ControlsPage.tsx` | `GET /api/v1/controls`, `GET /system/status`, `POST /api/v1/controls/{id}/{action}`, `POST /collector/pause`, `POST /collector/resume` |
| `DataStreamPage.tsx` | `GET /plcs`, `GET /tags`, `GET /telemetry/stream(/csv)` |
| `MachineDetailPage.tsx` | `GET /api/v1/machines/{id}(/telemetry, /analytics)` |
| `MachinesPage.tsx` | `GET /api/v1/machines` |
| `ManagePLCsPage.tsx` | `GET /plcs`, `POST /plcs/{id}/toggle` |
| `ProductionPage.tsx` | `GET /api/v1/production`, `GET /api/v1/downtime` |

Supporting/presentational components (no API calls): `Header.tsx`, `Sidebar.tsx`, `MachineSummaryCard.tsx`, `TelemetryChart.tsx` (a separate, lighter raw-`echarts` chart used elsewhere), `TelemetrySection.tsx`, `HistoricalTable.tsx`, `ResolutionSelector.tsx`, `TimeWindowSelector.tsx`, `components/ui/*` (shadcn-style primitives over Radix: `badge`, `button`, `card`, `input`, `select`, `table`), `lib/utils.ts` (`cn()` helper).

**API surface is inconsistently namespaced**: some routes live under `/api/v1/...` or `/api/v2/...`, others are unversioned root paths (`/plcs`, `/tags`, `/alarms*`, `/telemetry/*`, `/system/status`, `/collector/*`). Not changed this session — a breaking change best done deliberately, not as a side effect of cleanup.

---

## 4. Full endpoint reference

```
GET  /health
GET  /system/status

GET  /telemetry/latest
GET  /telemetry/latest/{plc_id}
GET  /telemetry/latest/{plc_id}/{tag_id}
GET  /telemetry/history
GET  /telemetry/aggregate
GET  /telemetry/stream
GET  /telemetry/stream/csv

GET  /plcs
GET  /plcs/{plc_id}
GET  /plcs/{plc_id}/status
GET  /plcs/{plc_id}/tags
POST /plcs/{plc_id}/toggle

GET  /tags
GET  /tags/{tag_id}

GET  /alarms
GET  /alarms/active
POST /alarms/acknowledge/{id}

GET  /collector/status
POST /collector/pause
POST /collector/resume

GET  /api/v1/machines
GET  /api/v1/machines/{id}
GET  /api/v1/machines/{id}/telemetry
GET  /api/v1/machines/{id}/analytics

GET  /api/v1/dashboard

GET  /api/v1/oee
GET  /api/v1/oee/{id}

GET  /api/v1/production
GET  /api/v1/production/active/{machine_id}
POST /api/v1/production/start
POST /api/v1/production/complete/{id}

GET  /api/v1/downtime
POST /api/v1/downtime/start
POST /api/v1/downtime/end/{id}

GET  /api/v1/controls
GET  /api/v1/controls/{id}
POST /api/v1/controls/{id}/start
POST /api/v1/controls/{id}/stop
POST /api/v1/controls/{id}/setpoint
POST /api/v1/controls/{id}/mode

GET  /api/v2/analytics/overview
GET  /api/v2/analytics/production
GET  /api/v2/analytics/quality
GET  /api/v2/analytics/machines
GET  /api/v2/analytics/energy
GET  /api/v2/analytics/alarms
GET  /api/v2/analytics/correlations
GET  /api/v2/analytics/maintenance
GET  /api/v2/analytics/insights
```

---

## 5. Known gaps and what changed this session

**What's genuinely unbuilt (not bugs — planned future work, see `docs/roadmap.md`):**
- Only `DriverOPCUA` has a working driver; `modbus`/`s7`/`mc`/`fins`/`ethernetip` are recognized `DriverType` values that `internal/plc/registry.New` explicitly rejects with a clear error, by design, until they're built.
- No real polling collector exists yet to drive the OPC UA driver against live PLCs — `internal/plc.Manager` and `internal/plc/registry` are ready for one, but nothing constructs it yet.
- No dashboard-driven machine/driver configuration UI — `make real` boots to an empty but working dashboard.
- `handlers/machine.go`'s `Status`/`CollectionStatus` and `handlers/plc.go`'s `GetStatus.Connected` are still hardcoded placeholders — real values require the collector/driver work above.
- API namespacing (`/api/v1`, `/api/v2`, and unversioned root paths co-existing) is inconsistent but intentionally untouched — a deliberate versioning cleanup, not incidental cleanup.

**Removed this session as dead/orphaned code** (previously present, now gone — mentioned here so the history isn't lost, not because they still exist):
`internal/aggregator` (unwired, its core method was a `// TODO` stub), `internal/postgres/writer.go` (no write methods, only used by the dead aggregator), `cmd/api` (unwired duplicate of `cmd/pharma-platform`), `cmd/collector-sim` and `internal/collector` (dev/testing-only pipeline, superseded by `internal/simulation` + `internal/telemetry.Tracker`), `internal/api/api.go`, `internal/config/errors.go`, `internal/service/telemetry.go` (empty placeholders), `handlers.PLCConfigStore`, `handlers.collectorStatusShim`, `routesTelemetryOnly`/`api.NewTelemetryOnly` (all unused), the duplicated `Routes()`/`RoutesBackend()` route lists (consolidated into one `registerRoutes` helper), three duplicate numeric-ID-parsing helpers (consolidated into `handlers/idutil.go`), `LICENSE` (was a 0-byte file), `docs/repo_layout.md` and `web/README.md` (redundant/boilerplate).

**Real bugs fixed this session (verified live, not just in theory):**
- OPC UA driver's `Read()`/`Close()` both used to `panic("not implemented")`; wired to the existing `readTag()` implementation and a proper `client.Close(ctx)` call.
- The simulator had no working fault/downtime model at all (`faulted`/`running` were never actually toggled), production counters were a single shared counter identical across every machine, and Availability was pinned at 100% forever — rewritten with real per-machine fault/recovery scheduling and independent per-machine counters.
- Alarm acknowledgement was completely broken end-to-end: `chi.URLParam` returns the raw percent-encoded path segment without decoding it, so any alarm ID (which always contains `|` and `:`) failed to parse; fixed by unescaping server-side (`url.PathUnescape`) and encoding client-side (`encodeURIComponent`) — confirmed with a live acknowledge round-trip against Postgres.
- `questdb/writer.go` silently dropped a batch of samples on the second consecutive flush failure; now retries via a bounded pending queue and only drops (with a clear log line) once that backlog is exceeded.
- `internal/config/validator.go` didn't default `QuestDB.BatchSize`/`FlushInterval`, which could panic `time.NewTicker` if `FlushInterval: 0` ever reached it.
- `alarm_acks.machine_id` was `TEXT` while every other machine-scoped Postgres table uses an integer FK to `machines(id)` — schema and store code both fixed, verified with a live insert.
- `production.go`'s `ListRuns`/`ListDowntime` always filtered `WHERE machine_id = $1` even when no filter was requested (defaulting to `0`, which matches nothing) — fixed to branch on whether a filter was actually given.
- `go.mod` marked `github.com/go-chi/chi/v5` as `// indirect` despite being a direct dependency throughout `internal/api` — fixed via `go mod tidy`.
- Two Postgres-connectivity bugs from earlier in this project's life: a redundant fake collector running in parallel with the real simulator (corrupting the sample stream), and `s.Value == 1`-style comparisons against an untyped constant that could never match an `any`-typed `float64` value.
- Two stale/broken QuestDB benchmark test files (`bench_encode_test.go`, `bench_pipeline_test.go`) referenced `models.Sample` fields (`PLCID`/`TagID`) removed by the ADR-0018 identity refactor, predating this session; updated to the current field names so `go vet ./...`/`go test ./...` are fully clean.
- ADR-0003's file header was internally mislabeled "ADR-0002" — fixed.

**Verified working end-to-end**: `go build ./...`, `go vet ./...`, and `go test ./...` are all
clean with no errors. `make simulate` was run against a fresh Postgres/QuestDB (port 5433)
with the new `alarm_acks` schema, confirming: differentiated per-machine OEE/production,
real downtime events, and a full alarm-acknowledge round trip.
