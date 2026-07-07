CREATE TABLE IF NOT EXISTS machines (
    id              SERIAL PRIMARY KEY,
    machine_name    TEXT NOT NULL,
    brand           TEXT NOT NULL,
    model           TEXT NOT NULL,
    protocol        TEXT NOT NULL,
    connection_type TEXT NOT NULL DEFAULT 'ethernet',
    ip_address      TEXT,
    port            INTEGER,
    notes           TEXT,
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tags (
    id              SERIAL PRIMARY KEY,
    machine_id      INTEGER REFERENCES machines(id) ON DELETE CASCADE,
    tag_name        TEXT NOT NULL,
    description     TEXT,
    data_type       TEXT NOT NULL DEFAULT 'float64',
    scale_factor    REAL DEFAULT 1.0,
    unit            TEXT,
    address         TEXT NOT NULL,
    enabled         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tags_machine_id ON tags(machine_id);
CREATE INDEX IF NOT EXISTS idx_tags_enabled ON tags(enabled);
