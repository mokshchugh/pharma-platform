CREATE TABLE IF NOT EXISTS machine_control_state (
    machine_id INTEGER PRIMARY KEY REFERENCES machines(id) ON DELETE CASCADE,
    running BOOLEAN NOT NULL DEFAULT false,
    speed REAL NOT NULL DEFAULT 0,
    setpoint REAL NOT NULL DEFAULT 100,
    mode TEXT NOT NULL DEFAULT 'auto',
    temperature REAL NOT NULL DEFAULT 25,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
