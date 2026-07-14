CREATE TABLE IF NOT EXISTS production_runs (
    id SERIAL PRIMARY KEY,
    machine_id INTEGER REFERENCES machines(id) ON DELETE CASCADE,
    batch_id TEXT NOT NULL DEFAULT '',
    product_name TEXT NOT NULL DEFAULT '',
    target_qty INTEGER NOT NULL DEFAULT 0,
    good_qty INTEGER NOT NULL DEFAULT 0,
    bad_qty INTEGER NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    end_time TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS downtime_events (
    id SERIAL PRIMARY KEY,
    machine_id INTEGER REFERENCES machines(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    end_time TIMESTAMPTZ,
    reason TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'unscheduled',
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS oee_targets (
    id SERIAL PRIMARY KEY,
    machine_id INTEGER REFERENCES machines(id) ON DELETE CASCADE UNIQUE,
    availability_target REAL NOT NULL DEFAULT 0.95,
    performance_target REAL NOT NULL DEFAULT 0.95,
    quality_target REAL NOT NULL DEFAULT 0.99,
    ideal_cycle_time_seconds REAL NOT NULL DEFAULT 60.0,
    planned_production_time_seconds INTEGER NOT NULL DEFAULT 28800
);

CREATE INDEX IF NOT EXISTS idx_production_runs_machine ON production_runs(machine_id);
CREATE INDEX IF NOT EXISTS idx_downtime_events_machine ON downtime_events(machine_id);
