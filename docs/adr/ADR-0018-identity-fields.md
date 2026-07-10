# ADR-0018: Replace plc_id/tag_id with machine_id/machine_name/tag_name

**Status:** Accepted

**Date:** 2026-07-10

## Context

The original QuestDB schema used `plc_id` and `tag_id` as identity columns for telemetry samples. These were opaque string identifiers (`"machine-1"`, `"tag-42"`) that required a JOIN to the PostgreSQL `machines` and `tags` tables to resolve human-readable names. Every telemetry query that needed to display a machine or tag name had to either join across databases or perform a separate lookup.

Additionally, the dual `plc_id`/`machine_id` terminology was inconsistent — the store layer used `machine_id` internally while the QuestDB schema, API routes, and frontend used `plc_id`. This caused confusion throughout the codebase and made documentation inconsistent.

## Decision

Replace the three identity columns in the QuestDB `plc_samples` table and its materialized views:

```
Old: plc_id SYMBOL, tag_id SYMBOL
New: machine_id SYMBOL, machine_name SYMBOL, tag_name SYMBOL
```

And update every layer of the pipeline accordingly.

### What Changed

| Layer | Old | New |
|-------|-----|-----|
| QuestDB schema | `plc_id`, `tag_id` | `machine_id`, `machine_name`, `tag_name` |
| ILP encoder | `plc_id=%s,tag_id=%s` | `machine_id=%s,machine_name=%s,tag_name=%s` |
| Go `Sample` model | `PLCID`, `TagID` | `MachineID`, `MachineName`, `TagName` |
| Go `Tag` model | `PLCID` string only | `PLCID` (kept), `MachineID int`, `MachineName string` |
| Reader SQL queries | `WHERE plc_id = ... AND tag_id = ...` | `WHERE machine_id = ... AND tag_name = ...` |
| API route params | `plc_id`, `tag_id` | Still `plc_id`, `tag_id` for URL backward compat, mapped to `machineID`/`tagName` internally |
| API response fields | `plcid`/`PLCID`, `tagid`/`TagID` | `MachineID`, `MachineName`, `TagName` |
| Frontend (HTML) | `s.plcid \|\| s.PLCID`, `s.tagid \|\| s.TagID` | `s.MachineName \|\| s.MachineID`, `s.TagName` |
| Frontend (React) | `plc_id`, `tag_id` in interfaces | `machine_id`, `machine_name`, `tag_name` |
| Materialized views | `GROUP BY plc_id, tag_id` | `GROUP BY machine_id, machine_name, tag_name` |
| Seed/migration | Mixed in one directory | Separated into `deploy/postgres/init/` (schema) and `deploy/postgres/seed/` (data) |

### What Was Kept

- **API URL parameters** (`plc_id`, `tag_id`) — preserved for backward compatibility; handlers internally map them to `machineID` and `tagName`
- **`Tag.PLCID` field** — kept because it is used by the collector for scheduling key generation (`tagKey = PLCID + ":" + ID`)

## Alternatives Considered

### Keep plc_id/tag_id, add machine_name as a separate column

Pros

* Minimal migration scope
* Backward compatible API responses

Cons

* Still have two parallel identity systems (plc_id and machine_id in the Go model)
* API responses would need to return both old and new field names
* No reduction in confusion

### Full rename across all layers including API routes

Pros

* Consistent everywhere

Cons

* Breaks all existing API clients
* URL path changes are disruptive during development
* Deferred to a future API version bump

## Rationale

The most impactful change was embedding `machine_name` and `tag_name` directly in the QuestDB row. This eliminates the cross-database lookup for every telemetry display query — the dashboard can show "Fluid Bed Dryer / Inlet_Air_Temp" directly from the time-series row without touching PostgreSQL.

Adding `machine_name` to each row is a storage trade-off (a SYMBOL column, cheaply dictionary-encoded in QuestDB) that saves a sequential scan and JOIN on every dashboard load.

## Consequences

### Positive

- Dashboard displays machine names and tag names without cross-database lookups.
- Single consistent `machine_id`/`tag_name` identity across models, DB, and API internals.
- Materialized views carry human-readable names natively.
- `store/tag.go` and `store/machine.go` load `machine_name` via simple JOIN.
- No duplicate identity fields in the Sample model.

### Negative

- **Existing QuestDB telemetry must be dropped** — the `plc_samples` table schema changed incompatibly. Migration files `001_plc_samples.sql` and `003_aggregation_views.sql` now use `DROP TABLE/VIEW IF EXISTS` followed by `CREATE ... IF NOT EXISTS`. This is a one-time manual step (`DROP TABLE plc_samples` via QuestDB console).
- API route params still use `plc_id`/`tag_id` — cosmetic inconsistency remains until a future API version.
- The `Tag` model retains `PLCID` alongside `MachineID` — a future cleanup could remove it.

## Migration

Existing QuestDB volumes must be migrated manually:

```sql
-- Drop old tables/views (data loss)
DROP TABLE plc_samples;
DROP VIEW plc_samples_1m;
DROP VIEW plc_samples_1h;
DROP VIEW plc_samples_1d;
DROP VIEW plc_samples_1w;

-- Then restart or re-run migration
```

The `make migrate` command creates the new schema in both QuestDB and PostgreSQL.
The `make seed` command populates PostgreSQL with sample machines and tags.
